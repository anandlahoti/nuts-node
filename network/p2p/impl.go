/*
 * Copyright (C) 2020. Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */

package p2p

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/nuts-foundation/nuts-node/core"
	"github.com/nuts-foundation/nuts-node/network/log"
	"github.com/nuts-foundation/nuts-node/network/transport"
	errors2 "github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	grpcPeer "google.golang.org/grpc/peer"
	"net"
	"strings"
	"sync"
	"time"
)

type Dialer func(ctx context.Context, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error)

const connectingQueueChannelSize = 100

type p2pNetwork struct {
	config P2PNetworkConfig

	grpcServer *grpc.Server
	listener   net.Listener

	// connectors contains the list of peers we're currently trying to connect to.
	connectors map[string]*connector
	// connectorAddChannel is used to communicate addresses of remote peers to connect to.
	connectorAddChannel chan string
	// peers is the list of peers we're actually connected to. Access MUST be wrapped in locking using peerReadLock and peerWriteLock.
	peers map[PeerID]*peer
	// peersByAddr access MUST be wrapped in locking using peerReadLock and peerWriteLock.
	peersByAddr      map[string]PeerID
	peerMutex        *sync.Mutex
	receivedMessages messageQueue
	peerDialer       Dialer
	configured       bool
}

func (n p2pNetwork) Configured() bool {
	return n.configured
}

func (n p2pNetwork) Diagnostics() []core.DiagnosticResult {
	peers := n.Peers()
	return []core.DiagnosticResult{
		NumberOfPeersStatistic{NumberOfPeers: len(peers)},
		PeersStatistic{Peers: peers},
	}
}

func (n *p2pNetwork) Peers() []PeerID {
	var result []PeerID
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()
	for _, peer := range n.peers {
		result = append(result, peer.id)
	}
	return result
}

func (n *p2pNetwork) Broadcast(message *transport.NetworkMessage) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()
	for _, peer := range n.peers {
		peer.outMessages <- message
	}
}

func (n p2pNetwork) ReceivedMessages() MessageQueue {
	return n.receivedMessages
}

func (n p2pNetwork) Send(peerId PeerID, message *transport.NetworkMessage) error {
	// TODO: Can't we optimize this so that we don't need this lock? Maybe by (secretly) embedding a pointer to the peer in the peer ID?
	var peer *peer
	n.peerMutex.Lock()
	{
		defer n.peerMutex.Unlock()
		peer = n.peers[peerId]
	}
	if peer == nil {
		return fmt.Errorf("unknown peer: %s", peerId)
	}
	peer.outMessages <- message
	return nil
}

type connector struct {
	address string
	backoff Backoff
	Dialer
}

func (c *connector) connect(ownID PeerID, config *tls.Config) (*peer, error) {
	log.Logger().Infof("Connecting to peer: %v", c.address)
	cxt := metadata.NewOutgoingContext(context.Background(), constructMetadata(ownID))
	dialContext, _ := context.WithTimeout(cxt, 5*time.Second)
	conn, err := c.Dialer(dialContext, c.address,
		grpc.WithBlock(), // Dial should block until connection succeeded (or time-out expired)
		grpc.WithTransportCredentials(credentials.NewTLS(config)), // TLS authentication
		grpc.WithReturnConnectionError())                          // This option causes underlying errors to be returned when connections fail, rather than just "context deadline exceeded"
	if err != nil {
		return nil, errors2.Wrap(err, "unable to connect")
	}
	client := transport.NewNetworkClient(conn)
	gate, err := client.Connect(cxt)
	if err != nil {
		log.Logger().Errorf("Failed to set up stream (peer=%s): %v", c.address, err)
		_ = conn.Close()
		return nil, err
	}

	peer := peer{
		conn:       conn,
		client:     client,
		gate:       gate,
		address:    c.address,
		closeMutex: &sync.Mutex{},
	}
	if serverHeader, err := gate.Header(); err != nil {
		log.Logger().Errorf("Error receiving headers from server (peer=%s): %v", c.address, err)
		_ = conn.Close()
		return nil, err
	} else {
		if serverPeerID, err := peerIDFromMetadata(serverHeader); err != nil {
			log.Logger().Errorf("Error parsing PeerID header from server (peer=%s): %v", c.address, err)
			_ = conn.Close()
			return nil, err
		} else if serverPeerID == "" {
			log.Logger().Warnf("Server didn't send a peer ID, dropping connection (peer=%s)", c.address)
			_ = conn.Close()
			return nil, err
		} else {
			peer.id = serverPeerID
		}
	}

	return &peer, nil
}

