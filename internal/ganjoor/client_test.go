package ganjoor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetPoemRetriesTransientFailure(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			http.Error(w, "temporary failure", http.StatusServiceUnavailable)
			return
		}
		if !strings.Contains(r.URL.RawQuery, "verseDetails=true") {
			t.Errorf("missing verseDetails query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":2130,"title":"test","verses":[{"vOrder":1,"text":"سلام"}]}`))
	}))
	defer server.Close()

	client := Client{BaseURL: server.URL, MaxRetries: 1, RetryDelay: time.Millisecond}
	poem, err := client.GetPoem(context.Background(), 2130)
	if err != nil {
		t.Fatalf("GetPoem returned error: %v", err)
	}
	if poem.ID != 2130 || len(poem.Verses) != 1 {
		t.Fatalf("unexpected poem: %+v", poem)
	}
	if calls.Load() != 2 {
		t.Fatalf("got %d calls, want 2", calls.Load())
	}
}

func TestGetPoemRejectsInvalidID(t *testing.T) {
	if _, err := (Client{}).GetPoem(context.Background(), 0); err == nil {
		t.Fatal("invalid ID accepted")
	}
}

func TestRateLimiterHonorsInterval(t *testing.T) {
	limiter := NewRateLimiter(10 * time.Millisecond)
	start := time.Now()
	if err := limiter.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := limiter.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed < 10*time.Millisecond {
		t.Fatalf("waited %s, want at least 10ms", elapsed)
	}
}

func TestGetContextExpandsAroundVerse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":2130,"fullUrl":"/hafez/ghazal/sh1",
			"verses":[
				{"vOrder":1,"text":"یک"},
				{"vOrder":2,"text":"دو"},
				{"vOrder":3,"text":"سه"},
				{"vOrder":4,"text":"چهار"}
			]
		}`))
	}))
	defer server.Close()

	contextResult, err := (Client{BaseURL: server.URL}).GetContext(context.Background(), 2130, 2, 1, 1)
	if err != nil {
		t.Fatalf("GetContext returned error: %v", err)
	}
	if contextResult.VerseStart != 1 || contextResult.VerseEnd != 3 || len(contextResult.Verses) != 3 {
		t.Fatalf("unexpected context: %+v", contextResult)
	}
}

func TestLookupMethodsDecodeResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/ganjoor/poet/2":
			_, _ = w.Write([]byte(`{"poet":{"id":2,"name":"حافظ شیرازی","nickname":"حافظ","fullUrl":"/hafez"}}`))
		case "/api/ganjoor/cat/24":
			_, _ = w.Write([]byte(`{"poet":{"id":2},"cat":{"id":24,"title":"غزلیات","fullUrl":"/hafez/ghazal"}}`))
		case "/api/ganjoor/poems/search":
			_, _ = w.Write([]byte(`[{"id":2130,"title":"غزل شمارهٔ ۱","fullUrl":"/hafez/ghazal/sh1"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client := Client{BaseURL: server.URL}

	poet, err := client.GetPoet(context.Background(), 2)
	if err != nil || poet.ID != 2 || poet.Nickname != "حافظ" {
		t.Fatalf("unexpected poet: %+v, error: %v", poet, err)
	}
	category, err := client.GetCategory(context.Background(), 24)
	if err != nil || category.ID != 24 || category.Title != "غزلیات" {
		t.Fatalf("unexpected category: %+v, error: %v", category, err)
	}
	poems, err := client.SearchPoems(context.Background(), "عشق", 2, 1, 10)
	if err != nil || len(poems) != 1 || poems[0].ID != 2130 {
		t.Fatalf("unexpected search results: %+v, error: %v", poems, err)
	}
}

func TestGetProvenanceMapsNestedMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":2130,"fullUrl":"/hafez/ghazal/sh1",
			"sourceName":"ویکی‌درج","sourceUrlSlug":"wikidorj",
			"claimedByMultiplePoets":true,
			"category":{"poet":{"id":2},"cat":{"id":24}}
		}`))
	}))
	defer server.Close()

	provenance, err := (Client{BaseURL: server.URL}).GetProvenance(context.Background(), 2130)
	if err != nil {
		t.Fatalf("GetProvenance returned error: %v", err)
	}
	if provenance.PoetID != 2 || provenance.CategoryID != 24 ||
		provenance.PoemURL != "https://ganjoor.net/hafez/ghazal/sh1" ||
		!provenance.ClaimedByMultiplePoets {
		t.Fatalf("unexpected provenance: %+v", provenance)
	}
}
