package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chrisallenlane/paperless-ngx-mcp/internal/client"
)

// toolTestEntry defines a tool for shared cross-cutting tests.
type toolTestEntry struct {
	name           string
	newTool        func(*client.Client) Tool
	serverArgs     string   // JSON args for ServerError test; "" to skip
	idArgsFmt      string   // fmt template for InvalidID/NegativeID; "" to skip
	missingIDArgs  string   // JSON args that omit id, for MissingID test; "" to skip
	required       []string // expected required schema fields; nil to skip
	expectedStatus int      // expected HTTP status for MalformedResponse test; 0 to skip
}

// allToolTests covers every tool.
var allToolTests = []toolTestEntry{
	// No-arg tools
	{
		name: "GetStatus",
		newTool: func(c *client.Client) Tool {
			return NewGetStatus(c)
		},
		serverArgs:     `{}`,
		expectedStatus: http.StatusOK,
	},
	{
		name: "GetConfig",
		newTool: func(c *client.Client) Tool {
			return NewGetConfig(c)
		},
		serverArgs:     `{}`,
		expectedStatus: http.StatusOK,
	},
	{
		name: "GetNextASN",
		newTool: func(c *client.Client) Tool {
			return NewGetNextASN(c)
		},
		serverArgs:     `{}`,
		expectedStatus: http.StatusOK,
	},
	{
		name: "GetStatistics",
		newTool: func(c *client.Client) Tool {
			return NewGetStatistics(c)
		},
		serverArgs:     `{}`,
		expectedStatus: http.StatusOK,
	},

	// Get (ID-based) tools
	{
		name: "GetSavedView",
		newTool: func(c *client.Client) Tool {
			return NewGetSavedView(c)
		},
		serverArgs:     `{"id": 1}`,
		idArgsFmt:      `{"id": %d}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "GetTag",
		newTool: func(c *client.Client) Tool {
			return NewGetTag(c)
		},
		serverArgs:     `{"id": 1}`,
		idArgsFmt:      `{"id": %d}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "GetStoragePath",
		newTool: func(c *client.Client) Tool {
			return NewGetStoragePath(c)
		},
		serverArgs:     `{"id": 1}`,
		idArgsFmt:      `{"id": %d}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "GetTask",
		newTool: func(c *client.Client) Tool {
			return NewGetTask(c)
		},
		serverArgs:     `{"id": 1}`,
		idArgsFmt:      `{"id": %d}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "GetCorrespondent",
		newTool: func(c *client.Client) Tool {
			return NewGetCorrespondent(c)
		},
		serverArgs:     `{"id": 1}`,
		idArgsFmt:      `{"id": %d}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "GetCustomField",
		newTool: func(c *client.Client) Tool {
			return NewGetCustomField(c)
		},
		serverArgs:     `{"id": 1}`,
		idArgsFmt:      `{"id": %d}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "GetDocumentType",
		newTool: func(c *client.Client) Tool {
			return NewGetDocumentType(c)
		},
		serverArgs:     `{"id": 1}`,
		idArgsFmt:      `{"id": %d}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "GetDocument",
		newTool: func(c *client.Client) Tool {
			return NewGetDocument(c)
		},
		serverArgs:     `{"id": 1}`,
		idArgsFmt:      `{"id": %d}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "GetDocumentMetadata",
		newTool: func(c *client.Client) Tool {
			return NewGetDocumentMetadata(c)
		},
		serverArgs:     `{"id": 1}`,
		idArgsFmt:      `{"id": %d}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "GetDocumentSuggestions",
		newTool: func(c *client.Client) Tool {
			return NewGetDocumentSuggestions(c)
		},
		serverArgs:     `{"id": 1}`,
		idArgsFmt:      `{"id": %d}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},

	// List tools
	{
		name: "ListCorrespondents",
		newTool: func(c *client.Client) Tool {
			return NewListCorrespondents(c)
		},
		serverArgs:     `{}`,
		expectedStatus: http.StatusOK,
	},
	{
		name: "ListCustomFields",
		newTool: func(c *client.Client) Tool {
			return NewListCustomFields(c)
		},
		serverArgs:     `{}`,
		expectedStatus: http.StatusOK,
	},
	{
		name: "ListDocumentTypes",
		newTool: func(c *client.Client) Tool {
			return NewListDocumentTypes(c)
		},
		serverArgs:     `{}`,
		expectedStatus: http.StatusOK,
	},
	{
		name: "ListSavedViews",
		newTool: func(c *client.Client) Tool {
			return NewListSavedViews(c)
		},
		serverArgs:     `{}`,
		expectedStatus: http.StatusOK,
	},
	{
		name: "ListTags",
		newTool: func(c *client.Client) Tool {
			return NewListTags(c)
		},
		serverArgs:     `{}`,
		expectedStatus: http.StatusOK,
	},
	{
		name: "ListStoragePaths",
		newTool: func(c *client.Client) Tool {
			return NewListStoragePaths(c)
		},
		serverArgs:     `{}`,
		expectedStatus: http.StatusOK,
	},
	{
		name: "ListDocuments",
		newTool: func(c *client.Client) Tool {
			return NewListDocuments(c)
		},
		serverArgs:     `{}`,
		expectedStatus: http.StatusOK,
	},
	{
		name: "ListTrash",
		newTool: func(c *client.Client) Tool {
			return NewListTrash(c)
		},
		serverArgs:     `{}`,
		expectedStatus: http.StatusOK,
	},
	{
		name: "ListTasks",
		newTool: func(c *client.Client) Tool {
			return NewListTasks(c)
		},
		serverArgs:     `{}`,
		expectedStatus: http.StatusOK,
	},

	// Create tools
	{
		name: "CreateCorrespondent",
		newTool: func(c *client.Client) Tool {
			return NewCreateCorrespondent(c)
		},
		serverArgs:     `{"name": "Test"}`,
		required:       []string{"name"},
		expectedStatus: http.StatusCreated,
	},
	{
		name: "CreateCustomField",
		newTool: func(c *client.Client) Tool {
			return NewCreateCustomField(c)
		},
		serverArgs:     `{"name": "Test", "data_type": "string"}`,
		required:       []string{"name", "data_type"},
		expectedStatus: http.StatusCreated,
	},
	{
		name: "CreateDocumentType",
		newTool: func(c *client.Client) Tool {
			return NewCreateDocumentType(c)
		},
		serverArgs:     `{"name": "Test"}`,
		required:       []string{"name"},
		expectedStatus: http.StatusCreated,
	},
	{
		name: "CreateSavedView",
		newTool: func(c *client.Client) Tool {
			return NewCreateSavedView(c)
		},
		serverArgs: `{"name": "Test", "show_on_dashboard": true, "show_in_sidebar": false, "filter_rules": []}`,
		required: []string{
			"name",
			"show_on_dashboard",
			"show_in_sidebar",
			"filter_rules",
		},
		expectedStatus: http.StatusCreated,
	},
	{
		name: "CreateTag",
		newTool: func(c *client.Client) Tool {
			return NewCreateTag(c)
		},
		serverArgs:     `{"name": "Test"}`,
		required:       []string{"name"},
		expectedStatus: http.StatusCreated,
	},
	{
		name: "CreateStoragePath",
		newTool: func(c *client.Client) Tool {
			return NewCreateStoragePath(c)
		},
		serverArgs:     `{"name": "Test", "path": "/test/"}`,
		required:       []string{"name", "path"},
		expectedStatus: http.StatusCreated,
	},

	// Update tools
	{
		name: "UpdateCorrespondent",
		newTool: func(c *client.Client) Tool {
			return NewUpdateCorrespondent(c)
		},
		serverArgs:     `{"id": 1, "name": "Test"}`,
		idArgsFmt:      `{"id": %d}`,
		missingIDArgs:  `{"name": "Test"}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "UpdateCustomField",
		newTool: func(c *client.Client) Tool {
			return NewUpdateCustomField(c)
		},
		serverArgs:     `{"id": 1, "name": "Test"}`,
		idArgsFmt:      `{"id": %d}`,
		missingIDArgs:  `{"name": "Test"}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "UpdateDocumentType",
		newTool: func(c *client.Client) Tool {
			return NewUpdateDocumentType(c)
		},
		serverArgs:     `{"id": 1, "name": "Test"}`,
		idArgsFmt:      `{"id": %d}`,
		missingIDArgs:  `{"name": "Test"}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "UpdateSavedView",
		newTool: func(c *client.Client) Tool {
			return NewUpdateSavedView(c)
		},
		serverArgs:     `{"id": 1, "name": "Test"}`,
		idArgsFmt:      `{"id": %d}`,
		missingIDArgs:  `{"name": "Test"}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "UpdateTag",
		newTool: func(c *client.Client) Tool {
			return NewUpdateTag(c)
		},
		serverArgs:     `{"id": 1, "name": "Test"}`,
		idArgsFmt:      `{"id": %d}`,
		missingIDArgs:  `{"name": "Test"}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "UpdateStoragePath",
		newTool: func(c *client.Client) Tool {
			return NewUpdateStoragePath(c)
		},
		serverArgs:     `{"id": 1, "name": "Test"}`,
		idArgsFmt:      `{"id": %d}`,
		missingIDArgs:  `{"name": "Test"}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "UpdateDocument",
		newTool: func(c *client.Client) Tool {
			return NewUpdateDocument(c)
		},
		serverArgs:     `{"id": 1, "title": "Test"}`,
		idArgsFmt:      `{"id": %d}`,
		missingIDArgs:  `{"title": "Test"}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "UpdateConfig",
		newTool: func(c *client.Client) Tool {
			return NewUpdateConfig(c)
		},
		serverArgs:     `{"id": 1, "deskew": true}`,
		idArgsFmt:      `{"id": %d}`,
		missingIDArgs:  `{"output_type": "pdfa"}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},

	// Delete tools — expectedStatus is 0 (omitted) because deleteTool
	// expects HTTP 204 with an empty body and performs no JSON
	// unmarshaling, so there is nothing to exercise with a malformed
	// response body.
	{
		name: "DeleteCorrespondent",
		newTool: func(c *client.Client) Tool {
			return NewDeleteCorrespondent(c)
		},
		serverArgs: `{"id": 1}`,
		idArgsFmt:  `{"id": %d}`,
		required:   []string{"id"},
	},
	{
		name: "DeleteCustomField",
		newTool: func(c *client.Client) Tool {
			return NewDeleteCustomField(c)
		},
		serverArgs: `{"id": 1}`,
		idArgsFmt:  `{"id": %d}`,
		required:   []string{"id"},
	},
	{
		name: "DeleteDocumentType",
		newTool: func(c *client.Client) Tool {
			return NewDeleteDocumentType(c)
		},
		serverArgs: `{"id": 1}`,
		idArgsFmt:  `{"id": %d}`,
		required:   []string{"id"},
	},
	{
		name: "DeleteSavedView",
		newTool: func(c *client.Client) Tool {
			return NewDeleteSavedView(c)
		},
		serverArgs: `{"id": 1}`,
		idArgsFmt:  `{"id": %d}`,
		required:   []string{"id"},
	},
	{
		name: "DeleteTag",
		newTool: func(c *client.Client) Tool {
			return NewDeleteTag(c)
		},
		serverArgs: `{"id": 1}`,
		idArgsFmt:  `{"id": %d}`,
		required:   []string{"id"},
	},
	{
		name: "DeleteStoragePath",
		newTool: func(c *client.Client) Tool {
			return NewDeleteStoragePath(c)
		},
		serverArgs: `{"id": 1}`,
		idArgsFmt:  `{"id": %d}`,
		required:   []string{"id"},
	},
	{
		name: "DeleteDocument",
		newTool: func(c *client.Client) Tool {
			return NewDeleteDocument(c)
		},
		serverArgs: `{"id": 1}`,
		idArgsFmt:  `{"id": %d}`,
		required:   []string{"id"},
	},

	// Document note tools
	{
		name: "ListDocumentNotes",
		newTool: func(c *client.Client) Tool {
			return NewListDocumentNotes(c)
		},
		serverArgs:     `{"id": 1}`,
		idArgsFmt:      `{"id": %d}`,
		required:       []string{"id"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "CreateDocumentNote",
		newTool: func(c *client.Client) Tool {
			return NewCreateDocumentNote(c)
		},
		serverArgs:     `{"id": 1, "note": "Test"}`,
		idArgsFmt:      `{"id": %d, "note": "Test"}`,
		required:       []string{"id", "note"},
		expectedStatus: http.StatusOK,
	},
	{
		name: "DeleteDocumentNote",
		newTool: func(c *client.Client) Tool {
			return NewDeleteDocumentNote(c)
		},
		serverArgs: `{"document_id": 1, "note_id": 1}`,
		required: []string{
			"document_id",
			"note_id",
		},
		expectedStatus: http.StatusOK,
	},

	// Special tools (Description + InputSchema only)
	{
		name: "UploadDocument",
		newTool: func(c *client.Client) Tool {
			return NewUploadDocument(c)
		},
		required: []string{"file_path"},
	},
	{
		name: "DownloadDocument",
		newTool: func(c *client.Client) Tool {
			return NewDownloadDocument(c)
		},
		required: []string{"id", "save_path"},
	},

	// BulkEditTags takes an array of document IDs rather than a single
	// "id", so idArgsFmt/missingIDArgs (which assume a single-ID shape)
	// don't apply; its own validation is covered in
	// bulk_edit_tags_test.go instead.
	{
		name: "BulkEditTags",
		newTool: func(c *client.Client) Tool {
			return NewBulkEditTags(c)
		},
		serverArgs: `{"document_ids": [1, 2], ` +
			`"add_tags": [3]}`,
		required:       []string{"document_ids"},
		expectedStatus: http.StatusOK,
	},
}

func TestAllTools_Description(t *testing.T) {
	for _, tt := range allToolTests {
		t.Run(tt.name, func(t *testing.T) {
			c := client.New(
				"http://localhost",
				"test-token",
			)
			tool := tt.newTool(c)
			if tool.Description() == "" {
				t.Error(
					"Description should not be empty",
				)
			}
		})
	}
}

func TestAllTools_InputSchema(t *testing.T) {
	for _, tt := range allToolTests {
		t.Run(tt.name, func(t *testing.T) {
			c := client.New(
				"http://localhost",
				"test-token",
			)
			tool := tt.newTool(c)

			schema := tool.InputSchema()
			if schema == nil {
				t.Fatal(
					"InputSchema should not be nil",
				)
			}

			schemaType, ok := schema["type"].(string)
			if !ok || schemaType != "object" {
				t.Errorf(
					"Schema type = %v, want object",
					schema["type"],
				)
			}

			if tt.required == nil {
				return
			}

			required, ok := schema["required"].([]string)
			if !ok {
				t.Fatal(
					"Schema should have required field",
				)
			}

			for _, want := range tt.required {
				found := false
				for _, r := range required {
					if r == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf(
						"%s should be in required fields",
						want,
					)
				}
			}
		})
	}
}

func TestAllTools_ServerError(t *testing.T) {
	for _, tt := range allToolTests {
		if tt.serverArgs == "" {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(
						w http.ResponseWriter,
						_ *http.Request,
					) {
						w.WriteHeader(
							http.StatusInternalServerError,
						)
						w.Write(
							[]byte(
								"Internal Server Error",
							),
						)
					},
				),
			)
			defer server.Close()

			c := client.NewWithHTTPClient(
				server.URL,
				"test-token",
				server.Client(),
			)
			tool := tt.newTool(c)

			_, err := tool.Execute(
				context.Background(),
				json.RawMessage(tt.serverArgs),
			)
			if err == nil {
				t.Fatal(
					"Expected error for server error",
				)
			}

			if !strings.Contains(
				err.Error(),
				"500",
			) {
				t.Errorf(
					"Error should mention 500, got: %s",
					err.Error(),
				)
			}
		})
	}
}

