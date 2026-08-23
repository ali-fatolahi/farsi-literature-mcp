package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alifatolahi/ganjoor-mcp/internal/ganjoor"
)

func TestGetPoemHandlerUsesAPIClient(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":2130,"title":"غزل شمارهٔ ۱"}`))
	}))
	defer apiServer.Close()

	service := &Server{API: ganjoor.Client{BaseURL: apiServer.URL}}
	result, poem, err := service.getPoem(context.Background(), nil, poemInput{ID: 2130})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if result != nil || poem.ID != 2130 || poem.Title != "غزل شمارهٔ ۱" {
		t.Fatalf("unexpected handler result: result=%v poem=%+v", result, poem)
	}
}

func TestSearchPoetryHandlerRejectsInvalidLimit(t *testing.T) {
	service := &Server{}
	_, _, err := service.searchPoetry(context.Background(), nil, searchInput{
		Query: "عشق", Limit: 1001,
	})
	if err == nil {
		t.Fatal("invalid limit accepted")
	}
}

func TestSearchPoetryHandlerReturnsObjectOutput(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":2130,"title":"غزل شمارهٔ ۱"}]`))
	}))
	defer apiServer.Close()

	service := &Server{API: ganjoor.Client{BaseURL: apiServer.URL}}
	_, output, err := service.searchPoetry(context.Background(), nil, searchInput{
		Query: "عشق", Limit: 5,
	})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if len(output.Results) != 1 || output.Results[0].ID != 2130 {
		t.Fatalf("unexpected output: %+v", output)
	}
	encoded, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("marshal output: %v", err)
	}
	if strings.Contains(string(encoded), `"sections":null`) {
		t.Fatalf("sparse output contains null sections: %s", encoded)
	}
}
