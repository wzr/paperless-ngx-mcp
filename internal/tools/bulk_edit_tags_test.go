package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chrisallenlane/paperless-ngx-mcp/internal/client"
)

const bulkEditTagsOKResponse = `{"result": "OK"}`

func TestBulkEditTags_Execute(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/documents/bulk_edit/" {
					t.Errorf(
						"Expected /api/documents/bulk_edit/, got %s",
						r.URL.Path,
					)
				}
				if r.Method != "POST" {
					t.Errorf("Expected POST, got %s", r.Method)
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("Failed to read body: %v", err)
				}

				var req map[string]interface{}
				if err := json.Unmarshal(body, &req); err != nil {
					t.Fatalf("Failed to parse body: %v", err)
				}

				if req["method"] != "modify_tags" {
					t.Errorf(
						"method = %v, want modify_tags",
						req["method"],
					)
				}

				documents, ok := req["documents"].([]interface{})
				if !ok || len(documents) != 2 {
					t.Errorf(
						"documents = %v, want [1, 2]",
						req["documents"],
					)
				}

				params, ok := req["parameters"].(map[string]interface{})
				if !ok {
					t.Fatalf(
						"parameters missing or wrong type: %v",
						req["parameters"],
					)
				}

				addTags, ok := params["add_tags"].([]interface{})
				if !ok || len(addTags) != 1 || addTags[0] != float64(5) {
					t.Errorf(
						"add_tags = %v, want [5]",
						params["add_tags"],
					)
				}

				removeTags, ok := params["remove_tags"].([]interface{})
				if !ok || len(removeTags) != 1 || removeTags[0] != float64(9) {
					t.Errorf(
						"remove_tags = %v, want [9]",
						params["remove_tags"],
					)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(bulkEditTagsOKResponse))
			},
		),
	)
	defer server.Close()

	c := client.NewWithHTTPClient(
		server.URL,
		"test-token",
		server.Client(),
	)
	tool := NewBulkEditTags(c)

	result, err := tool.Execute(
		context.Background(),
		json.RawMessage(
			`{"document_ids": [1, 2], "add_tags": [5], "remove_tags": [9]}`,
		),
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	checks := []string{
		"Bulk tag edit applied to 2 document(s)",
		"Added tags: 5",
		"Removed tags: 9",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf(
				"Output missing %q.\nGot:\n%s",
				check,
				result,
			)
		}
	}
}

func TestBulkEditTags_AddOnly(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(bulkEditTagsOKResponse))
			},
		),
	)
	defer server.Close()

	c := client.NewWithHTTPClient(
		server.URL,
		"test-token",
		server.Client(),
	)
	tool := NewBulkEditTags(c)

	result, err := tool.Execute(
		context.Background(),
		json.RawMessage(`{"document_ids": [1], "add_tags": [5]}`),
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "Removed tags: (none)") {
		t.Errorf(
			"Expected 'Removed tags: none' in output, got:\n%s",
			result,
		)
	}
}

func TestBulkEditTags_EmptyDocumentIDs(t *testing.T) {
	c := client.New("http://localhost", "test-token")
	tool := NewBulkEditTags(c)

	_, err := tool.Execute(
		context.Background(),
		json.RawMessage(`{"document_ids": [], "add_tags": [5]}`),
	)
	if err == nil {
		t.Fatal("Expected error for empty document_ids")
	}

	if !strings.Contains(err.Error(), "document_ids must contain") {
		t.Errorf(
			"Error should mention document_ids, got: %s",
			err.Error(),
		)
	}
}

func TestBulkEditTags_NoTagChanges(t *testing.T) {
	c := client.New("http://localhost", "test-token")
	tool := NewBulkEditTags(c)

	_, err := tool.Execute(
		context.Background(),
		json.RawMessage(`{"document_ids": [1, 2]}`),
	)
	if err == nil {
		t.Fatal("Expected error when neither add_tags nor remove_tags is set")
	}

	if !strings.Contains(
		err.Error(),
		"at least one of add_tags or remove_tags is required",
	) {
		t.Errorf(
			"Error should mention add_tags/remove_tags, got: %s",
			err.Error(),
		)
	}
}

func TestBulkEditTags_HTTPError(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
		),
	)
	defer server.Close()

	c := client.NewWithHTTPClient(
		server.URL,
		"test-token",
		server.Client(),
	)
	tool := NewBulkEditTags(c)

	_, err := tool.Execute(
		context.Background(),
		json.RawMessage(`{"document_ids": [1], "add_tags": [5]}`),
	)
	if err == nil {
		t.Fatal("Expected error for HTTP 500 response")
	}

	if !strings.Contains(err.Error(), "failed to bulk edit tags") {
		t.Errorf(
			"Error should mention failed to bulk edit tags, got: %s",
			err.Error(),
		)
	}
}

func TestBulkEditTags_MalformedJSON(t *testing.T) {
	c := client.New("http://localhost", "test-token")
	tool := NewBulkEditTags(c)

	_, err := tool.Execute(
		context.Background(),
		json.RawMessage("not json"),
	)
	if err == nil {
		t.Fatal("Expected error for malformed JSON input")
	}

	if !strings.Contains(err.Error(), "failed to parse arguments") {
		t.Errorf(
			"Error should mention failed to parse arguments, got: %s",
			err.Error(),
		)
	}
}