func TestAllTools_InvalidAndNegativeID(t *testing.T) {
	for _, tt := range allToolTests {
		if tt.idArgsFmt == "" {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			t.Run("InvalidID", func(t *testing.T) {
				c := client.New(
					"http://localhost",
					"test-token",
				)
				tool := tt.newTool(c)

				_, err := tool.Execute(
					context.Background(),
					json.RawMessage(
						fmt.Sprintf(
							tt.idArgsFmt,
							0,
						),
					),
				)
				if err == nil {
					t.Fatal(
						"Expected error for invalid id",
					)
				}

				if !strings.Contains(
					err.Error(),
					"positive integer",
				) {
					t.Errorf(
						"Error should mention positive integer, got: %s",
						err.Error(),
					)
				}
			})

			t.Run("NegativeID", func(t *testing.T) {
				c := client.New(
					"http://localhost",
					"test-token",
				)
				tool := tt.newTool(c)

				_, err := tool.Execute(
					context.Background(),
					json.RawMessage(
						fmt.Sprintf(
							tt.idArgsFmt,
							-1,
						),
					),
				)
				if err == nil {
					t.Fatal(
						"Expected error for negative id",
					)
				}

				if !strings.Contains(
					err.Error(),
					"positive integer",
				) {
					t.Errorf(
						"Error should mention positive integer, got: %s",
						err.Error(),
					)
				}
			})
		})
	}
}

