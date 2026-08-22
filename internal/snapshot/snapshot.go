package snapshot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	ManifestName = "manifest.json"
	RawDir       = "raw"
)

type Manifest struct {
	SchemaVersion    int            `json:"SchemaVersion"`
	GeneratedAtUTC   string         `json:"GeneratedAtUtc"`
	PoetsCount       int            `json:"PoetsCount"`
	PoemsCount       int            `json:"PoemsCount"`
	IDIndexShardSize int            `json:"IdIndexShardSize"`
	URLTemplates     URLTemplates   `json:"UrlTemplates"`
	Poets            []ManifestPoet `json:"Poets"`
}

type URLTemplates struct {
	Poet            string `json:"Poet"`
	Category        string `json:"Category"`
	Poem            string `json:"Poem"`
	PoetIDIndex     string `json:"PoetIdIndex"`
	CategoryIDShard string `json:"CatIdIndexShard"`
	PoemIDShard     string `json:"PoemIdIndexShard"`
}

type ManifestPoet struct {
	ID       int    `json:"Id"`
	Nickname string `json:"Nickname"`
	FullURL  string `json:"FullUrl"`
}

type Category struct {
	ID        int            `json:"Id"`
	FullURL   string         `json:"FullUrl"`
	ChildCats []CategoryLink `json:"ChildCats"`
	Poems     []PoemLink     `json:"Poems"`
}

type CategoryLink struct {
	FullURL string `json:"FullUrl"`
}

type PoemLink struct {
	ID      int    `json:"Id"`
	FullURL string `json:"FullUrl"`
}

type Metadata struct {
	Commit         string    `json:"commit"`
	GeneratedAtUTC string    `json:"generated_at_utc"`
	SchemaVersion  int       `json:"schema_version"`
	PoetsCount     int       `json:"poets_count"`
	PoemsCount     int       `json:"poems_count"`
	DownloadedAt   time.Time `json:"downloaded_at"`
	Files          int       `json:"files"`
}

type Downloader struct {
	Client  *http.Client
	BaseURL string
	Commit  string
	Output  string
}

func (d Downloader) Download(ctx context.Context) (Metadata, error) {
	if d.Client == nil {
		d.Client = http.DefaultClient
	}
	if !isSafeCommit(d.Commit) {
		return Metadata{}, errors.New("commit must contain only hexadecimal characters")
	}
	if d.Output == "" {
		return Metadata{}, errors.New("output directory is required")
	}

	manifestURL := d.url(ManifestName)
	var manifest Manifest
	if err := d.fetchJSON(ctx, manifestURL, filepath.Join(d.Output, RawDir, ManifestName), &manifest); err != nil {
		return Metadata{}, fmt.Errorf("fetch manifest: %w", err)
	}
	if err := validateManifest(manifest); err != nil {
		return Metadata{}, fmt.Errorf("validate manifest: %w", err)
	}

	files := 1
	seen := map[string]bool{ManifestName: true}
	for _, poet := range manifest.Poets {
		poetPath, err := contentPath(poet.FullURL, "poet.json")
		if err != nil {
			return Metadata{}, fmt.Errorf("poet %d: %w", poet.ID, err)
		}
		if !seen[poetPath] {
			if err := d.fetchRaw(ctx, d.url(poetPath), filepath.Join(d.Output, RawDir, poetPath)); err != nil {
				return Metadata{}, fmt.Errorf("fetch poet %d: %w", poet.ID, err)
			}
			seen[poetPath] = true
			files++
		}
		rootPath, err := contentPath(poet.FullURL, "_cat.json")
		if err != nil {
			return Metadata{}, fmt.Errorf("poet %d root category: %w", poet.ID, err)
		}
		count, err := d.downloadCategoryTree(ctx, rootPath, seen)
		if err != nil {
			return Metadata{}, fmt.Errorf("poet %d categories: %w", poet.ID, err)
		}
		files += count
	}

	metadata := Metadata{
		Commit: d.Commit, GeneratedAtUTC: manifest.GeneratedAtUTC,
		SchemaVersion: manifest.SchemaVersion, PoetsCount: manifest.PoetsCount,
		PoemsCount: manifest.PoemsCount, DownloadedAt: time.Now().UTC(), Files: files,
	}
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return Metadata{}, fmt.Errorf("encode snapshot metadata: %w", err)
	}
	if err := os.WriteFile(filepath.Join(d.Output, "snapshot.json"), append(data, '\n'), 0o644); err != nil {
		return Metadata{}, fmt.Errorf("write snapshot metadata: %w", err)
	}
	return metadata, nil
}

