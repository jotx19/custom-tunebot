package musicapi

import (
	"errors"
	"fmt"
	"strings"
)

func NormalizeSearchSongs(raw any) ([]SongLite, error) {
	switch t := raw.(type) {
	case []any:
		return songsFromArray(t), nil
	case map[string]any:
		if data, ok := t["data"]; ok {
			switch d := data.(type) {
			case []any:
				return songsFromArray(d), nil
			case map[string]any:
				if res, ok := d["results"].([]any); ok {
					return songsFromArray(res), nil
				}
			}
		}
		if res, ok := t["results"].([]any); ok {
			return songsFromArray(res), nil
		}
	}
	return nil, errors.New("could not parse search response (JSON shape not recognized)")
}

func NormalizeSongDetail(raw any) (SongDetail, error) {
	switch t := raw.(type) {
	case map[string]any:
		if data, ok := t["data"]; ok {
			switch d := data.(type) {
			case map[string]any:
				return songFromObj(d), nil
			case []any:
				if len(d) > 0 {
					if obj, ok := d[0].(map[string]any); ok {
						return songFromObj(obj), nil
					}
				}
			}
		}
		return songFromObj(t), nil
	}
	return SongDetail{}, fmt.Errorf("could not parse song detail")
}

func songsFromArray(arr []any) []SongLite {
	out := make([]SongLite, 0, len(arr))
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		s := songFromObj(obj)
		if s.ID != "" && s.Title != "" {
			out = append(out, s)
		}
	}
	return out
}

func songFromObj(obj map[string]any) SongLite {
	// Basic fields
	id := firstString(obj, "id", "song_id", "_id")
	title := firstString(obj, "title", "name", "song_name")

	// ✅ Artist: support many shapes so it doesn't become Unknown Artist
	artist := firstString(obj,
		"artist",
		"artists",
		"primaryArtists",
		"primary_artists",
		"subtitle", // often "Artist • Album"
		"song_artist",
	)
	if artist == "" {
		artist = nestedArtist(obj)
	}
	artist = cleanArtist(artist)

	link := firstString(obj, "link", "url", "perma_url")
	image := firstString(obj, "image", "thumbnail", "cover")

	// Audio / stream URL candidates
	stream := firstString(obj, "stream", "stream_url", "audio", "audio_url", "download_url", "downloadUrl")

	// image sometimes array: [{quality,url}]
	if image == "" {
		if arr, ok := obj["image"].([]any); ok {
			image = bestUrlFromQualityArray(arr)
		}
	}

	// stream sometimes array: downloadUrl: [{quality,url}]
	if stream == "" {
		if arr, ok := obj["downloadUrl"].([]any); ok {
			stream = bestUrlFromQualityArray(arr)
		}
		if arr, ok := obj["download_url"].([]any); ok {
			stream = bestUrlFromQualityArray(arr)
		}
	}

	return SongLite{
		ID:        id,
		Title:     title,
		Artist:    artist,
		Image:     image,
		Link:      link,
		StreamURL: stream,
	}
}

func firstString(obj map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := obj[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func bestUrlFromQualityArray(arr []any) string {
	var last string
	for _, it := range arr {
		obj, ok := it.(map[string]any)
		if !ok {
			continue
		}
		if u, ok := obj["url"].(string); ok && u != "" {
			last = u
		}
	}
	return last
}

// --- Artist helpers ---

func nestedArtist(obj map[string]any) string {
	// artists: [{name:"..."}]
	if arr, ok := obj["artists"].([]any); ok {
		for _, it := range arr {
			if m, ok := it.(map[string]any); ok {
				if name, ok := m["name"].(string); ok && strings.TrimSpace(name) != "" {
					return strings.TrimSpace(name)
				}
			}
		}
	}

	// artists: { primary: [{name:""}] } OR { all: [...] }
	if m, ok := obj["artists"].(map[string]any); ok {
		if prim, ok := m["primary"].([]any); ok {
			for _, it := range prim {
				if mm, ok := it.(map[string]any); ok {
					if name, ok := mm["name"].(string); ok && strings.TrimSpace(name) != "" {
						return strings.TrimSpace(name)
					}
				}
			}
		}
		if all, ok := m["all"].([]any); ok {
			for _, it := range all {
				if mm, ok := it.(map[string]any); ok {
					if name, ok := mm["name"].(string); ok && strings.TrimSpace(name) != "" {
						return strings.TrimSpace(name)
					}
				}
			}
		}
	}

	// primaryArtists might be an array sometimes
	if arr, ok := obj["primaryArtists"].([]any); ok {
		for _, it := range arr {
			if m, ok := it.(map[string]any); ok {
				if name, ok := m["name"].(string); ok && strings.TrimSpace(name) != "" {
					return strings.TrimSpace(name)
				}
			}
		}
	}

	return ""
}

func cleanArtist(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	// Common format: "Artist • Album"
	if strings.Contains(s, "•") {
		parts := strings.Split(s, "•")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	// Sometimes: "Artist - Something"
	if strings.Contains(s, " - ") {
		parts := strings.Split(s, " - ")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	return s
}