func TestAllTools_MissingID(t *testing.T) {
	for _, tt := range allToolTests {
		if tt.missingIDArgs == "" {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			c := client.New(
				"http://localhost",
				"test-token",
			)
			tool := tt.newTool(c)

			_, err := tool.Execute(
				context.Background(),
				json.RawMessage(tt.missingIDArgs),
			)
			if err == nil {
				t.Fatal("Expected error for missing id")
			}

			if !strings.Contains(
				err.Error(),
				"id is required",
			) {
				t.Errorf(
					"Error should mention id is required, got: %s",
					err.Error(),
				)
			}
		})
	}
}

// TestAllTools_MalformedResponse verifies that every tool which
// unmarshals an HTTP response body returns a non-nil error when the
// server replies with the expected success status code but an invalid
// JSON body. This exercises the "failed to parse response" branch that
// json.Unmarshal errors trigger inside fetchByID, patchByID,
// listResources, noArgGetTool.Execute, createTool.Execute, and their
// dedicated-file equivalents.
//
// Tools with expectedStatus == 0 are skipped. This currently covers:
//   - deleteTool variants (HTTP 204, no response body to unmarshal)
//   - UploadDocument (response body is a plain task-ID string, not JSON)
//   - DownloadDocument (response body is streamed to a file, not parsed)
func TestAllTools_MalformedResponse(t *testing.T) {
	for _, tt := range allToolTests {
		if tt.expectedStatus == 0 {
			continue
		}

		tt := tt // capture for parallel sub-tests
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(
						w http.ResponseWriter,
						_ *http.Request,
					) {
						w.WriteHeader(tt.expectedStatus)
						io.WriteString(
							w,
							"not valid json",
						)
					},
				),
			)
			defer server.Close()

			c := client.NewWithHTTPClient(
				server.URL,
				"test-token",
				server.Client(),
			)
			tool := tt.newTool(c)

			_, err := tool.Execute(
				context.Background(),
				json.RawMessage(tt.serverArgs),
			)
			if err == nil {
				t.Fatal(
					"Expected error for malformed response body",
				)
			}
		})
	}
}

