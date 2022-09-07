package neo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	client  string
	secret  string
	baseURL string
	doer    Doer
}

func NewProductionClient(client, secret string, doer Doer) *Client {
	return NewClient(client, secret, "https://api.neonomics.io", doer)
}

func NewSandboxClient(client, secret string, doer Doer) *Client {
	return NewClient(client, secret, "https://sandbox.neonomics.io", doer)
}

func NewClient(client, secret, baseURL string, doer Doer) *Client {
	return &Client{
		client:  client,
		secret:  secret,
		baseURL: baseURL,
		doer:    doer,
	}
}

type API struct {
	Client   *Client
	Token    *Token
	DeviceID string
	Mapper   SCAMapper
}

func (c *Client) API(ctx context.Context, deviceID string) (*API, error) {
	t, err := c.AccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create an API instance: %w", err)
	}
	return &API{
		Client:   c,
		Token:    t,
		DeviceID: deviceID,
		Mapper:   DefaultSCAMapper,
	}, nil
}

var (
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
	ContentTypeJSON           = "application/json"

	HeaderSessionID   = "x-session-id"
	HeaderRedirectURL = "x-redirect-url"
	HeaderPSUID       = "x-psu-id"
	HeaderPSUIP       = "x-psu-ip-address"
	HeaderDeviceID    = "x-device-id"
)

// Optional provides means to adjust the request sent to the server.
// In most cases, you should use one of the provided helpers:
// SessionID, RedirectURL, PsuID, PsuIP, DeviceID.
type Optional func(*http.Request)

// SessionID appends a session ID header to the request.
func SessionID(id string) Optional {
	return func(r *http.Request) {
		r.Header.Set(HeaderSessionID, id)
	}
}

// RedirectURL appends a redirect URL header to the request.
func RedirectURL(url string) Optional {
	return func(r *http.Request) {
		r.Header.Set(HeaderRedirectURL, url)
	}
}

// PsuID appends the payment service user ID header to the request.
// Note that this does not encrypt the provided id.
// See Encrypter & NewEncrypter.
func PsuID(id string) Optional {
	return func(r *http.Request) {
		r.Header.Set(HeaderPSUID, id)
	}
}

// PsuIP appends the payment service user IP header to the request.
func PsuIP(ip string) Optional {
	return func(r *http.Request) {
		r.Header.Set(HeaderPSUIP, ip)
	}
}

// DeviceID appends the device ID header to the request.
func DeviceID(id string) Optional {
	return func(r *http.Request) {
		r.Header.Set(HeaderDeviceID, id)
	}
}

func (a *API) request(ctx context.Context, method, url string, body io.Reader, opts ...Optional) *http.Request {
	if !strings.HasPrefix(url, "http") {
		url = a.Client.baseURL + url
	}
	r, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		panic(err)
	}
	r.Header.Set(HeaderDeviceID, a.DeviceID)
	for _, opt := range opts {
		opt(r)
	}
	// Append non-modifiable headers. This overrides any previously set headers with the same key.
	r.Header.Set("accept", ContentTypeJSON)
	r.Header.Set("authorization", "Bearer "+a.Token.AccessToken)
	r.Header.Set("content-type", ContentTypeJSON)
	return r
}

func (a *API) do(req *http.Request, status int, v interface{}) (*SCAHandler, error) { //nolint:cyclop
	resp, err := a.Client.doer.Do(req) //nolint:bodyclose
	if err != nil {
		return nil, fmt.Errorf("%s err: %w", req.URL.String(), err)
	}
	defer closeBody(resp.Body)
	switch resp.StatusCode {
	case status:
		if v != nil {
			if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
				return nil, fmt.Errorf("neo: failed to decode JSON: %w", err)
			}
		}
		return nil, nil
	case http.StatusUnauthorized:
		c := req.Context()
		t, err := a.Client.RefreshToken(c, a.Token)
		if err != nil {
			return nil, fmt.Errorf("neo: failed to refresh token: %w", err)
		}
		a.Token = t
		return a.do(req.Clone(c), status, v)
	case 510, 520, 530: // See https://docs.neonomics.io/documentation/development/error-handling.
		e := &Error{}
		if err := json.NewDecoder(resp.Body).Decode(e); err != nil {
			return nil, fmt.Errorf("neo: failed to decode neo.Error: %w", err)
		}
		if a.Mapper == nil || !(e.IsConsentError() || e.IsPaymentAuthError()) {
			return nil, e
		}
		sca, err := a.Mapper(a.Client.doer, req, e)
		if err != nil {
			return nil, fmt.Errorf("neo: SCAMapper: %w", err)
		}
		return &SCAHandler{
			SCA: sca,
			Retry: func(ctx context.Context, u interface{}) (*SCAHandler, error) {
				return a.do(req.Clone(ctx), status, u)
			},
		}, nil
	default:
		return nil, fmt.Errorf("unexpected HTTP response: %s", resp.Status)
	}
}

// closeBody ensures the body is both read & closed, as per the docs:
// https://pkg.go.dev/net/http#Client.Do.
func closeBody(b io.ReadCloser) {
	_, _ = io.Copy(io.Discard, b)
	_ = b.Close()
}
