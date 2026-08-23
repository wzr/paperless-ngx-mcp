package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/chrisallenlane/paperless-ngx-mcp/internal/client"
)

// BulkEditTags adds and/or removes tags across multiple documents in a
// single request.
type BulkEditTags struct {
	client *client.Client
}

// NewBulkEditTags creates a new BulkEditTags tool instance.
func NewBulkEditTags(c *client.Client) *BulkEditTags {
	return &BulkEditTags{client: c}
}

// Description returns a description of what this tool does.
func (t *BulkEditTags) Description() string {
	return "Add and/or remove tags across multiple documents " +
		"in Paperless-NGX in a single bulk request"
}

// InputSchema returns the JSON schema for the tool's input parameters.
func (t *BulkEditTags) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"document_ids": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "integer",
				},
				"description": "IDs of the documents to modify",
			},
			"add_tags": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "integer",
				},
				"description": "Tag IDs to add to every " +
					"listed document",
			},
			"remove_tags": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "integer",
				},
				"description": "Tag IDs to remove from every " +
					"listed document",
			},
		},
		"required": []string{"document_ids"},
	}
}

// bulkEditTagsRequest is the body sent to Paperless-NGX's bulk_edit
// endpoint for the "modify_tags" method.
type bulkEditTagsRequest struct {
	Documents  []int              `json:"documents"`
	Method     string             `json:"method"`
	Parameters bulkEditTagsParams `json:"parameters"`
}

type bulkEditTagsParams struct {
	AddTags    []int `json:"add_tags"`
	RemoveTags []int `json:"remove_tags"`
}

// bulkEditTagsResponse is Paperless-NGX's response body on success.
type bulkEditTagsResponse struct {
	Result string `json:"result"`
}

// Execute runs the tool and returns a confirmation message.
func (t *BulkEditTags) Execute(
	ctx context.Context,
	args json.RawMessage,
) (string, error) {
	var params struct {
		DocumentIDs []int `json:"document_ids"`
		AddTags     []int `json:"add_tags"`
		RemoveTags  []int `json:"remove_tags"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf(
			"failed to parse arguments: %w",
			err,
		)
	}

	if len(params.DocumentIDs) == 0 {
		return "", fmt.Errorf(
			"document_ids must contain at least one document ID",
		)
	}

	if len(params.AddTags) == 0 && len(params.RemoveTags) == 0 {
		return "", fmt.Errorf(
			"at least one of add_tags or remove_tags is required",
		)
	}

	body := bulkEditTagsRequest{
		Documents: params.DocumentIDs,
		Method:    "modify_tags",
		Parameters: bulkEditTagsParams{
			AddTags:    params.AddTags,
			RemoveTags: params.RemoveTags,
		},
	}

	resp, err := t.client.Post(ctx, "/api/documents/bulk_edit/", body)
	if err != nil {
		return "", fmt.Errorf(
			"failed to bulk edit tags: %w",
			err,
		)
	}

	respBody, err := readResponse(resp, http.StatusOK)
	if err != nil {
		return "", fmt.Errorf(
			"failed to bulk edit tags: %w",
			err,
		)
	}

	var result bulkEditTagsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf(
			"failed to parse response: %w",
			err,
		)
	}

	return fmt.Sprintf(
		"Bulk tag edit applied to %d document(s).\n"+
			"Added tags: %s\n"+
			"Removed tags: %s",
		len(params.DocumentIDs),
		formatIntSlice(params.AddTags),
		formatIntSlice(params.RemoveTags),
	), nil
}