// Response JSON constants shared across consolidated tests.

const correspondentResponse = `{
	"id": 1,
	"slug": "acme-corp",
	"name": "ACME Corp",
	"match": "acme",
	"matching_algorithm": 1,
	"is_insensitive": true,
	"document_count": 5,
	"last_correspondence": "2026-02-15T10:00:00Z"
}`

const customFieldResponse = `{
	"id": 1,
	"name": "Invoice Number",
	"data_type": "string",
	"extra_data": {"default_value": "N/A"},
	"document_count": 10
}`

const documentTypeResponse = `{
	"id": 1,
	"slug": "invoice",
	"name": "Invoice",
	"match": "invoice",
	"matching_algorithm": 1,
	"is_insensitive": true,
	"document_count": 10
}`

const documentResponse = `{
	"id": 1,
	"title": "Invoice 2024-001",
	"content": "Invoice from ACME Corp for services rendered in Q1 2024.",
	"correspondent": 1,
	"document_type": 2,
	"storage_path": 3,
	"tags": [1, 3, 5],
	"created": "2024-01-15T10:30:00Z",
	"created_date": "2024-01-15",
	"added": "2024-01-16T08:00:00Z",
	"modified": "2024-01-16T08:00:00Z",
	"archive_serial_number": 42,
	"original_file_name": "invoice-2024-001.pdf",
	"archived_file_name": "invoice-2024-001-archived.pdf",
	"mime_type": "application/pdf",
	"page_count": 2,
	"custom_fields": [
		{"field": 1, "value": "important"},
		{"field": 2, "value": 100.50}
	]
}`

const documentMetadataResponse = `{
	"original_checksum": "abc123def456",
	"original_size": 102400,
	"original_mime_type": "application/pdf",
	"media_filename": "documents/0000001.pdf",
	"original_filename": "invoice-2024.pdf",
	"original_metadata": [],
	"archive_checksum": "xyz789",
	"archive_size": 204800,
	"archive_media_filename": "documents/0000001-archive.pdf",
	"archive_metadata": [],
	"has_archive_version": true,
	"lang": "en"
}`

const documentSuggestionsResponse = `{
	"correspondents": [1, 3],
	"document_types": [2],
	"storage_paths": [1],
	"tags": [1, 5, 7],
	"dates": ["2024-01-15", "2024-02-01"]
}`

const correspondentListResponse = `{
	"count": 2,
	"next": null,
	"previous": null,
	"all": [1, 2],
	"results": [
		{
			"id": 1,
			"slug": "acme-corp",
			"name": "ACME Corp",
			"match": "acme",
			"matching_algorithm": 1,
			"is_insensitive": true,
			"document_count": 5
		},
		{
			"id": 2,
			"slug": "john-doe",
			"name": "John Doe",
			"match": "",
			"matching_algorithm": 6,
			"is_insensitive": true,
			"document_count": 0
		}
	]
}`

const customFieldListResponse = `{
	"count": 2,
	"next": null,
	"previous": null,
	"all": [1, 2],
	"results": [
		{
			"id": 1,
			"name": "Invoice Number",
			"data_type": "string",
			"extra_data": null,
			"document_count": 10
		},
		{
			"id": 2,
			"name": "Due Date",
			"data_type": "date",
			"extra_data": null,
			"document_count": 3
		}
	]
}`

const documentTypeListResponse = `{
	"count": 2,
	"next": null,
	"previous": null,
	"all": [1, 2],
	"results": [
		{
			"id": 1,
			"slug": "invoice",
			"name": "Invoice",
			"match": "invoice",
			"matching_algorithm": 1,
			"is_insensitive": true,
			"document_count": 10
		},
		{
			"id": 2,
			"slug": "receipt",
			"name": "Receipt",
			"match": "",
			"matching_algorithm": 6,
			"is_insensitive": true,
			"document_count": 3
		}
	]
}`

// GET happy-path table tests.

const savedViewResponse = `{
	"id": 1,
	"name": "Unpaid Invoices",
	"show_on_dashboard": true,
	"show_in_sidebar": true,
	"sort_field": "created",
	"sort_reverse": true,
	"filter_rules": [
		{"rule_type": 4, "value": "3"},
		{"rule_type": 6, "value": "7"}
	],
	"page_size": 50,
	"display_mode": "table"
}`

const savedViewListResponse = `{
	"count": 2,
	"next": null,
	"previous": null,
	"all": [1, 2],
	"results": [
		{
			"id": 1,
			"name": "Unpaid Invoices",
			"show_on_dashboard": true,
			"show_in_sidebar": true,
			"sort_field": "created",
			"sort_reverse": true,
			"filter_rules": [
				{"rule_type": 4, "value": "3"},
				{"rule_type": 6, "value": "7"}
			],
			"page_size": 50,
			"display_mode": "table"
		},
		{
			"id": 2,
			"name": "Recent Documents",
			"show_on_dashboard": false,
			"show_in_sidebar": true,
			"sort_field": "added",
			"sort_reverse": true,
			"filter_rules": [],
			"page_size": null,
			"display_mode": null
		}
	]
}`

const tagResponse = `{
	"id": 1,
	"slug": "important",
	"name": "Important",
	"color": "#a6cee3",
	"text_color": "#000000",
	"match": "important",
	"matching_algorithm": 1,
	"is_insensitive": true,
	"is_inbox_tag": false,
	"document_count": 8,
	"parent": null,
	"children": [3, 5]
}`

const tagListResponse = `{
	"count": 2,
	"next": null,
	"previous": null,
	"all": [1, 2],
	"results": [
		{
			"id": 1,
			"slug": "important",
			"name": "Important",
			"color": "#a6cee3",
			"text_color": "#000000",
			"match": "important",
			"matching_algorithm": 1,
			"is_insensitive": true,
			"is_inbox_tag": false,
			"document_count": 8,
			"parent": null,
			"children": [3, 5]
		},
		{
			"id": 2,
			"slug": "inbox",
			"name": "Inbox",
			"color": "#ff0000",
			"text_color": "#ffffff",
			"match": "",
			"matching_algorithm": 6,
			"is_insensitive": true,
			"is_inbox_tag": true,
			"document_count": 3,
			"parent": null,
			"children": []
		}
	]
}`

