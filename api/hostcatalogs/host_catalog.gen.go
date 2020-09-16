// Code generated by "make api"; DO NOT EDIT.
package hostcatalogs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/kr/pretty"

	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/scopes"
)

type HostCatalog struct {
	Id          string                 `json:"id,omitempty"`
	ScopeId     string                 `json:"scope_id,omitempty"`
	Scope       *scopes.ScopeInfo      `json:"scope,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	CreatedTime time.Time              `json:"created_time,omitempty"`
	UpdatedTime time.Time              `json:"updated_time,omitempty"`
	Version     uint32                 `json:"version,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`

	responseBody *bytes.Buffer
	responseMap  map[string]interface{}
}

func (n HostCatalog) ResponseBody() *bytes.Buffer {
	return n.responseBody
}

func (n HostCatalog) ResponseMap() map[string]interface{} {
	return n.responseMap
}

type HostCatalogReadResult struct {
	Item         *HostCatalog
	responseBody *bytes.Buffer
	responseMap  map[string]interface{}
}

func (n HostCatalogReadResult) GetItem() interface{} {
	return n.Item
}

func (n HostCatalogReadResult) GetResponseBody() *bytes.Buffer {
	return n.responseBody
}

func (n HostCatalogReadResult) GetResponseMap() map[string]interface{} {
	return n.responseMap
}

type HostCatalogCreateResult = HostCatalogReadResult
type HostCatalogUpdateResult = HostCatalogReadResult

type HostCatalogDeleteResult struct {
	responseBody *bytes.Buffer
	responseMap  map[string]interface{}
}

func (n HostCatalogDeleteResult) GetResponseBody() *bytes.Buffer {
	return n.responseBody
}

func (n HostCatalogDeleteResult) GetResponseMap() map[string]interface{} {
	return n.responseMap
}

type HostCatalogListResult struct {
	Items        []*HostCatalog
	responseBody *bytes.Buffer
	responseMap  map[string]interface{}
}

func (n HostCatalogListResult) GetItems() interface{} {
	return n.Items
}

func (n HostCatalogListResult) GetResponseBody() *bytes.Buffer {
	return n.responseBody
}

func (n HostCatalogListResult) GetResponseMap() map[string]interface{} {
	return n.responseMap
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

func (c *Client) Create(ctx context.Context, resourceType string, scopeId string, opt ...Option) (*HostCatalogCreateResult, *api.Error, error) {
	if scopeId == "" {
		return nil, nil, fmt.Errorf("empty scopeId value passed into Create request")
	}

	opts, apiOpts := getOpts(opt...)

	if c.client == nil {
		return nil, nil, fmt.Errorf("nil client")
	}
	if resourceType == "" {
		return nil, nil, fmt.Errorf("empty resourceType value passed into Create request")
	} else {
		opts.postMap["type"] = resourceType
	}

	opts.postMap["scope_id"] = scopeId

	req, err := c.client.NewRequest(ctx, "POST", "host-catalogs", opts.postMap, apiOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating Create request: %w", err)
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
		return nil, nil, fmt.Errorf("error performing client request during Create call: %w", err)
	}

	target := new(HostCatalogCreateResult)
	target.Item = new(HostCatalog)
	apiErr, err := resp.Decode(target.Item)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding Create response: %w", err)
	}
	if apiErr != nil {
		return nil, apiErr, nil
	}
	target.responseBody = resp.Body
	target.responseMap = resp.Map
	return target, apiErr, nil
}

func (c *Client) Read(ctx context.Context, hostCatalogId string, opt ...Option) (*HostCatalogReadResult, *api.Error, error) {
	if hostCatalogId == "" {
		return nil, nil, fmt.Errorf("empty  hostCatalogId value passed into Read request")
	}
	if c.client == nil {
		return nil, nil, fmt.Errorf("nil client")
	}

	opts, apiOpts := getOpts(opt...)

	req, err := c.client.NewRequest(ctx, "GET", fmt.Sprintf("host-catalogs/%s", hostCatalogId), nil, apiOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating Read request: %w", err)
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
		return nil, nil, fmt.Errorf("error performing client request during Read call: %w", err)
	}

	target := new(HostCatalogReadResult)
	target.Item = new(HostCatalog)
	apiErr, err := resp.Decode(target.Item)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding Read response: %w", err)
	}
	if apiErr != nil {
		return nil, apiErr, nil
	}
	target.responseBody = resp.Body
	target.responseMap = resp.Map
	return target, apiErr, nil
}

