package ganjoor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const DefaultBaseURL = "https://api.ganjoor.net"

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	MaxRetries int
	RetryDelay time.Duration
	Limiter    *RateLimiter
}

type RateLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	next     time.Time
}

func NewClient() Client {
	return Client{
		BaseURL:    DefaultBaseURL,
		MaxRetries: 3,
		RetryDelay: 500 * time.Millisecond,
		Limiter:    NewRateLimiter(250 * time.Millisecond),
	}
}

func NewRateLimiter(interval time.Duration) *RateLimiter {
	return &RateLimiter{interval: interval}
}

func (l *RateLimiter) Wait(ctx context.Context) error {
	if l == nil || l.interval <= 0 {
		return nil
	}
	l.mu.Lock()
	now := time.Now()
	wait := time.Until(l.next)
	if wait < 0 {
		wait = 0
	}
	l.next = now.Add(wait).Add(l.interval)
	l.mu.Unlock()
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type Poem struct {
	ID                     int               `json:"id"`
	Title                  string            `json:"title"`
	FullTitle              string            `json:"fullTitle"`
	FullURL                string            `json:"fullUrl"`
	PlainText              string            `json:"plainText"`
	HTMLText               string            `json:"htmlText"`
	SourceName             string            `json:"sourceName"`
	SourceURLSlug          string            `json:"sourceUrlSlug"`
	Language               string            `json:"language"`
	PoemSummary            string            `json:"poemSummary"`
	Category               *CategoryComplete `json:"category,omitempty"`
	Verses                 []Verse           `json:"verses,omitempty"`
	Sections               []Section         `json:"sections,omitempty"`
	ClaimedByMultiplePoets bool              `json:"claimedByMultiplePoets"`
	CoupletsCount          *int              `json:"coupletsCount,omitempty"`
}

type Poet struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Nickname    string `json:"nickname"`
	Description string `json:"description"`
	FullURL     string `json:"fullUrl"`
	ImageURL    string `json:"imageUrl"`
}

type Category struct {
	ID       int    `json:"id"`
	PoetID   int    `json:"poetId"`
	ParentID int    `json:"parentId"`
	Title    string `json:"title"`
	FullURL  string `json:"fullUrl"`
}

type CategoryComplete struct {
	Poet *Poet    `json:"poet"`
	Cat  Category `json:"cat"`
}

type Verse struct {
	VOrder       int    `json:"vOrder"`
	Position     string `json:"position"`
	Text         string `json:"text"`
	CoupletIndex int    `json:"coupletIndex"`
	SectionIndex int    `json:"sectionIndex1"`
}

type Section struct {
	Index         int    `json:"index"`
	Number        int    `json:"number"`
	SectionType   string `json:"sectionType"`
	VerseType     string `json:"verseType"`
	RhymeLetters  string `json:"rhymeLetters"`
	PlainText     string `json:"plainText"`
	HTMLText      string `json:"htmlText"`
	CoupletsCount int    `json:"coupletsCount"`
}

type Provenance struct {
	PoemID                 int    `json:"poem_id"`
	PoetID                 int    `json:"poet_id"`
	CategoryID             int    `json:"category_id"`
	PoemURL                string `json:"poem_url"`
	SourceName             string `json:"source_name"`
	SourceURLSlug          string `json:"source_url_slug"`
	ClaimedByMultiplePoets bool   `json:"claimed_by_multiple_poets"`
}

type Context struct {
	PoemID     int        `json:"poem_id"`
	VerseStart int        `json:"verse_start"`
	VerseEnd   int        `json:"verse_end"`
	Verses     []Verse    `json:"verses"`
	Provenance Provenance `json:"provenance"`
}

type APIError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("ganjoor API: %s", e.Status)
	}
	return fmt.Sprintf("ganjoor API: %s: %s", e.Status, e.Body)
}

func (c Client) GetPoem(ctx context.Context, id int) (Poem, error) {
	if id < 1 {
		return Poem{}, errors.New("poem ID must be positive")
	}
	query := url.Values{
		"catInfo":      {"true"},
		"catPoems":     {"false"},
		"rhymes":       {"false"},
		"recitations":  {"false"},
		"images":       {"false"},
		"songs":        {"false"},
		"comments":     {"false"},
		"verseDetails": {"true"},
		"navigation":   {"false"},
		"relatedpoems": {"false"},
	}
	var poem Poem
	if err := c.getJSON(ctx, "/api/ganjoor/poem/"+strconv.Itoa(id), query, &poem); err != nil {
		return Poem{}, err
	}
	return poem, nil
}

func (c Client) GetPoet(ctx context.Context, id int) (Poet, error) {
	if id < 1 {
		return Poet{}, errors.New("poet ID must be positive")
	}
	var response struct {
		Poet Poet `json:"poet"`
	}
	if err := c.getJSON(ctx, "/api/ganjoor/poet/"+strconv.Itoa(id), url.Values{"catPoems": {"false"}}, &response); err != nil {
		return Poet{}, err
	}
	return response.Poet, nil
}