func NewP2PNetwork() P2PNetwork {
	return &p2pNetwork{
		peers:               make(map[PeerID]*peer, 0),
		peersByAddr:         make(map[string]PeerID, 0),
		connectors:          make(map[string]*connector, 0),
		connectorAddChannel: make(chan string, connectingQueueChannelSize), // TODO: Does this number make sense?
		peerMutex:           &sync.Mutex{},
		receivedMessages:    messageQueue{c: make(chan PeerMessage, 100)}, // TODO: Does this number make sense?
		peerDialer:          grpc.DialContext,
	}
}

func NewP2PNetworkWithOptions(listener net.Listener, dialer Dialer) P2PNetwork {
	result := NewP2PNetwork().(*p2pNetwork)
	result.listener = listener
	result.peerDialer = dialer
	return result
}

type messageQueue struct {
	c chan PeerMessage
}

func (m messageQueue) Get() PeerMessage {
	return <-m.c
}

func (n *p2pNetwork) Configure(config P2PNetworkConfig) error {
	if config.PeerID == "" {
		return errors.New("PeerID is empty")
	}
	if config.TrustStore == nil {
		return errors.New("TrustStore is nil")
	}
	n.config = config
	n.configured = true
	for _, bootstrapNode := range n.config.BootstrapNodes {
		n.ConnectToPeer(bootstrapNode)
	}
	return nil
}

func (n *p2pNetwork) Start() error {
	log.Logger().Infof("Starting gRPC P2P node (ID: %s)", n.config.PeerID)
	if n.config.ListenAddress != "" {
		log.Logger().Infof("Starting gRPC server on %s", n.config.ListenAddress)
		var err error
		// We allow test code to set the listener to allow for in-memory (bufnet) channels
		var serverOpts = make([]grpc.ServerOption, 0)
		if n.listener == nil {
			n.listener, err = net.Listen("tcp", n.config.ListenAddress)
			if err != nil {
				return err
			}
			if n.config.ServerCert.PrivateKey == nil {
				log.Logger().Info("TLS is disabled on gRPC server side! Make sure SSL/TLS offloading is properly configured.")
			} else {
				serverOpts = append(serverOpts, grpc.Creds(credentials.NewTLS(&tls.Config{
					Certificates: []tls.Certificate{n.config.ServerCert},
					ClientAuth:   tls.RequireAndVerifyClientCert,
					ClientCAs:    n.config.TrustStore,
				})))
			}
		}
		n.grpcServer = grpc.NewServer(serverOpts...)
		transport.RegisterNetworkServer(n.grpcServer, n)
		go func() {
			err = n.grpcServer.Serve(n.listener)
			if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				log.Logger().Errorf("gRPC server errored: %v", err)
				_ = n.Stop()
			}
		}()
	}
	// Start client
	go n.connectToNewPeers()
	return nil
}

func (n *p2pNetwork) Stop() error {
	// Stop server
	if n.grpcServer != nil {
		n.grpcServer.Stop()
		n.grpcServer = nil
	}
	if n.listener != nil {
		if err := n.listener.Close(); err != nil {
			log.Logger().Warn("Error while closing server listener: ", err)
		}
		n.listener = nil
	}
	close(n.connectorAddChannel)
	// Stop client
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()
	for _, peer := range n.peers {
		peer.close()
	}
	return nil
}

func (n p2pNetwork) ConnectToPeer(address string) bool {
	if n.shouldConnectTo(address) && len(n.connectorAddChannel) < connectingQueueChannelSize {
		n.connectorAddChannel <- address
		return true
	}
	return false
}

func (n *p2pNetwork) sendAndReceiveForPeer(peer *peer) {
	peer.outMessages = make(chan *transport.NetworkMessage, 10) // TODO: Does this number make sense? Should also be configurable?
	go peer.sendMessages()
	n.addPeer(peer)
	// TODO: Check PeerID sent by peer
	receiveMessages(peer.gate, peer.id, n.receivedMessages)
	peer.close()
	// When we reach this line, receiveMessages has exited which means the connection has been closed.
	n.removePeer(peer)
}

