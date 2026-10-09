package main

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Item struct {
	Path string `json:"path"`
	Mod  int64  `json:"mod"` // unix seconds
}

type MediaResponse struct {
	Photos []Item `json:"photos"`
	Videos []Item `json:"videos"`
}

var (
	photoExt = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".avif": true}
	videoExt = map[string]bool{".mp4": true, ".webm": true, ".ogg": true, ".mov": true, ".mkv": true}

	mu             sync.Mutex
	cachedResponse []byte
	lastScan       time.Time
	// Re-scan the disk at most once per interval, however many clients poll.
	scanInterval = 3 * time.Second
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func scanMedia(root string) ([]byte, error) {
	resp := MediaResponse{Photos: []Item{}, Videos: []Item{}}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // unreadable entry: skip it, keep walking
		}
		// Skip hidden files and folders (.DS_Store, ._foo.jpg, .git, ...)
		if path != root && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		isPhoto, isVideo := photoExt[ext], videoExt[ext]
		if !isPhoto && !isVideo {
			return nil
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		item := Item{Path: filepath.ToSlash(rel)}
		if info, infoErr := d.Info(); infoErr == nil {
			item.Mod = info.ModTime().Unix()
		}
		if isPhoto {
			resp.Photos = append(resp.Photos, item)
		} else {
			resp.Videos = append(resp.Videos, item)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Newest first, path as a stable tiebreaker so the order never jitters.
	newestFirst := func(s []Item) {
		sort.Slice(s, func(i, j int) bool {
			if s[i].Mod != s[j].Mod {
				return s[i].Mod > s[j].Mod
			}
			return s[i].Path < s[j].Path
		})
	}
	newestFirst(resp.Photos)
	newestFirst(resp.Videos)

	return json.Marshal(resp)
}

// mediaFS serves files but refuses directories (no autoindex) and hidden paths.
type mediaFS struct{ fs http.FileSystem }

func (m mediaFS) Open(name string) (http.File, error) {
	for _, seg := range strings.Split(name, "/") {
		if strings.HasPrefix(seg, ".") && seg != "" {
			return nil, os.ErrNotExist
		}
	}
	f, err := m.fs.Open(name)
	if err != nil {
		return nil, err
	}
	st, err := f.Stat()
	if err != nil || st.IsDir() {
		f.Close()
		return nil, os.ErrNotExist
	}
	return f, nil
}

func main() {
	mediaDir := env("MEDIA_DIR", "/media")
	addr := env("ADDR", ":8080")
	os.MkdirAll(mediaDir, 0755)

	http.HandleFunc("/api/media", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		if cachedResponse == nil || time.Since(lastScan) > scanInterval {
			if data, err := scanMedia(mediaDir); err != nil {
				log.Println("scan error:", err)
			} else {
				cachedResponse, lastScan = data, time.Now()
			}
		}
		body := cachedResponse
		mu.Unlock()

		if body == nil {
			body = []byte(`{"photos":[],"videos":[]}`)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(body)
	})

	// http.FileServer handles Range requests, so video seeking works.
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(mediaFS{http.Dir(mediaDir)})))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "index.html")
	})

	log.Printf("gallery backend listening on %s, serving %s", addr, mediaDir)
	log.Fatal(http.ListenAndServe(addr, nil))
}