const storagePathResponse = `{
	"id": 1,
	"slug": "invoices",
	"name": "Invoices",
	"path": "{correspondent}/{document_type}/{title}",
	"match": "invoice",
	"matching_algorithm": 1,
	"is_insensitive": true,
	"document_count": 15
}`

const storagePathListResponse = `{
	"count": 2,
	"next": null,
	"previous": null,
	"all": [1, 2],
	"results": [
		{
			"id": 1,
			"slug": "invoices",
			"name": "Invoices",
			"path": "{correspondent}/{document_type}/{title}",
			"match": "invoice",
			"matching_algorithm": 1,
			"is_insensitive": true,
			"document_count": 15
		},
		{
			"id": 2,
			"slug": "archive",
			"name": "Archive",
			"path": "archive/{created_year}/{title}",
			"match": "",
			"matching_algorithm": 6,
			"is_insensitive": true,
			"document_count": 5
		}
	]
}`

const taskResponse = `{
	"id": 1,
	"task_id": "abc-123-def-456",
	"task_name": "consume_file",
	"task_file_name": "invoice.pdf",
	"date_created": "2026-02-27T10:00:00Z",
	"date_done": "2026-02-27T10:01:00Z",
	"type": "auto_task",
	"status": "SUCCESS",
	"result": "Success. New document id 42 created",
	"acknowledged": false,
	"related_document": "42"
}`

var getToolTests = []struct {
	name         string
	newTool      func(*client.Client) Tool
	path         string
	args         string
	responseJSON string
	checks       []string
}{
	{
		name: "GetSavedView",
		newTool: func(c *client.Client) Tool {
			return NewGetSavedView(c)
		},
		path:         "/api/saved_views/1/",
		args:         `{"id": 1}`,
		responseJSON: savedViewResponse,
		checks: []string{
			"Saved View (ID: 1)",
			"Name: Unpaid Invoices",
			"Show on Dashboard: true",
			"Show in Sidebar: true",
			"Sort Field: created",
			"Sort Reverse: true",
			"Page Size: 50",
			"Display Mode: table",
			"Document type is: 3",
			"Has tag: 7",
		},
	},
	{
		name: "GetTag",
		newTool: func(c *client.Client) Tool {
			return NewGetTag(c)
		},
		path:         "/api/tags/1/",
		args:         `{"id": 1}`,
		responseJSON: tagResponse,
		checks: []string{
			"Tag (ID: 1)",
			"Name: Important",
			"Slug: important",
			"Match: important",
			"Matching Algorithm: 1 (Any word)",
			"Case Insensitive: true",
			"Document Count: 8",
			"Color: #a6cee3",
			"Text Color: #000000",
			"Is Inbox Tag: false",
			"Children: 3, 5",
		},
	},
	{
		name: "GetStoragePath",
		newTool: func(c *client.Client) Tool {
			return NewGetStoragePath(c)
		},
		path:         "/api/storage_paths/1/",
		args:         `{"id": 1}`,
		responseJSON: storagePathResponse,
		checks: []string{
			"Storage Path (ID: 1)",
			"Name: Invoices",
			"Slug: invoices",
			"Path: {correspondent}/{document_type}/{title}",
			"Match: invoice",
			"Matching Algorithm: 1 (Any word)",
			"Case Insensitive: true",
			"Document Count: 15",
		},
	},
	{
		name: "GetTask",
		newTool: func(c *client.Client) Tool {
			return NewGetTask(c)
		},
		path:         "/api/tasks/1/",
		args:         `{"id": 1}`,
		responseJSON: taskResponse,
		checks: []string{
			"Task (ID: 1)",
			"Task UUID: abc-123-def-456",
			"Status: SUCCESS",
			"Type: auto_task",
			"Task Name: consume_file",
			"File Name: invoice.pdf",
			"Created: 2026-02-27",
			"Done: 2026-02-27",
			"Result: Success. New document id 42 created",
			"Acknowledged: false",
			"Related Document: 42",
		},
	},
	{
		name: "GetCorrespondent",
		newTool: func(c *client.Client) Tool {
			return NewGetCorrespondent(c)
		},
		path:         "/api/correspondents/1/",
		args:         `{"id": 1}`,
		responseJSON: correspondentResponse,
		checks: []string{
			"Correspondent (ID: 1)",
			"Name: ACME Corp",
			"Slug: acme-corp",
			"Match: acme",
			"Matching Algorithm: 1 (Any word)",
			"Case Insensitive: true",
			"Document Count: 5",
			"Last Correspondence: 2026-02-15",
		},
	},
	{
		name: "GetCustomField",
		newTool: func(c *client.Client) Tool {
			return NewGetCustomField(c)
		},
		path:         "/api/custom_fields/1/",
		args:         `{"id": 1}`,
		responseJSON: customFieldResponse,
		checks: []string{
			"Custom Field (ID: 1)",
			"Name: Invoice Number",
			"Data Type: string",
			"Extra Data:",
			"Document Count: 10",
		},
	},
	{
		name: "GetDocumentType",
		newTool: func(c *client.Client) Tool {
			return NewGetDocumentType(c)
		},
		path:         "/api/document_types/1/",
		args:         `{"id": 1}`,
		responseJSON: documentTypeResponse,
		checks: []string{
			"Document Type (ID: 1)",
			"Name: Invoice",
			"Slug: invoice",
			"Match: invoice",
			"Matching Algorithm: 1 (Any word)",
			"Case Insensitive: true",
			"Document Count: 10",
		},
	},
	{
		name: "GetDocument",
		newTool: func(c *client.Client) Tool {
			return NewGetDocument(c)
		},
		path:         "/api/documents/1/",
		args:         `{"id": 1}`,
		responseJSON: documentResponse,
		checks: []string{
			"Document (ID: 1)",
			"Title: Invoice 2024-001",
			"Correspondent: 1",
			"Document Type: 2",
			"Storage Path: 3",
			"Tags: 1, 3, 5",
			"ASN: 42",
			"Original File: invoice-2024-001.pdf",
			"MIME Type: application/pdf",
			"Page Count: 2",
			"Custom Fields:",
			"Field 1:",
			"Content: Invoice from ACME Corp",
		},
	},
	{
		name: "GetDocumentMetadata",
		newTool: func(c *client.Client) Tool {
			return NewGetDocumentMetadata(c)
		},
		path:         "/api/documents/1/metadata/",
		args:         `{"id": 1}`,
		responseJSON: documentMetadataResponse,
		checks: []string{
			"Document Metadata",
			"Filename: invoice-2024.pdf",
			"MIME Type: application/pdf",
			"Checksum: abc123def456",
			"Has Archive Version: true",
			"OCR Language: en",
			"100.00 KB",
		},
	},
	{
		name: "GetDocumentSuggestions",
		newTool: func(c *client.Client) Tool {
			return NewGetDocumentSuggestions(c)
		},
		path:         "/api/documents/1/suggestions/",
		args:         `{"id": 1}`,
		responseJSON: documentSuggestionsResponse,
		checks: []string{
			"Document Suggestions",
			"Correspondents: 1, 3",
			"Document Types: 2",
			"Storage Paths: 1",
			"Tags: 1, 5, 7",
			"Dates: 2024-01-15, 2024-02-01",
		},
	},
}