func (c Client) GetCategory(ctx context.Context, id int) (Category, error) {
	if id < 1 {
		return Category{}, errors.New("category ID must be positive")
	}
	query := url.Values{"poems": {"false"}, "mainSections": {"false"}}
	var response struct {
		Category Category `json:"cat"`
	}
	if err := c.getJSON(ctx, "/api/ganjoor/cat/"+strconv.Itoa(id), query, &response); err != nil {
		return Category{}, err
	}
	return response.Category, nil
}

func (c Client) SearchPoems(ctx context.Context, term string, poetID, page, pageSize int) ([]Poem, error) {
	if strings.TrimSpace(term) == "" {
		return nil, errors.New("search term must not be empty")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 20
	}
	query := url.Values{
		"term":       {term},
		"PageNumber": {strconv.Itoa(page)},
		"PageSize":   {strconv.Itoa(pageSize)},
		"poetId":     {strconv.Itoa(poetID)},
	}
	var poems []Poem
	if err := c.getJSON(ctx, "/api/ganjoor/poems/search", query, &poems); err != nil {
		return nil, err
	}
	return poems, nil
}

func (c Client) GetContext(ctx context.Context, poemID, verseOrder, before, after int) (Context, error) {
	if poemID < 1 {
		return Context{}, errors.New("poem ID must be positive")
	}
	if verseOrder < 0 || before < 0 || after < 0 {
		return Context{}, errors.New("verse order and context sizes must not be negative")
	}
	poem, err := c.GetPoem(ctx, poemID)
	if err != nil {
		return Context{}, err
	}
	if verseOrder > 0 && verseOrder > len(poem.Verses) {
		return Context{}, fmt.Errorf("verse order %d is outside poem %d", verseOrder, poemID)
	}
	start, end := 0, len(poem.Verses)
	if verseOrder > 0 {
		center := verseOrder - 1
		start = center - before
		if start < 0 {
			start = 0
		}
		end = center + after + 1
		if end > len(poem.Verses) {
			end = len(poem.Verses)
		}
	}
	verses := append([]Verse(nil), poem.Verses[start:end]...)
	verseStart, verseEnd := 0, 0
	if len(verses) > 0 {
		verseStart = verses[0].VOrder
		verseEnd = verses[len(verses)-1].VOrder
	}
	return Context{
		PoemID: poem.ID, VerseStart: verseStart, VerseEnd: verseEnd,
		Verses: verses, Provenance: provenanceFor(poem),
	}, nil
}

func (c Client) GetProvenance(ctx context.Context, poemID int) (Provenance, error) {
	if poemID < 1 {
		return Provenance{}, errors.New("poem ID must be positive")
	}
	poem, err := c.GetPoem(ctx, poemID)
	if err != nil {
		return Provenance{}, err
	}
	return provenanceFor(poem), nil
}

func provenanceFor(poem Poem) Provenance {
	var poetID, categoryID int
	if poem.Category != nil {
		if poem.Category.Poet != nil {
			poetID = poem.Category.Poet.ID
		}
		categoryID = poem.Category.Cat.ID
	}
	return Provenance{
		PoemID: poem.ID, PoetID: poetID, CategoryID: categoryID,
		PoemURL:    "https://ganjoor.net" + poem.FullURL,
		SourceName: poem.SourceName, SourceURLSlug: poem.SourceURLSlug,
		ClaimedByMultiplePoets: poem.ClaimedByMultiplePoets,
	}
}

func (c Client) getJSON(ctx context.Context, endpoint string, query url.Values, target any) error {
	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	maxRetries := c.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	retryDelay := c.RetryDelay
	if retryDelay <= 0 {
		retryDelay = 500 * time.Millisecond
	}

	requestURL := strings.TrimRight(baseURL, "/") + endpoint
	if encoded := query.Encode(); encoded != "" {
		requestURL += "?" + encoded
	}
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := c.Limiter.Wait(ctx); err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/json")
		resp, err := httpClient.Do(req)
		if err != nil {
			if attempt == maxRetries {
				return fmt.Errorf("request %s: %w", endpoint, err)
			}
			if err := sleep(ctx, retryDelay, attempt); err != nil {
				return err
			}
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read %s response: %w", endpoint, readErr)
		}
		if resp.StatusCode == http.StatusOK {
			if err := json.Unmarshal(body, target); err != nil {
				return fmt.Errorf("decode %s response: %w", endpoint, err)
			}
			return nil
		}
		apiErr := &APIError{StatusCode: resp.StatusCode, Status: resp.Status, Body: strings.TrimSpace(string(body))}
		if !retryable(resp.StatusCode) || attempt == maxRetries {
			return apiErr
		}
		if err := sleep(ctx, retryDelay, attempt); err != nil {
			return err
		}
	}
	return fmt.Errorf("request %s: retry limit exceeded", endpoint)
}

func retryable(status int) bool {
	return status == http.StatusTooManyRequests || status >= http.StatusInternalServerError
}

func sleep(ctx context.Context, base time.Duration, attempt int) error {
	delay := base * time.Duration(1<<min(attempt, 5))
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
