# paperless-ngx-mcp

An [MCP server][mcp] for [Paperless-NGX][pngx], written in Go. It enables
Claude and other AI assistants to query and manage your Paperless-NGX document
management system.

## Requirements

- Go 1.24+ (to build from source)
- A running Paperless-NGX instance
- A Paperless-NGX API token ([how to get one][pngx-api-token])

## Installation

```bash
git clone https://github.com/chrisallenlane/paperless-ngx-mcp.git
cd paperless-ngx-mcp
make build
```

This produces a binary at `dist/paperless-ngx-mcp`.

Optionally, install to `$GOPATH/bin`:

```bash
make install
```

## Configuration

### Claude Code

```bash
claude mcp add paperless-ngx /path/to/paperless-ngx-mcp \
  -s user \
  -e PAPERLESS_URL=https://paperless.example.com \
  -e PAPERLESS_TOKEN=your-api-token
```

Scope options:

- `-s user` -- Available in all projects (recommended)
- `-s local` -- Available only in the current project
- `-s project` -- Saved to `.mcp.json` for team sharing

Verify the server is registered:

```bash
claude mcp list
```

### Claude Desktop

Add the following to your Claude Desktop configuration file:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "paperless-ngx": {
      "command": "/path/to/paperless-ngx-mcp",
      "env": {
        "PAPERLESS_URL": "https://paperless.example.com",
        "PAPERLESS_TOKEN": "your-api-token"
      }
    }
  }
}
```

Restart Claude Desktop after updating the configuration.

### Other MCP Clients

The server communicates via JSON-RPC 2.0 over stdio. Any MCP-compatible client
can use it by running the binary with `PAPERLESS_URL` and `PAPERLESS_TOKEN`
set in the environment.

## Capabilities

The server provides 50 tools across the following areas:

| Area | Operations |
|------|------------|
| **Documents** | Search, view, update, delete, upload, download, metadata, AI suggestions, next ASN, bulk tag edit |
| **Correspondents** | List, view, create, update, delete |
| **Document Types** | List, view, create, update, delete |
| **Tags** | List, view, create, update, delete |
| **Storage Paths** | List, view, create, update, delete |
| **Custom Fields** | List, view, create, update, delete |
| **Saved Views** | List, view, create, update, delete |
| **Document Notes** | List, create, delete |
| **Tasks** | List, view (read-only) |
| **System** | Status, configuration, statistics |
| **Trash** | List (read-only) |

Once configured, you can ask Claude things like:

- "Search my documents for tax receipts from 2024"
- "Upload `/home/me/scan.pdf` to Paperless"
- "Create a tag called 'Medical' with color #ff0000"
- "What correspondents do I have?"
- "Show me the status of my Paperless instance"
- "Add a note to document 42"

## Troubleshooting

See [SETUP.md](SETUP.md).

## Development

```bash
make build     # Build the binary
make test      # Run tests
make check     # Format, lint, vet, and test
make fuzz      # Run fuzz tests
make coverage  # Tests with coverage report
```

Zero external dependencies -- production code uses only the Go standard library.

## Resources

- [MCP Specification][mcp]
- [Paperless-NGX Documentation][pngx]
- [Paperless-NGX API][pngx-api]

[mcp]: https://modelcontextprotocol.io/
[pngx]: https://docs.paperless-ngx.com/
[pngx-api]: https://docs.paperless-ngx.com/api/
[pngx-api-token]: https://docs.paperless-ngx.com/api/#authorization
