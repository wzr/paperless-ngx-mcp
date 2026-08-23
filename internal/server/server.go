// Package server implements the MCP (Model Context Protocol) JSON-RPC server.
package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/chrisallenlane/paperless-ngx-mcp/internal/client"
	"github.com/chrisallenlane/paperless-ngx-mcp/internal/tools"
)

// Constants for server configuration
const (
	MCPProtocolVersion   = "2024-11-05"
	ServerName           = "paperless-ngx-mcp"
	ServerVersion        = "0.2.0"
	ToolExecutionTimeout = 30 * time.Second
)

// Server represents an MCP server
type Server struct {
	client *client.Client
	tools  map[string]tools.Tool
}

// New creates a new MCP server
func New(c *client.Client) *Server {
	s := &Server{
		client: c,
		tools:  make(map[string]tools.Tool),
	}

	// Register tools
	s.registerTools()

	return s
}

// registerTools registers all available tools
func (s *Server) registerTools() {
	s.tools["get_status"] = tools.NewGetStatus(s.client)
	s.tools["get_config"] = tools.NewGetConfig(s.client)
	s.tools["update_config"] = tools.NewUpdateConfig(s.client)
	s.tools["list_correspondents"] = tools.NewListCorrespondents(
		s.client,
	)
	s.tools["get_correspondent"] = tools.NewGetCorrespondent(
		s.client,
	)
	s.tools["create_correspondent"] = tools.NewCreateCorrespondent(
		s.client,
	)
	s.tools["update_correspondent"] = tools.NewUpdateCorrespondent(
		s.client,
	)
	s.tools["delete_correspondent"] = tools.NewDeleteCorrespondent(
		s.client,
	)
	s.tools["list_custom_fields"] = tools.NewListCustomFields(
		s.client,
	)
	s.tools["get_custom_field"] = tools.NewGetCustomField(
		s.client,
	)
	s.tools["create_custom_field"] = tools.NewCreateCustomField(
		s.client,
	)
	s.tools["update_custom_field"] = tools.NewUpdateCustomField(
		s.client,
	)
	s.tools["delete_custom_field"] = tools.NewDeleteCustomField(
		s.client,
	)
	s.tools["list_document_types"] = tools.NewListDocumentTypes(
		s.client,
	)
	s.tools["get_document_type"] = tools.NewGetDocumentType(
		s.client,
	)
	s.tools["create_document_type"] = tools.NewCreateDocumentType(
		s.client,
	)
	s.tools["update_document_type"] = tools.NewUpdateDocumentType(
		s.client,
	)
	s.tools["delete_document_type"] = tools.NewDeleteDocumentType(
		s.client,
	)
	s.tools["list_documents"] = tools.NewListDocuments(
		s.client,
	)
	s.tools["get_document"] = tools.NewGetDocument(
		s.client,
	)
	s.tools["update_document"] = tools.NewUpdateDocument(
		s.client,
	)
	s.tools["delete_document"] = tools.NewDeleteDocument(
		s.client,
	)
	s.tools["get_document_metadata"] = tools.NewGetDocumentMetadata(
		s.client,
	)
	s.tools["get_document_suggestions"] = tools.NewGetDocumentSuggestions(
		s.client,
	)
	s.tools["get_next_asn"] = tools.NewGetNextASN(
		s.client,
	)
	s.tools["upload_document"] = tools.NewUploadDocument(
		s.client,
	)
	s.tools["download_document"] = tools.NewDownloadDocument(
		s.client,
	)
	s.tools["bulk_edit_tags"] = tools.NewBulkEditTags(
		s.client,
	)
	s.tools["list_tags"] = tools.NewListTags(
		s.client,
	)
	s.tools["get_tag"] = tools.NewGetTag(
		s.client,
	)
	s.tools["create_tag"] = tools.NewCreateTag(
		s.client,
	)
	s.tools["update_tag"] = tools.NewUpdateTag(
		s.client,
	)
	s.tools["delete_tag"] = tools.NewDeleteTag(
		s.client,
	)
	s.tools["list_storage_paths"] = tools.NewListStoragePaths(
		s.client,
	)
	s.tools["get_storage_path"] = tools.NewGetStoragePath(
		s.client,
	)
	s.tools["create_storage_path"] = tools.NewCreateStoragePath(
		s.client,
	)
	s.tools["update_storage_path"] = tools.NewUpdateStoragePath(
		s.client,
	)
	s.tools["delete_storage_path"] = tools.NewDeleteStoragePath(
		s.client,
	)
	s.tools["list_saved_views"] = tools.NewListSavedViews(
		s.client,
	)
	s.tools["get_saved_view"] = tools.NewGetSavedView(
		s.client,
	)
	s.tools["create_saved_view"] = tools.NewCreateSavedView(
		s.client,
	)
	s.tools["update_saved_view"] = tools.NewUpdateSavedView(
		s.client,
	)
	s.tools["delete_saved_view"] = tools.NewDeleteSavedView(
		s.client,
	)
	s.tools["list_document_notes"] = tools.NewListDocumentNotes(
		s.client,
	)
	s.tools["create_document_note"] = tools.NewCreateDocumentNote(
		s.client,
	)
	s.tools["delete_document_note"] = tools.NewDeleteDocumentNote(
		s.client,
	)
	s.tools["get_statistics"] = tools.NewGetStatistics(
		s.client,
	)
	s.tools["list_tasks"] = tools.NewListTasks(
		s.client,
	)
	s.tools["get_task"] = tools.NewGetTask(
		s.client,
	)
	s.tools["list_trash"] = tools.NewListTrash(
		s.client,
	)
}

