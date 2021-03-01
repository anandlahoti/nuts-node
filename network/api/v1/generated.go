// Package v1 provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/labstack/echo/v4"
)

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A callback for modifying requests which are generated before sending over
	// the network.
	RequestEditor RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = http.DefaultClient
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditor = fn
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// ListDocuments request
	ListDocuments(ctx context.Context) (*http.Response, error)

	// GetDocument request
	GetDocument(ctx context.Context, ref string) (*http.Response, error)

	// GetDocumentPayload request
	GetDocumentPayload(ctx context.Context, ref string) (*http.Response, error)
}

func (c *Client) ListDocuments(ctx context.Context) (*http.Response, error) {
	req, err := NewListDocumentsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) GetDocument(ctx context.Context, ref string) (*http.Response, error) {
	req, err := NewGetDocumentRequest(c.Server, ref)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) GetDocumentPayload(ctx context.Context, ref string) (*http.Response, error) {
	req, err := NewGetDocumentPayloadRequest(c.Server, ref)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

// NewListDocumentsRequest generates requests for ListDocuments
func NewListDocumentsRequest(server string) (*http.Request, error) {
	var err error

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/api/document")
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDocumentRequest generates requests for GetDocument
func NewGetDocumentRequest(server string, ref string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "ref", ref)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/api/document/%s", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDocumentPayloadRequest generates requests for GetDocumentPayload
func NewGetDocumentPayloadRequest(server string, ref string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "ref", ref)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/api/document/%s/payload", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// ListDocuments request
	ListDocumentsWithResponse(ctx context.Context) (*ListDocumentsResponse, error)

	// GetDocument request
	GetDocumentWithResponse(ctx context.Context, ref string) (*GetDocumentResponse, error)

	// GetDocumentPayload request
	GetDocumentPayloadWithResponse(ctx context.Context, ref string) (*GetDocumentPayloadResponse, error)
}

type ListDocumentsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]string
}

// Status returns HTTPResponse.Status
func (r ListDocumentsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListDocumentsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDocumentResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r GetDocumentResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDocumentResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDocumentPayloadResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r GetDocumentPayloadResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDocumentPayloadResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// ListDocumentsWithResponse request returning *ListDocumentsResponse
func (c *ClientWithResponses) ListDocumentsWithResponse(ctx context.Context) (*ListDocumentsResponse, error) {
	rsp, err := c.ListDocuments(ctx)
	if err != nil {
		return nil, err
	}
	return ParseListDocumentsResponse(rsp)
}

// GetDocumentWithResponse request returning *GetDocumentResponse
func (c *ClientWithResponses) GetDocumentWithResponse(ctx context.Context, ref string) (*GetDocumentResponse, error) {
	rsp, err := c.GetDocument(ctx, ref)
	if err != nil {
		return nil, err
	}
	return ParseGetDocumentResponse(rsp)
}

// GetDocumentPayloadWithResponse request returning *GetDocumentPayloadResponse
func (c *ClientWithResponses) GetDocumentPayloadWithResponse(ctx context.Context, ref string) (*GetDocumentPayloadResponse, error) {
	rsp, err := c.GetDocumentPayload(ctx, ref)
	if err != nil {
		return nil, err
	}
	return ParseGetDocumentPayloadResponse(rsp)
}

// ParseListDocumentsResponse parses an HTTP response from a ListDocumentsWithResponse call
func ParseListDocumentsResponse(rsp *http.Response) (*ListDocumentsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ListDocumentsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []string
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDocumentResponse parses an HTTP response from a GetDocumentWithResponse call
func ParseGetDocumentResponse(rsp *http.Response) (*GetDocumentResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &GetDocumentResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	}

	return response, nil
}

// ParseGetDocumentPayloadResponse parses an HTTP response from a GetDocumentPayloadWithResponse call
func ParseGetDocumentPayloadResponse(rsp *http.Response) (*GetDocumentPayloadResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &GetDocumentPayloadResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	}

	return response, nil
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Lists the documents on the DAG
	// (GET /api/document)
	ListDocuments(ctx echo.Context) error
	// Retrieves a document
	// (GET /api/document/{ref})
	GetDocument(ctx echo.Context, ref string) error
	// Gets the document payload
	// (GET /api/document/{ref}/payload)
	GetDocumentPayload(ctx echo.Context, ref string) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// ListDocuments converts echo context to params.
func (w *ServerInterfaceWrapper) ListDocuments(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ListDocuments(ctx)
	return err
}

// GetDocument converts echo context to params.
func (w *ServerInterfaceWrapper) GetDocument(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "ref" -------------
	var ref string

	err = runtime.BindStyledParameter("simple", false, "ref", ctx.Param("ref"), &ref)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter ref: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetDocument(ctx, ref)
	return err
}

// GetDocumentPayload converts echo context to params.
func (w *ServerInterfaceWrapper) GetDocumentPayload(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "ref" -------------
	var ref string

	err = runtime.BindStyledParameter("simple", false, "ref", ctx.Param("ref"), &ref)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter ref: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetDocumentPayload(ctx, ref)
	return err
}

// PATCH: This template file was taken from pkg/codegen/templates/register.tmpl

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	Add(method string, path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.Add(http.MethodGet, baseURL+"/api/document", wrapper.ListDocuments)
	router.Add(http.MethodGet, baseURL+"/api/document/:ref", wrapper.GetDocument)
	router.Add(http.MethodGet, baseURL+"/api/document/:ref/payload", wrapper.GetDocumentPayload)

}