func TestGet_Execute(t *testing.T) {
	for _, tt := range getToolTests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(
						w http.ResponseWriter,
						r *http.Request,
					) {
						if r.URL.Path != tt.path {
							t.Errorf(
								"Expected %s, got %s",
								tt.path,
								r.URL.Path,
							)
						}
						if r.Method != "GET" {
							t.Errorf(
								"Expected GET, got %s",
								r.Method,
							)
						}
						w.Header().Set(
							"Content-Type",
							"application/json",
						)
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(tt.responseJSON))
					},
				),
			)
			defer server.Close()

			c := client.NewWithHTTPClient(
				server.URL,
				"test-token",
				server.Client(),
			)
			tool := tt.newTool(c)

			result, err := tool.Execute(
				context.Background(),
				json.RawMessage(tt.args),
			)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			for _, check := range tt.checks {
				if !strings.Contains(result, check) {
					t.Errorf(
						"Output missing %q.\nGot:\n%s",
						check,
						result,
					)
				}
			}
		})
	}
}

// LIST happy-path table tests.

var listToolTests = []struct {
	name            string
	newTool         func(*client.Client) Tool
	path            string
	responseJSON    string
	checks          []string
	emptyMessage    string
	nameFilterName  string
	nameFilterCheck string
}{
	{
		name: "ListCorrespondents",
		newTool: func(c *client.Client) Tool {
			return NewListCorrespondents(c)
		},
		path:         "/api/correspondents/",
		responseJSON: correspondentListResponse,
		checks: []string{
			"Correspondents: 2 total",
			"ACME Corp (ID: 1)",
			"5 documents",
			"John Doe (ID: 2)",
			"0 documents",
		},
		emptyMessage:    "No correspondents found.",
		nameFilterName:  "acme",
		nameFilterCheck: "ACME Corp",
	},
	{
		name: "ListCustomFields",
		newTool: func(c *client.Client) Tool {
			return NewListCustomFields(c)
		},
		path:         "/api/custom_fields/",
		responseJSON: customFieldListResponse,
		checks: []string{
			"Custom Fields: 2 total",
			"Invoice Number (ID: 1)",
			"type: string",
			"10 documents",
			"Due Date (ID: 2)",
			"type: date",
			"3 documents",
		},
		emptyMessage:    "No custom fields found.",
		nameFilterName:  "invoice",
		nameFilterCheck: "Invoice Number",
	},
	{
		name: "ListDocumentTypes",
		newTool: func(c *client.Client) Tool {
			return NewListDocumentTypes(c)
		},
		path:         "/api/document_types/",
		responseJSON: documentTypeListResponse,
		checks: []string{
			"Document Types: 2 total",
			"Invoice (ID: 1)",
			"10 documents",
			"Receipt (ID: 2)",
			"3 documents",
		},
		emptyMessage:    "No document types found.",
		nameFilterName:  "invoice",
		nameFilterCheck: "Invoice",
	},
	{
		name: "ListSavedViews",
		newTool: func(c *client.Client) Tool {
			return NewListSavedViews(c)
		},
		path:         "/api/saved_views/",
		responseJSON: savedViewListResponse,
		checks: []string{
			"Saved Views: 2 total",
			"Unpaid Invoices (ID: 1)",
			"2 filter rules",
			"[dashboard]",
			"[sidebar]",
			"Recent Documents (ID: 2)",
			"0 filter rules",
		},
		emptyMessage:    "No saved views found.",
		nameFilterName:  "",
		nameFilterCheck: "",
	},
	{
		name: "ListTags",
		newTool: func(c *client.Client) Tool {
			return NewListTags(c)
		},
		path:         "/api/tags/",
		responseJSON: tagListResponse,
		checks: []string{
			"Tags: 2 total",
			"Important (ID: 1)",
			"8 documents",
			"Inbox (ID: 2)",
			"3 documents",
			"[inbox]",
		},
		emptyMessage:    "No tags found.",
		nameFilterName:  "important",
		nameFilterCheck: "Important",
	},
	{
		name: "ListStoragePaths",
		newTool: func(c *client.Client) Tool {
			return NewListStoragePaths(c)
		},
		path:         "/api/storage_paths/",
		responseJSON: storagePathListResponse,
		checks: []string{
			"Storage Paths: 2 total",
			"Invoices (ID: 1)",
			"15 documents",
			"Archive (ID: 2)",
			"5 documents",
		},
		emptyMessage:    "No storage paths found.",
		nameFilterName:  "invoices",
		nameFilterCheck: "Invoices",
	},
}

func TestList_Execute(t *testing.T) {
	for _, tt := range listToolTests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(
						w http.ResponseWriter,
						r *http.Request,
					) {
						if r.URL.Path != tt.path {
							t.Errorf(
								"Expected %s, got %s",
								tt.path,
								r.URL.Path,
							)
						}
						if r.Method != "GET" {
							t.Errorf(
								"Expected GET, got %s",
								r.Method,
							)
						}
						w.Header().Set(
							"Content-Type",
							"application/json",
						)
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(tt.responseJSON))
					},
				),
			)
			defer server.Close()

			c := client.NewWithHTTPClient(
				server.URL,
				"test-token",
				server.Client(),
			)
			tool := tt.newTool(c)

			result, err := tool.Execute(
				context.Background(),
				json.RawMessage(`{}`),
			)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			for _, check := range tt.checks {
				if !strings.Contains(result, check) {
					t.Errorf(
						"Output missing %q.\nGot:\n%s",
						check,
						result,
					)
				}
			}
		})
	}
}