// Run starts the MCP server and processes requests
func (s *Server) Run(
	ctx context.Context,
	stdin io.Reader,
	stdout io.Writer,
) error {
	scanner := bufio.NewScanner(stdin)
	encoder := json.NewEncoder(stdout)

	for scanner.Scan() {
		line := scanner.Bytes()

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("Failed to parse request: %v", err)
			// Send error response for malformed JSON-RPC request
			errResp := &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      nil, // ID is unknown for malformed requests
				Error: &JSONRPCError{
					Code:    -32700, // Parse error
					Message: fmt.Sprintf("Parse error: %v", err),
				},
			}
			if encErr := encoder.Encode(errResp); encErr != nil {
				log.Printf("Failed to encode error response: %v", encErr)
			}
			continue
		}

		resp := s.handleRequest(ctx, &req)
		if resp == nil {
			continue
		}
		if err := encoder.Encode(resp); err != nil {
			log.Printf("Failed to encode response: %v", err)
			return err
		}
	}

	return scanner.Err()
}

// handleRequest processes a JSON-RPC request. Returns nil for
// notifications (JSON-RPC 2.0 requests without an id), which MUST
// NOT receive a response per the spec. The MCP handshake sends
// "notifications/initialized" after initialize; responding to it
// causes strict clients to abort the handshake.
func (s *Server) handleRequest(
	ctx context.Context,
	req *JSONRPCRequest,
) *JSONRPCResponse {
	if req.ID == nil {
		return nil
	}

	resp := &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = s.handleInitialize(ctx, req.Params)
	case "tools/list":
		resp.Result = s.handleListTools(ctx)
	case "tools/call":
		result, err := s.handleCallTool(ctx, req.Params)
		if err != nil {
			resp.Error = &JSONRPCError{
				Code:    -32603,
				Message: err.Error(),
			}
		} else {
			resp.Result = result
		}
	default:
		resp.Error = &JSONRPCError{
			Code:    -32601,
			Message: fmt.Sprintf("Method not found: %s", req.Method),
		}
	}

	return resp
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(
	_ context.Context,
	_ json.RawMessage,
) interface{} {
	return map[string]interface{}{
		"protocolVersion": MCPProtocolVersion,
		"capabilities": map[string]interface{}{
			"tools": map[string]bool{},
		},
		"serverInfo": map[string]string{
			"name":    ServerName,
			"version": ServerVersion,
		},
	}
}

// handleListTools returns the list of available tools
func (s *Server) handleListTools(_ context.Context) interface{} {
	toolList := make([]map[string]interface{}, 0, len(s.tools))

	for name, tool := range s.tools {
		toolList = append(toolList, map[string]interface{}{
			"name":        name,
			"description": tool.Description(),
			"inputSchema": tool.InputSchema(),
		})
	}

	return map[string]interface{}{
		"tools": toolList,
	}
}

// handleCallTool executes a tool
func (s *Server) handleCallTool(
	ctx context.Context,
	params json.RawMessage,
) (interface{}, error) {
	var callParams struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}

	if err := json.Unmarshal(params, &callParams); err != nil {
		return nil, fmt.Errorf("failed to parse tool call params: %w", err)
	}

	tool, exists := s.tools[callParams.Name]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", callParams.Name)
	}

	// Create context with timeout for tool execution
	toolCtx, cancel := context.WithTimeout(ctx, ToolExecutionTimeout)
	defer cancel()

	result, err := tool.Execute(toolCtx, callParams.Arguments)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	}, nil
}
