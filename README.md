# MCPHub

MCPHub is a command-line tool for building and managing Model Context Protocol (MCP) servers with Docker.

## Features

- 🚀 **Initialize** MCP server configurations
- 📦 **Build** Docker images from MCP server zip files
- 🔄 **Load** Docker images from tar files
- ▶️ **Run** Docker containers with custom configurations

## Installation

### Prerequisites

- Docker installed and running
- Go 1.22+ (for building from source)

### Build from Source

```bash
git clone <repository-url>
cd MCPHub
go build -o mcphub cmd/main.go
```

## Usage

### Initialize a new MCP configuration

```bash
mcphub init [--yes]
```

Creates a new `mcp.json` configuration file. Use `--yes` to skip prompts and use defaults.

### Build Docker image from zip file

```bash
mcphub push <zip-file>
```

Extracts the zip file, reads the MCP configuration, and builds a Docker image.

### Load Docker image from tar file

```bash
mcphub pull <tar-file>
```

Loads a Docker image from a tar file.

### Run Docker container

```bash
mcphub run <image-name> [flags]
```

**Flags:**

- `--detach, -d`: Run container in detached mode (default: true)
- `--port, -p`: Port mapping (e.g., 8080:8080)
- `--name, -n`: Container name (defaults to image name)

## MCP Configuration

The `mcp.json` file structure:

```json
{
  "name": "my-mcp-server",
  "version": "1.0.0",
  "description": "My MCP server",
  "author": "Author Name",
  "license": "MIT",
  "keywords": ["mcp", "server"],
  "repository": "https://github.com/user/repo",
  "run": {
    "command": "node",
    "args": ["server.js"],
    "port": 3000
  }
}
```

## Examples

1. **Create a new MCP server configuration:**

   ```bash
   mcphub init --yes
   ```

2. **Build and run a server:**

   ```bash
   mcphub push my-server.zip
   mcphub run my-mcp-server --port 3000:3000
   ```

3. **Run in interactive mode:**
   ```bash
   mcphub run my-server --detach=false
   ```
