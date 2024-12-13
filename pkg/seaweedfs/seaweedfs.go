package seaweedfs

import (
	"context"
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
	pipe      *pipe
}

func NewClient(masterURL string, filerURL string) Client {
	return &client{
		masterURL: masterURL,
		filerURL:  filerURL,
		c:         &http.Client{},
		pipe:      newPipe(),
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

	// Copy response headers from SeaweedFS
	header := w.Header()
	for k, v := range resp.Header {
		header.Set(k, v[0])
	}
	header.Set("Server", "Otter v1.0.0")

	// Copy response status code from SeaweedFS
	w.WriteHeader(resp.StatusCode)

	return c.pipe.Pipe(w, resp.Body)
}