// connectToNewPeers reads from connectorAddChannel to start connecting to new peers
func (n *p2pNetwork) connectToNewPeers() {
	for address := range n.connectorAddChannel {
		if _, present := n.peersByAddr[address]; present {
			log.Logger().Infof("Not connecting to peer, already connected (address=%s)", address)
		} else if n.connectors[address] != nil {
			log.Logger().Infof("Not connecting to peer, already trying to connect (address=%s)", address)
		} else {
			newConnector := &connector{
				address: address,
				backoff: defaultBackoff(),
				Dialer:  n.peerDialer,
			}
			n.connectors[address] = newConnector
			log.Logger().Infof("Added remote peer (address=%s)", address)
			go func() {
				for {
					if n.shouldConnectTo(address) {
						tlsConfig := tls.Config{
							Certificates: []tls.Certificate{n.config.ClientCert},
							RootCAs:      n.config.TrustStore,
						}
						if peer, err := newConnector.connect(n.config.PeerID, &tlsConfig); err != nil {
							waitPeriod := newConnector.backoff.Backoff()
							log.Logger().Warnf("Couldn't connect to peer, reconnecting in %d seconds (peer=%s,err=%v)", int(waitPeriod.Seconds()), newConnector.address, err)
							time.Sleep(waitPeriod)
						} else {
							n.sendAndReceiveForPeer(peer)
							newConnector.backoff.Reset()
							log.Logger().Infof("Connected to peer (address=%s)", newConnector.address)
						}
					}
					time.Sleep(5 * time.Second)
				}
			}()
		}
	}
}

// shouldConnectTo checks whether we should connect to the given node.
func (n p2pNetwork) shouldConnectTo(address string) bool {
	normalizedAddress := normalizeAddress(address)
	if normalizedAddress == normalizeAddress(n.getLocalAddress()) {
		// We're not going to connect to our own node
		log.Logger().Debug("Not connecting since it's localhost")
		return false
	}
	var result = true
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()
	if _, present := n.peersByAddr[normalizedAddress]; present {
		// We're not going to connect to a node we're already connected to
		log.Logger().Tracef("Not connecting since we're already connected (address=%s)", normalizedAddress)
		result = false
	}
	return result
}

func (n p2pNetwork) getLocalAddress() string {
	if n.config.PublicAddress != "" {
		return n.config.PublicAddress
	} else {
		if strings.HasPrefix(n.config.ListenAddress, ":") {
			// Interface's address not included in listening address (e.g. :5555), so prepend with localhost
			return "localhost" + n.config.ListenAddress
		} else {
			// Interface's address included in listening address (e.g. localhost:5555), so return as-is.
			return n.config.ListenAddress
		}
	}
}

func (n p2pNetwork) isRunning() bool {
	return n.grpcServer != nil
}

func (n p2pNetwork) Connect(stream transport.Network_ConnectServer) error {
	peerCtx, _ := grpcPeer.FromContext(stream.Context())
	log.Logger().Tracef("New peer connected from %s", peerCtx.Addr)
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.New("unable to get metadata")
	}
	peerID, err := peerIDFromMetadata(md)
	if err != nil {
		return err
	}
	log.Logger().Infof("New peer connected (add=%s, id=%s)", peerCtx.Addr, peerID)
	// We received our peer's PeerID, now send our own.
	if err := stream.SendHeader(constructMetadata(n.config.PeerID)); err != nil {
		return errors2.Wrap(err, "unable to send headers")
	}
	peer := &peer{
		id:         peerID,
		gate:       stream,
		address:    peerCtx.Addr.String(),
		closeMutex: &sync.Mutex{},
	}
	n.sendAndReceiveForPeer(peer)
	return nil
}

func (n *p2pNetwork) addPeer(peer *peer) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()

	n.peers[peer.id] = peer
	n.peersByAddr[normalizeAddress(peer.address)] = peer.id
}

func (n *p2pNetwork) removePeer(peer *peer) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()

	peer = n.peers[peer.id]
	if peer == nil {
		return
	}

	delete(n.peers, peer.id)
	delete(n.peersByAddr, normalizeAddress(peer.address))
}