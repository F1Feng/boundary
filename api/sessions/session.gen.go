// Code generated by "make api"; DO NOT EDIT.
package sessions

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/scopes"
)

type Session struct {
	Id                string            `json:"id,omitempty"`
	TargetId          string            `json:"target_id,omitempty"`
	Scope             *scopes.ScopeInfo `json:"scope,omitempty"`
	CreatedTime       time.Time         `json:"created_time,omitempty"`
	UpdatedTime       time.Time         `json:"updated_time,omitempty"`
	Version           uint32            `json:"version,omitempty"`
	Type              string            `json:"type,omitempty"`
	ExpirationTime    time.Time         `json:"expiration_time,omitempty"`
	AuthTokenId       string            `json:"auth_token_id,omitempty"`
	UserId            string            `json:"user_id,omitempty"`
	HostSetId         string            `json:"host_set_id,omitempty"`
	HostId            string            `json:"host_id,omitempty"`
	ScopeId           string            `json:"scope_id,omitempty"`
	Endpoint          string            `json:"endpoint,omitempty"`
	States            []*SessionState   `json:"states,omitempty"`
	Status            string            `json:"status,omitempty"`
	WorkerInfo        []*WorkerInfo     `json:"worker_info,omitempty"`
	Certificate       []byte            `json:"certificate,omitempty"`
	TerminationReason string            `json:"termination_reason,omitempty"`
	AuthorizedActions []string          `json:"authorized_actions,omitempty"`
	Connections       []*Connection     `json:"connections,omitempty"`

	response *api.Response
}

type SessionReadResult struct {
	Item     *Session
	response *api.Response
}

func (n SessionReadResult) GetItem() interface{} {
	return n.Item
}

func (n SessionReadResult) GetResponse() *api.Response {
	return n.response
}

type SessionCreateResult = SessionReadResult
type SessionUpdateResult = SessionReadResult

type SessionDeleteResult struct {
	response *api.Response
}

// GetItem will always be nil for SessionDeleteResult
func (n SessionDeleteResult) GetItem() interface{} {
	return nil
}

func (n SessionDeleteResult) GetResponse() *api.Response {
	return n.response
}

type SessionListResult struct {
	Items    []*Session
	response *api.Response
}

func (n SessionListResult) GetItems() interface{} {
	return n.Items
}

func (n SessionListResult) GetResponse() *api.Response {
	return n.response
}

// Client is a client for this collection
type Client struct {
	client *api.Client
}

// Creates a new client for this collection. The submitted API client is cloned;
// modifications to it after generating this client will not have effect. If you
// need to make changes to the underlying API client, use ApiClient() to access
// it.
func NewClient(c *api.Client) *Client {
	return &Client{client: c.Clone()}
}

// ApiClient returns the underlying API client
func (c *Client) ApiClient() *api.Client {
	return c.client
}

func (c *Client) Read(ctx context.Context, id string, opt ...Option) (*SessionReadResult, error) {
	if id == "" {
		return nil, fmt.Errorf("empty id value passed into Read request")
	}
	if c.client == nil {
		return nil, fmt.Errorf("nil client")
	}

	opts, apiOpts := getOpts(opt...)

	req, err := c.client.NewRequest(ctx, "GET", fmt.Sprintf("sessions/%s", id), nil, apiOpts...)
	if err != nil {
		return nil, fmt.Errorf("error creating Read request: %w", err)
	}

	if len(opts.queryMap) > 0 {
		q := url.Values{}
		for k, v := range opts.queryMap {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.client.Do(req, apiOpts...)
	if err != nil {
		return nil, fmt.Errorf("error performing client request during Read call: %w", err)
	}

	target := new(SessionReadResult)
	target.Item = new(Session)
	apiErr, err := resp.Decode(target.Item)
	if err != nil {
		return nil, fmt.Errorf("error decoding Read response: %w", err)
	}
	if apiErr != nil {
		return nil, apiErr
	}
	target.response = resp
	return target, nil
}

func (c *Client) List(ctx context.Context, scopeId string, opt ...Option) (*SessionListResult, error) {
	if scopeId == "" {
		return nil, fmt.Errorf("empty scopeId value passed into List request")
	}
	if c.client == nil {
		return nil, fmt.Errorf("nil client")
	}

	opts, apiOpts := getOpts(opt...)
	opts.queryMap["scope_id"] = scopeId

	req, err := c.client.NewRequest(ctx, "GET", "sessions", nil, apiOpts...)
	if err != nil {
		return nil, fmt.Errorf("error creating List request: %w", err)
	}

	if len(opts.queryMap) > 0 {
		q := url.Values{}
		for k, v := range opts.queryMap {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing client request during List call: %w", err)
	}

	target := new(SessionListResult)
	apiErr, err := resp.Decode(target)
	if err != nil {
		return nil, fmt.Errorf("error decoding List response: %w", err)
	}
	if apiErr != nil {
		return nil, apiErr
	}
	target.response = resp
	return target, nil
}
