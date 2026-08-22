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
	"time"
)

const DefaultBaseURL = "https://api.ganjoor.net"

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	MaxRetries int
	RetryDelay time.Duration
}

type Poem struct {
	ID                     int       `json:"id"`
	Title                  string    `json:"title"`
	FullTitle              string    `json:"fullTitle"`
	FullURL                string    `json:"fullUrl"`
	PlainText              string    `json:"plainText"`
	HTMLText               string    `json:"htmlText"`
	SourceName             string    `json:"sourceName"`
	SourceURLSlug          string    `json:"sourceUrlSlug"`
	Language               string    `json:"language"`
	PoemSummary            string    `json:"poemSummary"`
	Category               *PageInfo `json:"category"`
	Verses                 []Verse   `json:"verses"`
	Sections               []Section `json:"sections"`
	ClaimedByMultiplePoets bool      `json:"claimedByMultiplePoets"`
	CoupletsCount          *int      `json:"coupletsCount"`
}

type PageInfo struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	FullURL string `json:"fullUrl"`
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
