package seaweedfs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
}

func NewClient(masterURL string, filerURL string) Client {
	return &client{
		masterURL: masterURL,
		filerURL:  filerURL,
		c:         &http.Client{},
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

func (c *client) PipeFile(ctx context.Context, path string, w http.ResponseWriter) error {
	resp, err := c.GetFile(ctx, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Copy response headers from SeaweedFS and override headers as needed
	header := w.Header()
	for k, v := range resp.Header {
		header.Set(k, v[0])
	}
	header.Set("Server", "Otter v1.0.0")

	if resp.StatusCode == http.StatusNotFound {
		w.WriteHeader(http.StatusNotFound)
		return errors.New("nothing in this path")
	}

	// If the target is a directory, return a list of contents inside the directory in JSON format
	// by checking the Etag header. This doesn't exist if the path is a directory
	if resp.Header.Get("Etag") == "" {
		var directoryList directoryList
		if err := json.NewDecoder(resp.Body).Decode(&directoryList); err != nil {
			return fmt.Errorf("cannot list directory entries: %w", err)
		}
		if err := json.NewEncoder(w).Encode(directoryList.Entries); err != nil {
			return fmt.Errorf("cannot marshal directory entries list: %w", err)
		}
	}

	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("streaming seaweedfs: %w", err)
	}
	return nil
}
