# Changelog

## v0.2.0

### Added
- **`bulk_edit_tags` tool**: add and/or remove tags across multiple documents in a single request, using Paperless-NGX's `modify_tags` bulk-edit endpoint.

## v0.1.1

### Fixed
- **JSON-RPC notification handling**: The server no longer responds to notifications (JSON-RPC 2.0 requests without an `id`). Previously it returned a "Method not found" error for `notifications/initialized`, which caused strict MCP clients (including current Claude Code) to abort the handshake and expose zero tools.

## v0.1.0

Initial release.

### Features

- 49 MCP tools for managing a Paperless-NGX instance
- Full CRUD for correspondents, document types, tags, storage paths, custom
  fields, and saved views
- Document search, view, update, delete, upload, and download
- Document metadata and AI-generated suggestions
- Document notes (list, create, delete)
- Task listing and viewing
- Trash listing (read-only)
- System status, configuration (read and update), and statistics
- Zero external production dependencies (Go stdlib only)
