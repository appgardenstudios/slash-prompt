# `/prompt`

A Model Context Protocol (MCP) server that enables sharing and discovering prompts and resources from Git repositories.

## Features

- **Share Prompts and Resources**: Use shared prompts and resources in any tool that supports MCP
- **Backed By Git**: Load prompts and resources directly from public or private Git repositories
- **Multiple Repository Support**: Support for multiple repositories with custom paths and patterns
- **Argument Substitution**: Dynamic prompt templates with variable substitution
- **Embedded Prompt Resources**: Embed resources in prompts

## Quick Start

1. Create a git repository with [prompts](#prompts) and resources.
2. Create a [configuration file](#configuration).
3. [Install](#installation) the MCP server.
4. Start [using](#usage) `/prompt` with your AI assistant.

## Prompts

Prompts are Markdown files with YAML frontmatter. Example:

```markdown
---
name: load-documentation
description: Use hyaline to load relevant documentation into context
arguments:
  - name: task
    description: A description of the task being performed
    required: true
resources:
  - path: ../resources/documentation-summary.template.md
    name: documentation-summary-template
---

Based on the following task, use the hyaline MCP server to get relevant documentation.

Task: ${task}

Use the provided documentation-summary-template to summarize the documentation that was retrieved.
```

### Frontmatter Fields

The YAML frontmatter supports the following fields:

- `name` (string, optional): Name of the prompt. Defaults to the filename without extension.
- `description` (string, optional): Description of what the prompt does.
- `arguments` (array, optional): Arguments the prompt accepts.
  - `name` (string, required): Argument name.
  - `description` (string, optional): Argument description.
  - `required` (boolean, optional): Whether the argument is required. Defaults to `false`.
- `resources` (array, optional): Resources to embed in the prompt.
  - `path` (string, required): Path to the resource file (relative to the prompt file).
  - `name` (string, optional): Resource name. Defaults to the filename.
  - `mimeType` (string, optional): MIME type. Auto-detected if not specified.

## Configuration

`/prompt` is configured by a YAML file. Example:

```yml
repos:
  - repo: https://github.com/appgardenstudios/hyaline.git
    id: hyaline
    ref: main
    prompts:
      include:
        - "contributing/prompts/**/*.md"
      exclude:
        - "contributing/prompts/resources/**/*"
    resources:
      include:
        - "contributing/prompts/resources/**/*.template.md"
```

### Configuration Options

- `repo` (string, required): Git repository URL (HTTPS only).
- `id` (string, optional): Unique identifier for the repository. Defaults to repository name.
- `ref` (string, optional): Git branch, tag, or commit to use. Defaults to the repository's default branch.
- `auth` (object, optional): Authentication configuration for private repositories.
  - `username` (string, required if auth specified): Username for authentication.
  - `password` (string, required if auth specified): Password or personal access token.
- `prompts` (object, optional): Configuration for prompt files.
  - `include` (array, optional): Glob patterns for files to include. Defaults to `["**/*.md"]`.
  - `exclude` (array, optional): Glob patterns for files to exclude.
- `resources` (object, optional): Configuration for resource files.
  - `include` (array, optional): Glob patterns for files to include. Defaults to `["**/*.md"]`.
  - `exclude` (array, optional): Glob patterns for files to exclude.

**Note:** Glob pattern matching is powered by [doublestar](https://pkg.go.dev/github.com/bmatcuk/doublestar/v4).

For private repositories, create a [personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-fine-grained-personal-access-token) with read-only permissions and provide it as the `password`.

**Note:** SSH authentication is not supported intentionally due to the difficulty of implementing fine-grained access control with SSH keys. If you believe SSH support should be added, please [reach out](https://github.com/appgardenstudios/slash-prompt/discussions/categories/feedback).

### Environment Variable Substitution

Configuration files support environment variable substitution using `${VARIABLE_NAME}` syntax. This allows you to keep sensitive information like credentials out of your configuration files.

Example:

```yml
repos:
  - repo: https://github.com/appgardenstudios/hyaline.git
    id: hyaline
    ref: main
      username: ${SLASH_PROMPT_USERNAME}
      password: ${SLASH_PROMPT_PASSWORD}
    prompts:
      path: contributing/prompts
      include:
        - "**/*.md"
      exclude:
        - "resources/**/*"
```

## Installation

### Docker

The MCP server is run as a docker container over STDIO, and so Docker is a prerequisite. When running the docker container, the configuration file must be mounted.

Example:
```bash
$ docker run -i --rm -v ./prompt.yml:/home/appuser/.prompt.yml:ro ghcr.io/appgardenstudios/slash-prompt:latest
```

The default location for the mounted configuration file is `/home/appuser/.prompt.yml`; however, this can be overridden by providing a `SLASH_PROMPT_CONFIG_PATH` environment variable.

See [https://github.com/appgardenstudios/slash-prompt/pkgs/container/slash-prompt](https://github.com/appgardenstudios/slash-prompt/pkgs/container/slash-prompt) for the list of available images.

### Host Applications

Installation varies depending on the host application. Refer to the current documentation for your specific host application.

#### Claude Code

Install the MCP server using the Claude Code CLI:

```bash
claude mcp add prompt -- docker run -i --rm -v ./prompt.yml:/home/appuser/.prompt.yml:ro ghcr.io/appgardenstudios/slash-prompt:latest
```

#### Claude Desktop

Add the following configuration to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "prompt": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "-v", "./prompt.yml:/home/appuser/.prompt.yml:ro",
        "ghcr.io/appgardenstudios/slash-prompt:latest"
      ]
    }
  }
}
```

## Usage

### MCP Tools

The following MCP tools are available:

- **`listErrors`**: List all loading errors that occurred during startup
- **`listResources`**: List all available resources, optionally filtered by repository
  - `repo` (optional): Repository filter
- **`getResource`**: Get a specific resource by its URI
  - `resource_uri` (required): The fully qualified resource URI. See [MCP Resources](#mcp-resources)

### MCP Prompts

Host applications expose prompts in different ways. Refer to the current documentation for your specific host application.

Prompts can be accessed by name (when they are unique) or with a repository ID prefix:

- `generate-spec` - Access the "generate-spec" prompt 
- `hyaline:analyze` - Access the "analyze" prompt specifically from the "hyaline" repository

In Claude Code:
```
/prompt:load-documentation
```

### MCP Resources

Resources are available through the MCP resource system. Resource URIs follow the schema `file://<repo-id>/<path-to-resource>`. Example: `file://hyaline/contributing/resources/documentation-summary.template.md`

Resources that are referenced by [prompt files](#prompts) are automatically embedded in the prompt. They can also be retrieved using the provided [resource tools](#mcp-tools).

Host applications expose prompts in different ways. Refer to the current documentation for your specific host application.

In Claude Code:
```
@prompt:file://hyaline/contributing/prompts/resources/spec.template.md
```

## Development

### Prerequisites

- Go 1.24.4 or later
- Git
- Docker (optional)

### Building

Build the Docker image:

```bash
make build-docker
```

### Testing

#### Unit Tests

Run unit tests:

```bash
make test
```

#### E2E Tests

Run end-to-end tests:

```bash
make e2e
```

Update E2E golden files:

```bash
make e2e-update
```

#### Manual Testing

For local development and testing:

1. Build a local Docker image:
   ```bash
   make build-docker
   ```
2. Set the `SLASH_PROMPT_IMAGE` environment variable to use your local image:
   ```bash
   export SLASH_PROMPT_IMAGE=slash-prompt:development
   ```
3. The `.mcp.json` file is configured to use this environment variable, allowing you to test with your local build using Claude Code.
    - Note: Due to a current bug with Claude Code, you must inline the environment variable: `SLASH_PROMPT_IMAGE=slash-prompt:development claude`

### Releasing

The release process is automated with the release script. It must be run from the `main` branch:

```bash
$ ./scripts/release.sh
```

## License

This project is licensed under the terms of the MIT open source license. Please refer to [MIT](./LICENSE) for the full terms.

## Support

- [Create a Bug](https://github.com/appgardenstudios/slash-prompt/issues) for bug reports
- [Start a discussion](https://github.com/appgardenstudios/slash-prompt/discussions/categories/questions) for questions
