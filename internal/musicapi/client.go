package musicapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Client struct {
	Base   string
	Prefix string
	http   *http.Client
}

func New(base, prefix string) *Client {
	return &Client{
		Base:   strings.TrimRight(base, "/"),
		Prefix: "/" + strings.Trim(prefix, "/"),
		http:   &http.Client{Timeout: 12 * time.Second},
	}
}

func (c *Client) SearchSongs(query string) ([]SongLite, error) {
	u, _ := url.Parse(c.Base)
	u.Path = path.Join(u.Path, c.Prefix, "/search/songs")
	q := u.Query()
	q.Set("query", query)
	u.RawQuery = q.Encode()

	raw, err := c.getJSON(u.String())
	if err != nil {
		return nil, err
	}
	return NormalizeSearchSongs(raw)
}

func (c *Client) GetSongByID(id string) (*SongDetail, error) {
	u, _ := url.Parse(c.Base)
	// ✅ IMPORTANT: /songs/{id}
	u.Path = path.Join(u.Path, c.Prefix, "/songs", id)

	raw, err := c.getJSON(u.String())
	if err != nil {
		return nil, err
	}
	d, err := NormalizeSongDetail(raw)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (c *Client) getJSON(fullURL string) (any, error) {
	resp, err := c.http.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api status %d", resp.StatusCode)
	}

	var v any
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}
