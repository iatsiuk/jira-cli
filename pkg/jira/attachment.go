package jira

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type httpGetFunc func(context.Context, string, Header) (*http.Response, error)

func (c *Client) getAttachmentContent(path string, get httpGetFunc) (io.ReadCloser, error) {
	res, err := get(context.Background(), path, Header{"Accept": "*/*"})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrEmptyResponse
	}
	if res.StatusCode != http.StatusOK {
		defer func() { _ = res.Body.Close() }()
		return nil, formatUnexpectedResponse(res)
	}
	return res.Body, nil
}

// GetAttachmentContent downloads attachment content by ID using v3 API.
// Uses GET /rest/api/3/attachment/content/{id}?redirect=false endpoint.
// The redirect=false parameter returns content directly instead of redirecting to S3.
func (c *Client) GetAttachmentContent(id string) (io.ReadCloser, error) {
	return c.getAttachmentContent(fmt.Sprintf("/attachment/content/%s?redirect=false", id), c.Get)
}

// GetAttachmentContentV2 downloads attachment content by ID using v2 API.
// Server installations typically don't redirect to external storage.
func (c *Client) GetAttachmentContentV2(id string) (io.ReadCloser, error) {
	return c.getAttachmentContent(fmt.Sprintf("/attachment/content/%s", id), c.GetV2)
}
