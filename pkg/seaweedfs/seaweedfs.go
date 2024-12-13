package seaweedfs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type directoryList struct {
	Path    string           `json:"Path"`
	Entries []directoryEntry `json:"Entries"`
}

type directoryEntry struct {
	FullPath     string    `json:"FullPath"`
	Size         int       `json:"FileSize"`
	Mime         string    `json:"Mime"`
	ModifiedTime time.Time `json:"Mtime"`
	CreatedTime  time.Time `json:"Crtime"`
}

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
	req.Header.Set("Accept", "application/json")
	return c.c.Do(req)
}

func (c *client) PipeFile(ctx context.Context, path string, w http.ResponseWriter) (err error) {
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

	if resp.StatusCode == http.StatusNotFound {
		w.WriteHeader(http.StatusNotFound)
		return errors.New("nothing in this path")
	}

	// Directory doesn't have Etag header
	if resp.Header.Get("Etag") == "" {
		var directoryList directoryList
		if err := json.NewDecoder(resp.Body).Decode(&directoryList); err != nil {
			return fmt.Errorf("cannot list directory entries: %w", err)
		}
		output, err := json.Marshal(directoryList.Entries)
		if err != nil {
			return fmt.Errorf("cannot marshal directory entries list: %w", err)
		}
		header.Set("Content-Length", strconv.Itoa(len(output)))
		return c.pipe.Pipe(w, bytes.NewReader(output))
	}

	return c.pipe.Pipe(w, resp.Body)
}
