package monime

import (
	"context"
	"net/http"
	"net/url"
)

func (c *Client) get(ctx context.Context, path string, query url.Values, config *RequestConfig) (*APIResponse, error) {
	response := &APIResponse{}
	err := c.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   path,
		Query:  query,
		Config: config,
	}, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) getList(ctx context.Context, path string, query url.Values, config *RequestConfig) (*APIListResponse, error) {
	response := &APIListResponse{}
	err := c.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   path,
		Query:  query,
		Config: config,
	}, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) post(ctx context.Context, path string, body any, config *RequestConfig) (*APIResponse, error) {
	response := &APIResponse{}
	err := c.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   path,
		Body:   body,
		Config: config,
	}, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) patch(ctx context.Context, path string, body any, config *RequestConfig) (*APIResponse, error) {
	response := &APIResponse{}
	err := c.request(ctx, requestOptions{
		Method: http.MethodPatch,
		Path:   path,
		Body:   body,
		Config: config,
	}, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) delete(ctx context.Context, path string, config *RequestConfig) (*APIDeleteResponse, error) {
	response := &APIDeleteResponse{}
	err := c.request(ctx, requestOptions{
		Method: http.MethodDelete,
		Path:   path,
		Config: config,
	}, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
