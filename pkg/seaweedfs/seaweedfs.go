package seaweedfs

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type Client interface {
	GetFile(ctx context.Context, path string) (*http.Response, error)
	PipeFile(ctx context.Context, path string, w http.ResponseWriter) error
}

type client struct {
	masterURL string
	filerURL  string
	c         *http.Client

	// Buffer to pipe the file from SeaweedFS to io.Writer
	pipeBuf *pipeBufferPool
}

func NewClient(masterURL string, filerURL string) Client {
	return &client{
		masterURL: masterURL,
		filerURL:  filerURL,
		c:         &http.Client{},
		pipeBuf:   newPipeBufferPool(),
	}
}

func (c *client) GetFile(ctx context.Context, path string) (*http.Response, error) {
	url, err := url.JoinPath(c.filerURL, path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.c.Do(req)
}

func (c *client) PipeFile(ctx context.Context, path string, w http.ResponseWriter) error {
	resp, err := c.GetFile(ctx, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	header := w.Header()
	for k, v := range resp.Header {
		header.Set(k, v[0])
	}
	header.Set("Server", "Otter v1.0.0")

	buf := c.pipeBuf.Get()
	defer c.pipeBuf.Release(buf)

	_, err = io.CopyBuffer(w, resp.Body, *buf)
	return err
}