func TestList_Execute_NameFilter(t *testing.T) {
	for _, tt := range listToolTests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(
						w http.ResponseWriter,
						r *http.Request,
					) {
						got := r.URL.Query().Get(
							"name__icontains",
						)
						if got != tt.nameFilterName {
							t.Errorf(
								"Expected name__icontains=%s, got %s",
								tt.nameFilterName,
								r.URL.RawQuery,
							)
						}
						w.Header().Set(
							"Content-Type",
							"application/json",
						)
						w.WriteHeader(http.StatusOK)
						// Return a minimal valid list
						// containing one item that has
						// the expected check string in
						// its name.
						w.Write([]byte(tt.responseJSON))
					},
				),
			)
			defer server.Close()

			c := client.NewWithHTTPClient(
				server.URL,
				"test-token",
				server.Client(),
			)
			tool := tt.newTool(c)

			args := fmt.Sprintf(
				`{"name": "%s"}`,
				tt.nameFilterName,
			)
			result, err := tool.Execute(
				context.Background(),
				json.RawMessage(args),
			)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !strings.Contains(
				result,
				tt.nameFilterCheck,
			) {
				t.Errorf(
					"Output missing %q.\nGot:\n%s",
					tt.nameFilterCheck,
					result,
				)
			}
		})
	}
}

