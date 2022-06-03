// Package v1 provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.11.0 DO NOT EDIT.
package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/labstack/echo/v4"
)

// DIDCreateRequest defines model for DIDCreateRequest.
type DIDCreateRequest struct {
	// indicates if the generated key pair can be used for assertions.
	AssertionMethod *bool `json:"assertionMethod,omitempty"`

	// indicates if the generated key pair can be used for authentication.
	Authentication *bool `json:"authentication,omitempty"`

	// indicates if the generated key pair can be used for capability delegations.
	CapabilityDelegation *bool `json:"capabilityDelegation,omitempty"`

	// indicates if the generated key pair can be used for altering DID Documents.
	// In combination with selfControl = true, the key can be used to alter the new DID Document.
	// Defaults to true when not given.
	// default: true
	CapabilityInvocation *bool `json:"capabilityInvocation,omitempty"`

	// List of DIDs that can control the new DID Document. If selfControl = true and controllers is not empty,
	// the newly generated DID will be added to the list of controllers.
	Controllers *[]string `json:"controllers,omitempty"`

	// indicates if the generated key pair can be used for Key agreements.
	KeyAgreement *bool `json:"keyAgreement,omitempty"`

	// whether the generated DID Document can be altered with its own capabilityInvocation key.
	SelfControl *bool `json:"selfControl,omitempty"`
}

// DIDResolutionResult defines model for DIDResolutionResult.
type DIDResolutionResult struct {
	// A DID document according to the W3C spec following the Nuts Method rules as defined in [Nuts RFC006]
	Document DIDDocument `json:"document"`

	// The DID document metadata.
	DocumentMetadata DIDDocumentMetadata `json:"documentMetadata"`
}

// DIDUpdateRequest defines model for DIDUpdateRequest.
type DIDUpdateRequest struct {
	// The hash of the document in hex format.
	CurrentHash string `json:"currentHash"`

	// A DID document according to the W3C spec following the Nuts Method rules as defined in [Nuts RFC006]
	Document DIDDocument `json:"document"`
}

// CreateDIDJSONBody defines parameters for CreateDID.
type CreateDIDJSONBody = DIDCreateRequest

// GetDIDParams defines parameters for GetDID.
type GetDIDParams struct {
	// If a versionId parameter is provided, the DID resolution algorithm returns a specific version of the DID document.
	// The version is the Sha256 hash of the document.
	// The DID parameters versionId and versionTime are mutually exclusive.
	//
	// See [the did resolution spec about versioning](https://w3c-ccg.github.io/did-resolution/#versioning)
	VersionId *string `form:"versionId,omitempty" json:"versionId,omitempty"`

	// If a versionTime parameter is provided, the DID resolution algorithm returns a specific version of the DID document.
	// The DID parameters versionId and versionTime are mutually exclusive.
	//
	// See [the did resolution spec about versioning](https://w3c-ccg.github.io/did-resolution/#versioning)
	VersionTime *string `form:"versionTime,omitempty" json:"versionTime,omitempty"`
}

// UpdateDIDJSONBody defines parameters for UpdateDID.
type UpdateDIDJSONBody = DIDUpdateRequest

// CreateDIDJSONRequestBody defines body for CreateDID for application/json ContentType.
type CreateDIDJSONRequestBody = CreateDIDJSONBody

