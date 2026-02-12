# Jira MCP

An opinionated Jira MCP server built from years of real-world software development experience.

Unlike generic Jira integrations, this MCP is crafted from the daily workflows of engineers and automation QC teams. You'll find sophisticated tools designed for actual development needs—like retrieving all pull requests linked to an issue, managing complex sprint transitions, or tracking development information across your entire workflow.

This isn't just another API wrapper. It's a reflection of how professionals actually use Jira: managing sprints, tracking development work, coordinating releases, and maintaining visibility across teams. Every tool is designed to solve real problems that arise in modern software development.

## Available Tools

### Issue Management
- **jira_get_issue** - Retrieve detailed information about a specific issue including status, assignee, description, subtasks, and available transitions
- **jira_create_issue** - Create a new issue with specified details (returns key, ID, and URL)
- **jira_create_child_issue** - Create a child issue (sub-task) linked to a parent issue
- **jira_update_issue** - Modify an existing issue's details (supports partial updates)
- **jira_delete_issue** - Delete an issue permanently
- **jira_list_issue_types** - List all available issue types in a project with their IDs, names, and descriptions

### Search
- **jira_search_issue** - Search for issues using JQL (Jira Query Language) with customizable fields and expand options

### Sprint Management
- **jira_list_sprints** - List all active and future sprints for a specific board or project
- **jira_get_sprint** - Retrieve detailed information about a specific sprint by its ID
- **jira_get_active_sprint** - Get the currently active sprint for a given board or project
- **jira_search_sprint_by_name** - Search for sprints by name with exact or partial matching

### Status & Transitions
- **jira_list_statuses** - Retrieve all available issue status IDs and their names for a project
- **jira_transition_issue** - Transition an issue through its workflow using a valid transition ID

### Comments
- **jira_add_comment** - Add a comment to an issue (uses Atlassian Document Format)
- **jira_get_comments** - Retrieve all comments from an issue

### Worklogs
- **jira_add_worklog** - Add a worklog entry to track time spent on an issue

### History & Audit
- **jira_get_issue_history** - Retrieve the complete change history of an issue

### Issue Relationships
- **jira_get_related_issues** - Retrieve issues that have a relationship (blocks, is blocked by, relates to, etc.)
- **jira_link_issues** - Create a link between two issues, defining their relationship

### Version Management
- **jira_get_version** - Retrieve detailed information about a specific project version
- **jira_list_project_versions** - List all versions in a project with their details

### Development Information
- **jira_get_development_information** - Retrieve branches, pull requests, and commits linked to an issue via development tool integrations (GitHub, GitLab, Bitbucket)

### Attachments
- **jira_download_attachment** - Download a Jira attachment to a local temporary file

## Installation

### Docker (recommended)

```bash
docker pull tedyno/jira-mcp:latest
```

### Go

```bash
go install github.com/nguyenvanduocit/jira-mcp@latest
```

Or build from source:

```bash
git clone https://github.com/MountainLift/jira-mcp.git
cd jira-mcp
go build -o jira-mcp .
```

## Configuration

Create an API token at [Atlassian API tokens](https://id.atlassian.com/manage-profile/security/api-tokens).

Set the following environment variables:

- **ATLASSIAN_HOST** — your Atlassian instance URL (e.g. `https://your-company.atlassian.net`)
- **ATLASSIAN_EMAIL** — your Atlassian account email
- **ATLASSIAN_TOKEN** — API token

Or use a `.env` file:

```bash
ATLASSIAN_HOST=https://your-company.atlassian.net
ATLASSIAN_EMAIL=your-email@company.com
ATLASSIAN_TOKEN=your-api-token
```

## Usage with Claude Code

### Docker

```json
{
  "mcpServers": {
    "jira": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-e", "ATLASSIAN_HOST", "-e", "ATLASSIAN_EMAIL", "-e", "ATLASSIAN_TOKEN", "tedyno/jira-mcp:latest"],
      "env": {
        "ATLASSIAN_HOST": "https://your-company.atlassian.net",
        "ATLASSIAN_EMAIL": "your-email@company.com",
        "ATLASSIAN_TOKEN": "your-api-token"
      }
    }
  }
}
```

Or via CLI:

```bash
claude mcp add jira -e ATLASSIAN_HOST=https://your-company.atlassian.net -e ATLASSIAN_EMAIL=your-email@company.com -e ATLASSIAN_TOKEN=your-api-token -- docker run -i --rm -e ATLASSIAN_HOST -e ATLASSIAN_EMAIL -e ATLASSIAN_TOKEN tedyno/jira-mcp:latest
```

### Binary

```json
{
  "mcpServers": {
    "jira": {
      "command": "jira-mcp",
      "env": {
        "ATLASSIAN_HOST": "https://your-company.atlassian.net",
        "ATLASSIAN_EMAIL": "your-email@company.com",
        "ATLASSIAN_TOKEN": "your-api-token"
      }
    }
  }
}
```

Or via CLI:

```bash
claude mcp add jira -e ATLASSIAN_HOST=https://your-company.atlassian.net -e ATLASSIAN_EMAIL=your-email@company.com -e ATLASSIAN_TOKEN=your-api-token -- jira-mcp
```

## License

MIT — see `LICENSE`.

## FOR AI

> THIS SECTION IS FOR AI ONLY

When working with this codebase, read these files to understand the project structure:

1. **CLAUDE.md** - Comprehensive project documentation including architecture, development commands, and coding conventions
2. **main.go** - Entry point that shows how the MCP server is initialized and tools are registered
3. **services/jira_client.go** - Singleton Jira client initialization and authentication
4. **tools/** - Individual tool implementations following consistent patterns
5. **docs/** - Detailed documentation (see structure below)

Key concepts:
- This is a Go-based MCP server that connects AI assistants to Jira
- Each tool follows a registration + handler pattern with typed input validation
- Tools are organized by category (issues, sprints, comments, worklogs, etc.)
- All Jira operations use the `github.com/ctreminiom/go-atlassian` client library
- Development principles documented in `.specify/memory/constitution.md`

Before making changes, review:
- **CLAUDE.md** for architecture patterns and development commands
- **.specify/memory/constitution.md** for governance principles
