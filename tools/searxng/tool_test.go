package searxng

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func startSearxngServer(t *testing.T, port int, results *Output) *http.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(results)
		io.Copy(w, buf)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/search", handler)
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			t.Errorf("start searxng server failed: %v", err)
		}
	}()
	return srv
}

func TestSearxngSearchWithCategory(t *testing.T) {
	mockPort := 8080
	mockSearchURL := fmt.Sprintf("http://localhost:%d", mockPort)
	mockQuery := "test query with category"
	mockItem := SearchResultItem{
		URL:      "https://example.com/test-category",
		Title:    "Test Result with Category",
		Content:  "This is a test result content with category.",
		Category: NewsCategory,
	}
	mockResult := Output{
		Results: []SearchResultItem{mockItem},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := startSearxngServer(t, mockPort, &mockResult)
	defer srv.Shutdown(ctx)
	// Initialize the tool
	tool := New(WithBaseURL(mockSearchURL))
	input := NewInput(NewsCategory, []string{mockQuery})
	result, err := tool.Run(ctx, input)
	if err != nil {
		t.Fatalf("Error running SearxngSearch: %v", err)
		return
	}
	if len(result.Results) != 1 {
		t.Fatalf("Error number of results, expect 1, bug got %d", len(result.Results))
		return
	}
	item := result.Results[0]
	if item.Title != mockItem.Title {
		t.Errorf("Expect title %s, but got %s", mockItem.Title, item.Title)
	}
	if item.URL != mockItem.URL {
		t.Errorf("Expect url %s, but got %s", mockItem.URL, item.URL)
	}
	if item.Content != mockItem.Content {
		t.Errorf("Expect content %s, but got %s", mockItem.Content, item.Content)
	}
	if item.Category != mockItem.Category {
		t.Errorf("Expect category %s, but got %s", mockItem.Category, item.Category)
	}
}

func TestSearxngSearchMissingFields(t *testing.T) {
	mockPort := 8080
	mockSearchURL := fmt.Sprintf("http://localhost:%d", mockPort)
	mockQuery := "query with missing fields"
	mockResult := Output{
		Results: []SearchResultItem{
			{Title: "Result Missing Content", URL: "https://example.com/1", Query: mockQuery},
			{Content: "Result Missing Title", URL: "https://example.com/2", Query: mockQuery},
			{Title: "Result Missing URL", Content: "Some content", Query: mockQuery},
			{Title: "Result Missing Query", Content: "Some content", URL: "https://example.com/4"},
			{Title: "Valid Result", Content: "Some content", URL: "https://example.com/5", Query: mockQuery},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := startSearxngServer(t, mockPort, &mockResult)
	defer srv.Shutdown(ctx)
	// Initialize the tool
	tool := New(WithBaseURL(mockSearchURL))
	input := NewInput(EmptyCategory, []string{mockQuery})
	result, err := tool.Run(ctx, input)
	if err != nil {
		t.Fatalf("Error running SearxngSearch: %v", err)
		return
	}
	if len(result.Results) != 2 {
		t.Fatalf("Error number of results, expect 2, bug got %d", len(result.Results))
		return
	}
	if title := result.Results[0].Title; title != "Result Missing Query" {
		t.Errorf("Expect title Result Missing Query, but got %s", title)
	}
	if title := result.Results[1].Title; title != "Valid Result" {
		t.Errorf("Expect title Vaid Result, but got %s", title)
	}
}

func TestSearxngSearchWithMetadataAndPublishedDate(t *testing.T) {
	mockPort := 8080
	mockSearchURL := fmt.Sprintf("http://localhost:%d", mockPort)
	mockQuery := "query with dates"
	mockResult := Output{
		Results: []SearchResultItem{
			{Title: "Result with Metadata", URL: "https://example.com/metadata", Content: "Content with metadata", Query: mockQuery, Metadata: "2021-01-01"},
			{Title: "Result with Published Date", Content: "Content with published date", URL: "https://example.com/published-data", Query: mockQuery, PublishedDate: "2022-01-01"},
			{Title: "Result without dates", Content: "Content without dates", URL: "https://example.com/no-dates", Query: mockQuery},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := startSearxngServer(t, mockPort, &mockResult)
	defer srv.Shutdown(ctx)
	// Initialize the tool
	tool := New(WithBaseURL(mockSearchURL))
	input := NewInput(EmptyCategory, []string{mockQuery})
	result, err := tool.Run(ctx, input)
	if err != nil {
		t.Fatalf("Error running SearxngSearch: %v", err)
		return
	}
	if len(result.Results) != 3 {
		t.Fatalf("Error number of results, expect 3, bug got %d", len(result.Results))
		return
	}
	for idx, got := range result.Results {
		expect := mockResult.Results[idx]
		if got.Title != expect.Title {
			t.Errorf("Expect title %s, but got %s", expect.Title, got.Title)
		}
		if got.Metadata != expect.Metadata {
			t.Errorf("Expect metadata %s, but got %s", expect.Metadata, got.Metadata)
		}
		if got.PublishedDate != expect.PublishedDate {
			t.Errorf("Expect metadata %s, but got %s", expect.PublishedDate, got.PublishedDate)
		}
	}
}

func TestSearxngSearchWithMaxResults(t *testing.T) {
	mockPort := 8080
	mockSearchURL := fmt.Sprintf("http://localhost:%d", mockPort)
	mockQuery := "query with max results"
	mockResult := Output{
		Results: []SearchResultItem{
			{Title: "Result with Metadata", URL: "https://example.com/metadata", Content: "Content with metadata", Query: mockQuery, Metadata: "2021-01-01"},
			{Title: "Result with Published Date", Content: "Content with published date", URL: "https://example.com/published-data", Query: mockQuery, PublishedDate: "2022-01-01"},
			{Title: "Result without dates", Content: "Content without dates", URL: "https://example.com/no-dates", Query: mockQuery},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := startSearxngServer(t, mockPort, &mockResult)
	defer srv.Shutdown(ctx)
	// Initialize the tool
	tool := New(WithBaseURL(mockSearchURL), WithMaxResults(2))
	input := NewInput(EmptyCategory, []string{mockQuery})
	result, err := tool.Run(ctx, input)
	if err != nil {
		t.Fatalf("Error running SearxngSearch: %v", err)
		return
	}
	if len(result.Results) != 2 {
		t.Errorf("Error number of results, expect 2, bug got %d", len(result.Results))
	}
}

func TestSearxngSearchWithNoResults(t *testing.T) {
	mockPort := 8080
	mockSearchURL := fmt.Sprintf("http://localhost:%d", mockPort)
	mockQuery := "query with max results"
	mockResult := Output{
		Results: []SearchResultItem{},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := startSearxngServer(t, mockPort, &mockResult)
	defer srv.Shutdown(ctx)
	// Initialize the tool
	tool := New(WithBaseURL(mockSearchURL), WithMaxResults(2))
	input := NewInput(EmptyCategory, []string{mockQuery})
	result, err := tool.Run(ctx, input)
	if err != nil {
		t.Fatalf("Error running SearxngSearch: %v", err)
		return
	}
	if len(result.Results) != 0 {
		t.Errorf("Error number of results, expect 0, bug got %d", len(result.Results))
	}
}
