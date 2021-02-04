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

package network

import (
	"github.com/nuts-foundation/nuts-node/crypto/hash"
	"github.com/nuts-foundation/nuts-node/network/dag"
	"time"
)

// Network is the interface to be implemented by any remote or local client
type Network interface {
	// Subscribe makes a subscription for the specified document type. The receiver is called when a document
	// is received for the specified type.
	Subscribe(documentType string, receiver dag.Receiver)
	// GetDocumentPayload retrieves the document payload for the given document. If the document or payload is not found
	// nil is returned.
	GetDocumentPayload(documentRef hash.SHA256Hash) ([]byte, error)
	// GetDocument retrieves the document for the given reference. If the document is not known, an error is returned.
	GetDocument(documentRef hash.SHA256Hash) (dag.Document, error)
	// CreateDocument creates a new document with the specified payload, and signs it using the specified key.
	// If the key should be inside the document (instead of being referred to) `attachKey` should be true.
	CreateDocument(payloadType string, payload []byte, signingKeyID string, attachKey bool, timestamp time.Time, fieldsOpts ...dag.FieldOpt) (dag.Document, error)
	// ListDocuments returns all documents known to this NetworkEngine instance.
	ListDocuments() ([]dag.Document, error)
}