// UpdateDIDJSONRequestBody defines body for UpdateDID for application/json ContentType.
type UpdateDIDJSONRequestBody = UpdateDIDJSONBody

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

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
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
		client.Client = &http.Client{}
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
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// CreateDID request with any body
	CreateDIDWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateDID(ctx context.Context, body CreateDIDJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ConflictedDIDs request
	ConflictedDIDs(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeactivateDID request
	DeactivateDID(ctx context.Context, did string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDID request
	GetDID(ctx context.Context, did string, params *GetDIDParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdateDID request with any body
	UpdateDIDWithBody(ctx context.Context, did string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdateDID(ctx context.Context, did string, body UpdateDIDJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// AddNewVerificationMethod request
	AddNewVerificationMethod(ctx context.Context, did string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteVerificationMethod request
	DeleteVerificationMethod(ctx context.Context, did string, kid string, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) CreateDIDWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateDIDRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateDID(ctx context.Context, body CreateDIDJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateDIDRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ConflictedDIDs(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewConflictedDIDsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeactivateDID(ctx context.Context, did string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeactivateDIDRequest(c.Server, did)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDID(ctx context.Context, did string, params *GetDIDParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDIDRequest(c.Server, did, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDIDWithBody(ctx context.Context, did string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateDIDRequestWithBody(c.Server, did, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDID(ctx context.Context, did string, body UpdateDIDJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateDIDRequest(c.Server, did, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) AddNewVerificationMethod(ctx context.Context, did string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewAddNewVerificationMethodRequest(c.Server, did)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteVerificationMethod(ctx context.Context, did string, kid string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteVerificationMethodRequest(c.Server, did, kid)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewCreateDIDRequest calls the generic CreateDID builder with application/json body
func NewCreateDIDRequest(server string, body CreateDIDJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateDIDRequestWithBody(server, "application/json", bodyReader)
}

// NewCreateDIDRequestWithBody generates requests for CreateDID with any type of body
func NewCreateDIDRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/vdr/v1/did")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewConflictedDIDsRequest generates requests for ConflictedDIDs
func NewConflictedDIDsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/vdr/v1/did/conflicted")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewDeactivateDIDRequest generates requests for DeactivateDID
func NewDeactivateDIDRequest(server string, did string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "did", runtime.ParamLocationPath, did)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/vdr/v1/did/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDIDRequest generates requests for GetDID
func NewGetDIDRequest(server string, did string, params *GetDIDParams) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "did", runtime.ParamLocationPath, did)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/vdr/v1/did/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	queryValues := queryURL.Query()

	if params.VersionId != nil {

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "versionId", runtime.ParamLocationQuery, *params.VersionId); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	if params.VersionTime != nil {

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "versionTime", runtime.ParamLocationQuery, *params.VersionTime); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	queryURL.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewUpdateDIDRequest calls the generic UpdateDID builder with application/json body
func NewUpdateDIDRequest(server string, did string, body UpdateDIDJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateDIDRequestWithBody(server, did, "application/json", bodyReader)
}

// NewUpdateDIDRequestWithBody generates requests for UpdateDID with any type of body
func NewUpdateDIDRequestWithBody(server string, did string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "did", runtime.ParamLocationPath, did)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/vdr/v1/did/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewAddNewVerificationMethodRequest generates requests for AddNewVerificationMethod
func NewAddNewVerificationMethodRequest(server string, did string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "did", runtime.ParamLocationPath, did)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/vdr/v1/did/%s/verificationmethod", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewDeleteVerificationMethodRequest generates requests for DeleteVerificationMethod
func NewDeleteVerificationMethodRequest(server string, did string, kid string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "did", runtime.ParamLocationPath, did)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "kid", runtime.ParamLocationPath, kid)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/vdr/v1/did/%s/verificationmethod/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
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
	// CreateDID request with any body
	CreateDIDWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateDIDResponse, error)

	CreateDIDWithResponse(ctx context.Context, body CreateDIDJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateDIDResponse, error)

	// ConflictedDIDs request
	ConflictedDIDsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ConflictedDIDsResponse, error)

	// DeactivateDID request
	DeactivateDIDWithResponse(ctx context.Context, did string, reqEditors ...RequestEditorFn) (*DeactivateDIDResponse, error)

	// GetDID request
	GetDIDWithResponse(ctx context.Context, did string, params *GetDIDParams, reqEditors ...RequestEditorFn) (*GetDIDResponse, error)

	// UpdateDID request with any body
	UpdateDIDWithBodyWithResponse(ctx context.Context, did string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateDIDResponse, error)

	UpdateDIDWithResponse(ctx context.Context, did string, body UpdateDIDJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateDIDResponse, error)

	// AddNewVerificationMethod request
	AddNewVerificationMethodWithResponse(ctx context.Context, did string, reqEditors ...RequestEditorFn) (*AddNewVerificationMethodResponse, error)

	// DeleteVerificationMethod request
	DeleteVerificationMethodWithResponse(ctx context.Context, did string, kid string, reqEditors ...RequestEditorFn) (*DeleteVerificationMethodResponse, error)
}

type CreateDIDResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r CreateDIDResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateDIDResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ConflictedDIDsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]DIDResolutionResult
}

// Status returns HTTPResponse.Status
func (r ConflictedDIDsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ConflictedDIDsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeactivateDIDResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r DeactivateDIDResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeactivateDIDResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDIDResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DIDResolutionResult
}

// Status returns HTTPResponse.Status
func (r GetDIDResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDIDResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateDIDResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r UpdateDIDResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateDIDResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type AddNewVerificationMethodResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r AddNewVerificationMethodResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r AddNewVerificationMethodResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteVerificationMethodResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r DeleteVerificationMethodResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteVerificationMethodResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// CreateDIDWithBodyWithResponse request with arbitrary body returning *CreateDIDResponse
func (c *ClientWithResponses) CreateDIDWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateDIDResponse, error) {
	rsp, err := c.CreateDIDWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateDIDResponse(rsp)
}

func (c *ClientWithResponses) CreateDIDWithResponse(ctx context.Context, body CreateDIDJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateDIDResponse, error) {
	rsp, err := c.CreateDID(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateDIDResponse(rsp)
}

// ConflictedDIDsWithResponse request returning *ConflictedDIDsResponse
func (c *ClientWithResponses) ConflictedDIDsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ConflictedDIDsResponse, error) {
	rsp, err := c.ConflictedDIDs(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseConflictedDIDsResponse(rsp)
}

// DeactivateDIDWithResponse request returning *DeactivateDIDResponse
func (c *ClientWithResponses) DeactivateDIDWithResponse(ctx context.Context, did string, reqEditors ...RequestEditorFn) (*DeactivateDIDResponse, error) {
	rsp, err := c.DeactivateDID(ctx, did, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeactivateDIDResponse(rsp)
}

// GetDIDWithResponse request returning *GetDIDResponse
func (c *ClientWithResponses) GetDIDWithResponse(ctx context.Context, did string, params *GetDIDParams, reqEditors ...RequestEditorFn) (*GetDIDResponse, error) {
	rsp, err := c.GetDID(ctx, did, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDIDResponse(rsp)
}

// UpdateDIDWithBodyWithResponse request with arbitrary body returning *UpdateDIDResponse
func (c *ClientWithResponses) UpdateDIDWithBodyWithResponse(ctx context.Context, did string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateDIDResponse, error) {
	rsp, err := c.UpdateDIDWithBody(ctx, did, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDIDResponse(rsp)
}

func (c *ClientWithResponses) UpdateDIDWithResponse(ctx context.Context, did string, body UpdateDIDJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateDIDResponse, error) {
	rsp, err := c.UpdateDID(ctx, did, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDIDResponse(rsp)
}

// AddNewVerificationMethodWithResponse request returning *AddNewVerificationMethodResponse
func (c *ClientWithResponses) AddNewVerificationMethodWithResponse(ctx context.Context, did string, reqEditors ...RequestEditorFn) (*AddNewVerificationMethodResponse, error) {
	rsp, err := c.AddNewVerificationMethod(ctx, did, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseAddNewVerificationMethodResponse(rsp)
}

// DeleteVerificationMethodWithResponse request returning *DeleteVerificationMethodResponse
func (c *ClientWithResponses) DeleteVerificationMethodWithResponse(ctx context.Context, did string, kid string, reqEditors ...RequestEditorFn) (*DeleteVerificationMethodResponse, error) {
	rsp, err := c.DeleteVerificationMethod(ctx, did, kid, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteVerificationMethodResponse(rsp)
}

// ParseCreateDIDResponse parses an HTTP response from a CreateDIDWithResponse call
func ParseCreateDIDResponse(rsp *http.Response) (*CreateDIDResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateDIDResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ParseConflictedDIDsResponse parses an HTTP response from a ConflictedDIDsWithResponse call
func ParseConflictedDIDsResponse(rsp *http.Response) (*ConflictedDIDsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ConflictedDIDsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []DIDResolutionResult
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeactivateDIDResponse parses an HTTP response from a DeactivateDIDWithResponse call
func ParseDeactivateDIDResponse(rsp *http.Response) (*DeactivateDIDResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeactivateDIDResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ParseGetDIDResponse parses an HTTP response from a GetDIDWithResponse call
func ParseGetDIDResponse(rsp *http.Response) (*GetDIDResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDIDResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DIDResolutionResult
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateDIDResponse parses an HTTP response from a UpdateDIDWithResponse call
func ParseUpdateDIDResponse(rsp *http.Response) (*UpdateDIDResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdateDIDResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ParseAddNewVerificationMethodResponse parses an HTTP response from a AddNewVerificationMethodWithResponse call
func ParseAddNewVerificationMethodResponse(rsp *http.Response) (*AddNewVerificationMethodResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &AddNewVerificationMethodResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ParseDeleteVerificationMethodResponse parses an HTTP response from a DeleteVerificationMethodWithResponse call
func ParseDeleteVerificationMethodResponse(rsp *http.Response) (*DeleteVerificationMethodResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteVerificationMethodResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Creates a new Nuts DID
	// (POST /internal/vdr/v1/did)
	CreateDID(ctx echo.Context) error
	// Retrieve the list of conflicted DID documents
	// (GET /internal/vdr/v1/did/conflicted)
	ConflictedDIDs(ctx echo.Context) error
	// Deactivates a Nuts DID document according to the specification.
	// (DELETE /internal/vdr/v1/did/{did})
	DeactivateDID(ctx echo.Context, did string) error
	// Resolves a Nuts DID document
	// (GET /internal/vdr/v1/did/{did})
	GetDID(ctx echo.Context, did string, params GetDIDParams) error
	// Updates a Nuts DID document.
	// (PUT /internal/vdr/v1/did/{did})
	UpdateDID(ctx echo.Context, did string) error
	// Creates and adds a new verificationMethod to the DID document.
	// (POST /internal/vdr/v1/did/{did}/verificationmethod)
	AddNewVerificationMethod(ctx echo.Context, did string) error
	// Delete a specific verification method
	// (DELETE /internal/vdr/v1/did/{did}/verificationmethod/{kid})
	DeleteVerificationMethod(ctx echo.Context, did string, kid string) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// CreateDID converts echo context to params.
func (w *ServerInterfaceWrapper) CreateDID(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateDID(ctx)
	return err
}

// ConflictedDIDs converts echo context to params.
func (w *ServerInterfaceWrapper) ConflictedDIDs(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ConflictedDIDs(ctx)
	return err
}

// DeactivateDID converts echo context to params.
func (w *ServerInterfaceWrapper) DeactivateDID(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "did" -------------
	var did string

	err = runtime.BindStyledParameterWithLocation("simple", false, "did", runtime.ParamLocationPath, ctx.Param("did"), &did)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter did: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DeactivateDID(ctx, did)
	return err
}

// GetDID converts echo context to params.
func (w *ServerInterfaceWrapper) GetDID(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "did" -------------
	var did string

	err = runtime.BindStyledParameterWithLocation("simple", false, "did", runtime.ParamLocationPath, ctx.Param("did"), &did)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter did: %s", err))
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params GetDIDParams
	// ------------- Optional query parameter "versionId" -------------

	err = runtime.BindQueryParameter("form", true, false, "versionId", ctx.QueryParams(), &params.VersionId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter versionId: %s", err))
	}

	// ------------- Optional query parameter "versionTime" -------------

	err = runtime.BindQueryParameter("form", true, false, "versionTime", ctx.QueryParams(), &params.VersionTime)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter versionTime: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetDID(ctx, did, params)
	return err
}

// UpdateDID converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateDID(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "did" -------------
	var did string

	err = runtime.BindStyledParameterWithLocation("simple", false, "did", runtime.ParamLocationPath, ctx.Param("did"), &did)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter did: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateDID(ctx, did)
	return err
}

// AddNewVerificationMethod converts echo context to params.
func (w *ServerInterfaceWrapper) AddNewVerificationMethod(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "did" -------------
	var did string

	err = runtime.BindStyledParameterWithLocation("simple", false, "did", runtime.ParamLocationPath, ctx.Param("did"), &did)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter did: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.AddNewVerificationMethod(ctx, did)
	return err
}

// DeleteVerificationMethod converts echo context to params.
func (w *ServerInterfaceWrapper) DeleteVerificationMethod(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "did" -------------
	var did string

	err = runtime.BindStyledParameterWithLocation("simple", false, "did", runtime.ParamLocationPath, ctx.Param("did"), &did)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter did: %s", err))
	}

	// ------------- Path parameter "kid" -------------
	var kid string

	err = runtime.BindStyledParameterWithLocation("simple", false, "kid", runtime.ParamLocationPath, ctx.Param("kid"), &kid)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter kid: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DeleteVerificationMethod(ctx, did, kid)
	return err
}

// PATCH: This template file was taken from pkg/codegen/templates/echo/echo-register.tmpl

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

type Preprocessor interface {
	Preprocess(operationID string, context echo.Context)
}

type ErrorStatusCodeResolver interface {
	ResolveStatusCode(err error) int
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

	// PATCH: This alteration wraps the call to the implementation in a function that sets the "OperationId" context parameter,
	// so it can be used in error reporting middleware.
	router.POST(baseURL+"/internal/vdr/v1/did", func(context echo.Context) error {
		si.(Preprocessor).Preprocess("CreateDID", context)
		return wrapper.CreateDID(context)
	})
	router.GET(baseURL+"/internal/vdr/v1/did/conflicted", func(context echo.Context) error {
		si.(Preprocessor).Preprocess("ConflictedDIDs", context)
		return wrapper.ConflictedDIDs(context)
	})
	router.DELETE(baseURL+"/internal/vdr/v1/did/:did", func(context echo.Context) error {
		si.(Preprocessor).Preprocess("DeactivateDID", context)
		return wrapper.DeactivateDID(context)
	})
	router.GET(baseURL+"/internal/vdr/v1/did/:did", func(context echo.Context) error {
		si.(Preprocessor).Preprocess("GetDID", context)
		return wrapper.GetDID(context)
	})
	router.PUT(baseURL+"/internal/vdr/v1/did/:did", func(context echo.Context) error {
		si.(Preprocessor).Preprocess("UpdateDID", context)
		return wrapper.UpdateDID(context)
	})
	router.POST(baseURL+"/internal/vdr/v1/did/:did/verificationmethod", func(context echo.Context) error {
		si.(Preprocessor).Preprocess("AddNewVerificationMethod", context)
		return wrapper.AddNewVerificationMethod(context)
	})
	router.DELETE(baseURL+"/internal/vdr/v1/did/:did/verificationmethod/:kid", func(context echo.Context) error {
		si.(Preprocessor).Preprocess("DeleteVerificationMethod", context)
		return wrapper.DeleteVerificationMethod(context)
	})

}