func TestList_Execute_Empty(t *testing.T) {
	for _, tt := range listToolTests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(
						w http.ResponseWriter,
						_ *http.Request,
					) {
						w.Header().Set(
							"Content-Type",
							"application/json",
						)
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{
							"count": 0,
							"next": null,
							"previous": null,
							"all": [],
							"results": []
						}`))
					},
				),
			)
			defer server.Close()

			c := client.NewWithHTTPClient(
				server.URL,
				"test-token",
				server.Client(),
			)
			tool := tt.newTool(c)

			result, err := tool.Execute(
				context.Background(),
				json.RawMessage(`{}`),
			)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.emptyMessage {
				t.Errorf(
					"Expected %q, got: %s",
					tt.emptyMessage,
					result,
				)
			}
		})
	}
}

// UPDATE happy-path table tests.

var updateToolTests = []struct {
	name         string
	newTool      func(*client.Client) Tool
	path         string
	args         string
	fieldName    string
	fieldValue   string
	responseJSON string
	checks       []string
}{
	{
		name: "UpdateCorrespondent",
		newTool: func(c *client.Client) Tool {
			return NewUpdateCorrespondent(c)
		},
		path:       "/api/correspondents/1/",
		args:       `{"id": 1, "name": "Updated Corp"}`,
		fieldName:  "name",
		fieldValue: "Updated Corp",
		responseJSON: `{
			"id": 1,
			"slug": "updated-corp",
			"name": "Updated Corp",
			"match": "acme",
			"matching_algorithm": 1,
			"is_insensitive": true,
			"document_count": 5,
			"last_correspondence": null
		}`,
		checks: []string{
			"Correspondent (ID: 1)",
			"Name: Updated Corp",
		},
	},
	{
		name: "UpdateCustomField",
		newTool: func(c *client.Client) Tool {
			return NewUpdateCustomField(c)
		},
		path:       "/api/custom_fields/1/",
		args:       `{"id": 1, "name": "Updated Field"}`,
		fieldName:  "name",
		fieldValue: "Updated Field",
		responseJSON: `{
			"id": 1,
			"name": "Updated Field",
			"data_type": "string",
			"extra_data": null,
			"document_count": 10
		}`,
		checks: []string{
			"Custom Field (ID: 1)",
			"Name: Updated Field",
		},
	},
	{
		name: "UpdateDocumentType",
		newTool: func(c *client.Client) Tool {
			return NewUpdateDocumentType(c)
		},
		path:       "/api/document_types/1/",
		args:       `{"id": 1, "name": "Updated Type"}`,
		fieldName:  "name",
		fieldValue: "Updated Type",
		responseJSON: `{
			"id": 1,
			"slug": "updated-type",
			"name": "Updated Type",
			"match": "invoice",
			"matching_algorithm": 1,
			"is_insensitive": true,
			"document_count": 10
		}`,
		checks: []string{
			"Document Type (ID: 1)",
			"Name: Updated Type",
		},
	},
	{
		name: "UpdateSavedView",
		newTool: func(c *client.Client) Tool {
			return NewUpdateSavedView(c)
		},
		path:       "/api/saved_views/1/",
		args:       `{"id": 1, "name": "Updated View"}`,
		fieldName:  "name",
		fieldValue: "Updated View",
		responseJSON: `{
			"id": 1,
			"name": "Updated View",
			"show_on_dashboard": true,
			"show_in_sidebar": false,
			"sort_field": "created",
			"sort_reverse": true,
			"filter_rules": [],
			"page_size": null,
			"display_mode": null
		}`,
		checks: []string{
			"Saved View (ID: 1)",
			"Name: Updated View",
		},
	},
	{
		name: "UpdateTag",
		newTool: func(c *client.Client) Tool {
			return NewUpdateTag(c)
		},
		path:       "/api/tags/1/",
		args:       `{"id": 1, "name": "Updated Tag"}`,
		fieldName:  "name",
		fieldValue: "Updated Tag",
		responseJSON: `{
			"id": 1,
			"slug": "updated-tag",
			"name": "Updated Tag",
			"color": "#a6cee3",
			"text_color": "#000000",
			"match": "important",
			"matching_algorithm": 1,
			"is_insensitive": true,
			"is_inbox_tag": false,
			"document_count": 8,
			"parent": null,
			"children": []
		}`,
		checks: []string{
			"Tag (ID: 1)",
			"Name: Updated Tag",
		},
	},
	{
		name: "UpdateStoragePath",
		newTool: func(c *client.Client) Tool {
			return NewUpdateStoragePath(c)
		},
		path:       "/api/storage_paths/1/",
		args:       `{"id": 1, "name": "Updated Path"}`,
		fieldName:  "name",
		fieldValue: "Updated Path",
		responseJSON: `{
			"id": 1,
			"slug": "updated-path",
			"name": "Updated Path",
			"path": "{correspondent}/{title}",
			"match": "invoice",
			"matching_algorithm": 1,
			"is_insensitive": true,
			"document_count": 15
		}`,
		checks: []string{
			"Storage Path (ID: 1)",
			"Name: Updated Path",
		},
	},
}

func TestUpdate_Execute(t *testing.T) {
	for _, tt := range updateToolTests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(
						w http.ResponseWriter,
						r *http.Request,
					) {
						if r.URL.Path != tt.path {
							t.Errorf(
								"Expected %s, got %s",
								tt.path,
								r.URL.Path,
							)
						}
						if r.Method != "PATCH" {
							t.Errorf(
								"Expected PATCH, got %s",
								r.Method,
							)
						}

						body, err := io.ReadAll(r.Body)
						if err != nil {
							t.Fatalf(
								"Failed to read body: %v",
								err,
							)
						}

						var patch map[string]interface{}
						if err := json.Unmarshal(
							body,
							&patch,
						); err != nil {
							t.Fatalf(
								"Failed to parse body: %v",
								err,
							)
						}

						if _, ok := patch["id"]; ok {
							t.Error(
								"Body should not contain id",
							)
						}

						if patch[tt.fieldName] != tt.fieldValue {
							t.Errorf(
								"%s = %v, want %s",
								tt.fieldName,
								patch[tt.fieldName],
								tt.fieldValue,
							)
						}

						w.Header().Set(
							"Content-Type",
							"application/json",
						)
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(tt.responseJSON))
					},
				),
			)
			defer server.Close()

			c := client.NewWithHTTPClient(
				server.URL,
				"test-token",
				server.Client(),
			)
			tool := tt.newTool(c)

			result, err := tool.Execute(
				context.Background(),
				json.RawMessage(tt.args),
			)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			for _, check := range tt.checks {
				if !strings.Contains(result, check) {
					t.Errorf(
						"Output missing %q.\nGot:\n%s",
						check,
						result,
					)
				}
			}
		})
	}
}

// CREATE name validation table tests.

var createNameValidationTests = []struct {
	name    string
	newTool func(*client.Client) Tool
}{
	{
		"Correspondent",
		func(c *client.Client) Tool {
			return NewCreateCorrespondent(c)
		},
	},
	{
		"DocumentType",
		func(c *client.Client) Tool {
			return NewCreateDocumentType(c)
		},
	},
}

func TestCreateMatchable_NameValidation(t *testing.T) {
	for _, tt := range createNameValidationTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("MissingName", func(t *testing.T) {
				c := client.New(
					"http://localhost",
					"test-token",
				)
				tool := tt.newTool(c)

				_, err := tool.Execute(
					context.Background(),
					json.RawMessage(`{}`),
				)
				if err == nil {
					t.Fatal(
						"Expected error for missing name",
					)
				}

				if !strings.Contains(
					err.Error(),
					"name is required",
				) {
					t.Errorf(
						"Error should mention name is required, got: %s",
						err.Error(),
					)
				}
			})

			t.Run("EmptyName", func(t *testing.T) {
				c := client.New(
					"http://localhost",
					"test-token",
				)
				tool := tt.newTool(c)

				_, err := tool.Execute(
					context.Background(),
					json.RawMessage(`{"name": ""}`),
				)
				if err == nil {
					t.Fatal(
						"Expected error for empty name",
					)
				}

				if !strings.Contains(
					err.Error(),
					"name is required",
				) {
					t.Errorf(
						"Error should mention name is required, got: %s",
						err.Error(),
					)
				}
			})
		})
	}
}

// CREATE happy-path table tests.

var createMatchableTests = []struct {
	name         string
	newTool      func(*client.Client) Tool
	path         string
	args         string
	fieldName    string
	fieldValue   string
	responseJSON string
	checks       []string
}{
	{
		name: "CreateCorrespondent",
		newTool: func(c *client.Client) Tool {
			return NewCreateCorrespondent(c)
		},
		path:       "/api/correspondents/",
		args:       `{"name": "ACME Corp"}`,
		fieldName:  "name",
		fieldValue: "ACME Corp",
		responseJSON: `{
			"id": 1,
			"slug": "acme-corp",
			"name": "ACME Corp",
			"match": "",
			"matching_algorithm": 1,
			"is_insensitive": true
		}`,
		checks: []string{
			"Correspondent (ID: 1)",
			"Name: ACME Corp",
		},
	},
	{
		name: "CreateDocumentType",
		newTool: func(c *client.Client) Tool {
			return NewCreateDocumentType(c)
		},
		path:       "/api/document_types/",
		args:       `{"name": "Invoice"}`,
		fieldName:  "name",
		fieldValue: "Invoice",
		responseJSON: `{
			"id": 1,
			"slug": "invoice",
			"name": "Invoice",
			"match": "",
			"matching_algorithm": 1,
			"is_insensitive": true
		}`,
		checks: []string{
			"Document Type (ID: 1)",
			"Name: Invoice",
		},
	},
}

func TestCreateMatchable_Execute(t *testing.T) {
	for _, tt := range createMatchableTests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(
						w http.ResponseWriter,
						r *http.Request,
					) {
						if r.URL.Path != tt.path {
							t.Errorf(
								"Expected %s, got %s",
								tt.path,
								r.URL.Path,
							)
						}
						if r.Method != "POST" {
							t.Errorf(
								"Expected POST, got %s",
								r.Method,
							)
						}

						body, err := io.ReadAll(r.Body)
						if err != nil {
							t.Fatalf(
								"Failed to read body: %v",
								err,
							)
						}

						var req map[string]interface{}
						if err := json.Unmarshal(
							body,
							&req,
						); err != nil {
							t.Fatalf(
								"Failed to parse body: %v",
								err,
							)
						}

						if req[tt.fieldName] != tt.fieldValue {
							t.Errorf(
								"%s = %v, want %s",
								tt.fieldName,
								req[tt.fieldName],
								tt.fieldValue,
							)
						}

						w.Header().Set(
							"Content-Type",
							"application/json",
						)
						w.WriteHeader(http.StatusCreated)
						w.Write([]byte(tt.responseJSON))
					},
				),
			)
			defer server.Close()

			c := client.NewWithHTTPClient(
				server.URL,
				"test-token",
				server.Client(),
			)
			tool := tt.newTool(c)

			result, err := tool.Execute(
				context.Background(),
				json.RawMessage(tt.args),
			)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			for _, check := range tt.checks {
				if !strings.Contains(result, check) {
					t.Errorf(
						"Output missing %q.\nGot:\n%s",
						check,
						result,
					)
				}
			}
		})
	}
}
