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