func (d Downloader) downloadCategoryTree(ctx context.Context, categoryPath string, seen map[string]bool) (int, error) {
	if seen[categoryPath] {
		return 0, nil
	}
	var category Category
	if err := d.fetchJSON(ctx, d.url(categoryPath), filepath.Join(d.Output, RawDir, categoryPath), &category); err != nil {
		return 0, err
	}
	seen[categoryPath] = true
	count := 1

	poemPaths := make([]string, 0, len(category.Poems))
	for _, poem := range category.Poems {
		poemPath, err := contentPath(poem.FullURL, ".json")
		if err != nil {
			return 0, fmt.Errorf("poem %d: %w", poem.ID, err)
		}
		if seen[poemPath] {
			continue
		}
		seen[poemPath] = true
		poemPaths = append(poemPaths, poemPath)
	}
	if err := d.fetchParallel(ctx, poemPaths); err != nil {
		return 0, err
	}
	count += len(poemPaths)
	for _, child := range category.ChildCats {
		childPath, err := contentPath(child.FullURL, "_cat.json")
		if err != nil {
			return 0, err
		}
		childCount, err := d.downloadCategoryTree(ctx, childPath, seen)
		if err != nil {
			return 0, err
		}
		count += childCount
	}
	return count, nil
}

const (
	downloadWorkers = 4
	maxAttempts     = 6
)

func (d Downloader) fetchParallel(ctx context.Context, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	jobs := make(chan string)
	errs := make(chan error, len(paths))
	var workers sync.WaitGroup
	workerCount := downloadWorkers
	if len(paths) < workerCount {
		workerCount = len(paths)
	}
	for i := 0; i < workerCount; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for filePath := range jobs {
				if err := d.fetchRaw(ctx, d.url(filePath), filepath.Join(d.Output, RawDir, filePath)); err != nil {
					errs <- fmt.Errorf("fetch %s: %w", filePath, err)
				}
			}
		}()
	}
	for _, filePath := range paths {
		select {
		case jobs <- filePath:
		case <-ctx.Done():
			close(jobs)
			workers.Wait()
			return ctx.Err()
		}
	}
	close(jobs)
	workers.Wait()
	close(errs)
	for err := range errs {
		return err
	}
	return nil
}

func (d Downloader) url(filePath string) string {
	return strings.TrimRight(d.BaseURL, "/") + "/" + d.Commit + "/" + filePath
}

func (d Downloader) fetchJSON(ctx context.Context, url, destination string, target any) error {
	if err := d.fetchRaw(ctx, url, destination); err != nil {
		return err
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode %s: %w", url, err)
	}
	return nil
}

func (d Downloader) fetchRaw(ctx context.Context, url, destination string) error {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		resp, err := d.Client.Do(req)
		if err != nil {
			if attempt == maxAttempts {
				return err
			}
			if err := waitRetry(ctx, attempt, ""); err != nil {
				return err
			}
			continue
		}
		if resp.StatusCode == http.StatusOK {
			err := writeResponse(resp, destination)
			resp.Body.Close()
			return err
		}
		retryAfter := resp.Header.Get("Retry-After")
		resp.Body.Close()
		if (resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500) || attempt == maxAttempts {
			return fmt.Errorf("GET %s: HTTP %s", url, resp.Status)
		}
		if err := waitRetry(ctx, attempt, retryAfter); err != nil {
			return err
		}
	}
	return fmt.Errorf("GET %s: retry limit exceeded", url)
}

func writeResponse(resp *http.Response, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, resp.Body); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func waitRetry(ctx context.Context, attempt int, retryAfter string) error {
	delay := time.Duration(1<<min(attempt-1, 5)) * time.Second
	if seconds, err := time.ParseDuration(retryAfter + "s"); err == nil && retryAfter != "" && seconds > delay {
		delay = seconds
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func validateManifest(m Manifest) error {
	if m.SchemaVersion != 1 {
		return fmt.Errorf("unsupported schema version %d", m.SchemaVersion)
	}
	if m.GeneratedAtUTC == "" || m.PoetsCount != len(m.Poets) || m.PoemsCount < 1 {
		return errors.New("missing or inconsistent counts and generation timestamp")
	}
	if m.IDIndexShardSize < 1 || m.URLTemplates.Poet == "" || m.URLTemplates.Category == "" || m.URLTemplates.Poem == "" {
		return errors.New("missing URL templates or shard size")
	}
	if len(m.Poets) == 0 {
		return errors.New("manifest contains no poets")
	}
	return nil
}

func contentPath(fullURL, suffix string) (string, error) {
	clean := path.Clean(strings.TrimPrefix(fullURL, "/"))
	if clean == "." || strings.HasPrefix(clean, "../") || clean == ".." || strings.Contains(clean, "/../") {
		return "", fmt.Errorf("invalid upstream path %q", fullURL)
	}
	var filePath string
	switch suffix {
	case "poet.json", "_cat.json":
		filePath = path.Join(clean, suffix)
	case ".json":
		filePath = clean + suffix
	default:
		return "", fmt.Errorf("unsupported content suffix %q", suffix)
	}
	return path.Join("poets", filePath), nil
}

func isSafeCommit(commit string) bool {
	if commit == "" {
		return false
	}
	for _, r := range commit {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