func (c *Client) Update(ctx context.Context, hostCatalogId string, version uint32, opt ...Option) (*HostCatalogUpdateResult, *api.Error, error) {
	if hostCatalogId == "" {
		return nil, nil, fmt.Errorf("empty hostCatalogId value passed into Update request")
	}
	if c.client == nil {
		return nil, nil, fmt.Errorf("nil client")
	}

	opts, apiOpts := getOpts(opt...)

	if version == 0 {
		if !opts.withAutomaticVersioning {
			return nil, nil, errors.New("zero version number passed into Update request and automatic versioning not specified")
		}
		existingTarget, existingApiErr, existingErr := c.Read(ctx, hostCatalogId, opt...)
		if existingErr != nil {
			return nil, nil, fmt.Errorf("error performing initial check-and-set read: %w", existingErr)
		}
		if existingApiErr != nil {
			return nil, nil, fmt.Errorf("error from controller when performing initial check-and-set read: %s", pretty.Sprint(existingApiErr))
		}
		if existingTarget == nil {
			return nil, nil, errors.New("nil resource response found when performing initial check-and-set read")
		}
		if existingTarget.Item == nil {
			return nil, nil, errors.New("nil resource found when performing initial check-and-set read")
		}
		version = existingTarget.Item.Version
	}

	opts.postMap["version"] = version

	req, err := c.client.NewRequest(ctx, "PATCH", fmt.Sprintf("host-catalogs/%s", hostCatalogId), opts.postMap, apiOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating Update request: %w", err)
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
		return nil, nil, fmt.Errorf("error performing client request during Update call: %w", err)
	}

	target := new(HostCatalogUpdateResult)
	target.Item = new(HostCatalog)
	apiErr, err := resp.Decode(target.Item)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding Update response: %w", err)
	}
	if apiErr != nil {
		return nil, apiErr, nil
	}
	target.responseBody = resp.Body
	target.responseMap = resp.Map
	return target, apiErr, nil
}

func (c *Client) Delete(ctx context.Context, hostCatalogId string, opt ...Option) (*HostCatalogDeleteResult, *api.Error, error) {
	if hostCatalogId == "" {
		return nil, nil, fmt.Errorf("empty hostCatalogId value passed into Delete request")
	}
	if c.client == nil {
		return nil, nil, fmt.Errorf("nil client")
	}

	opts, apiOpts := getOpts(opt...)

	req, err := c.client.NewRequest(ctx, "DELETE", fmt.Sprintf("host-catalogs/%s", hostCatalogId), nil, apiOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating Delete request: %w", err)
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
		return nil, nil, fmt.Errorf("error performing client request during Delete call: %w", err)
	}

	apiErr, err := resp.Decode(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding Delete response: %w", err)
	}
	if apiErr != nil {
		return nil, apiErr, nil
	}

	target := &HostCatalogDeleteResult{
		responseBody: resp.Body,
		responseMap:  resp.Map,
	}
	return target, nil, nil
}

func (c *Client) List(ctx context.Context, scopeId string, opt ...Option) (*HostCatalogListResult, *api.Error, error) {
	if scopeId == "" {
		return nil, nil, fmt.Errorf("empty scopeId value passed into List request")
	}
	if c.client == nil {
		return nil, nil, fmt.Errorf("nil client")
	}

	opts, apiOpts := getOpts(opt...)
	opts.queryMap["scope_id"] = scopeId

	req, err := c.client.NewRequest(ctx, "GET", "host-catalogs", nil, apiOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating List request: %w", err)
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
		return nil, nil, fmt.Errorf("error performing client request during List call: %w", err)
	}

	target := new(HostCatalogListResult)
	apiErr, err := resp.Decode(target)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding List response: %w", err)
	}
	if apiErr != nil {
		return nil, apiErr, nil
	}
	target.responseBody = resp.Body
	target.responseMap = resp.Map
	return target, apiErr, nil
}
