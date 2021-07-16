package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// DefaultHTTPTimeout is the default timeout for HTTP requests made by Client
const DefaultHTTPTimeout = time.Minute

// Client is the programmatic interface to the server API.
type Client struct {
	address    string
	headers    map[string]string
	httpClient *http.Client
}

// NewClient creates a client to the server at the given address.
func NewClient(address string) *Client {
	return &Client{
		address:    address,
		headers:    make(map[string]string),
		httpClient: &http.Client{Timeout: DefaultHTTPTimeout},
	}
}

// NewClientWithHeaders creates a client to the server at the given
// address and uses the provided headers.
func NewClientWithHeaders(address string, headers map[string]string) *Client {
	return &Client{
		address:    address,
		headers:    headers,
		httpClient: &http.Client{Timeout: DefaultHTTPTimeout},
	}
}

// closeBody ensures the Body of an http.Response is properly closed.
func closeBody(r *http.Response) {
	if r.Body != nil {
		_, _ = io.ReadAll(r.Body)
		_ = r.Body.Close()
	}
}

func readAPIError(resp *http.Response) error {
	apiErr, err := APIErrorFromReader(resp.Body)
	if err != nil || apiErr == nil {
		return errors.Errorf("failed with status code %d for url %s", resp.StatusCode, resp.Request.URL)
	}

	return errors.Wrapf(errors.New(apiErr.Message), "failed with status code %d", resp.StatusCode)
}

// Headers will return a copy of the HTTP headers that the client sends with all requests.
func (c *Client) Headers() map[string]string {
	returnHeaders := map[string]string{}
	for k, v := range c.headers {
		returnHeaders[k] = v
	}
	return returnHeaders
}

// BuildURL builds the request URL based on the configured address.
func (c *Client) BuildURL(urlPath string, args ...interface{}) string {
	return fmt.Sprintf("%s%s", c.address, fmt.Sprintf(urlPath, args...))
}

func (c *Client) doGet(u string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}
	for k, v := range c.headers {
		req.Header.Add(k, v)
	}

	return c.httpClient.Do(req)
}

func (c *Client) doPost(u string, request interface{}) (*http.Response, error) {
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request")
	}

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(requestBytes))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}
	for k, v := range c.headers {
		req.Header.Add(k, v)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c *Client) doPut(u string, request interface{}) (*http.Response, error) {
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request")
	}

	req, err := http.NewRequest(http.MethodPut, u, bytes.NewReader(requestBytes))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}
	for k, v := range c.headers {
		req.Header.Add(k, v)
	}

	return c.httpClient.Do(req)
}

func (c *Client) doDelete(u string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}
	for k, v := range c.headers {
		req.Header.Add(k, v)
	}

	return c.httpClient.Do(req)
}

// SignUp creates a user.
func (c *Client) SignUp(request *SignUpRequest) (*SignUpResponse, error) {
	resp, err := c.doPost(c.BuildURL("/api/v1/users/signup"), request)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusCreated:
		c.headers[HeaderAuthorization] = fmt.Sprintf("%s %s", AuthorizationBearer, resp.Header.Get(SessionHeader))
		return SignUpResponseFromReader(resp.Body)
	}
	return nil, readAPIError(resp)
}

// ForgotPassword starts the forgot-password flow for a user
func (c *Client) ForgotPassword(request *ForgotPasswordRequest) error {
	resp, err := c.doPost(c.BuildURL("/api/v1/users/forgot-password"), request)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	}
	return readAPIError(resp)
}

// ResetPassword finishes the forgot-password flow for a user
func (c *Client) ResetPassword(request *ResetPasswordRequest) error {
	resp, err := c.doPost(c.BuildURL("/api/v1/users/reset-password-complete"), request)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	}
	return readAPIError(resp)
}

// UpdatePassword updates a user's password.
func (c *Client) UpdatePassword(request *UpdatePasswordRequest) error {
	resp, err := c.doPut(c.BuildURL("/api/v1/users/me/password"), request)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		c.headers[HeaderAuthorization] = fmt.Sprintf("%s %s", AuthorizationBearer, resp.Header.Get(SessionHeader))
		return nil
	}
	return readAPIError(resp)
}

// VerifyEmailStart starts the verify email flow for the user
func (c *Client) VerifyEmailStart() error {
	resp, err := c.doPost(c.BuildURL("/api/v1/users/verify-email"), nil)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	}
	return readAPIError(resp)
}

// VerifyEmailComplete finishes the verify email flow for the user
func (c *Client) VerifyEmailComplete(request *VerifyEmailRequest) error {
	resp, err := c.doPost(c.BuildURL("/api/v1/users/verify-email-complete"), request)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	}
	return readAPIError(resp)
}

// Login will log a user in.
func (c *Client) Login(request *LoginRequest) (*User, error) {
	resp, err := c.doPost(c.BuildURL("/api/v1/users/login"), request)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		c.headers[HeaderAuthorization] = fmt.Sprintf("%s %s", AuthorizationBearer, resp.Header.Get(SessionHeader))
		return UserFromReader(resp.Body)
	}
	return nil, readAPIError(resp)
}

// Logout will log out the current user.
func (c *Client) Logout() error {
	resp, err := c.doPost(c.BuildURL("/api/v1/users/logout"), nil)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		delete(c.headers, HeaderAuthorization)
		return nil
	}
	return readAPIError(resp)
}

// GetMe gets the currently logged in user.
func (c *Client) GetMe() (*User, error) {
	resp, err := c.doGet(c.BuildURL("/api/v1/users/me"))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return UserFromReader(resp.Body)
	}
	return nil, readAPIError(resp)
}

// UpdateMe updates the logged in user.
func (c *Client) UpdateMe(user *User) (*User, error) {
	resp, err := c.doPut(c.BuildURL("/api/v1/users/me"), user)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return UserFromReader(resp.Body)
	}
	return nil, readAPIError(resp)
}
