# Jules API Documentation

This directory contains the official API specifications and documentation for
the Jules API integration.

## Overview

The Jules API lets you programmatically access Jules's capabilities to automate
and enhance your software development lifecycle. You can use the API to create
custom workflows, automate tasks like bug fixing and code reviews, and embed
Jules's intelligence directly into the tools you use every day.

**Status**: Alpha release (experimental - specifications may change)

**Important Notes:**

- The Jules API is in alpha release and is experimental
- Specifications, API keys, and definitions may change
- Keep API keys secure - publicly exposed keys are automatically disabled
- Sources must be connected via Jules GitHub app before API usage
- Sessions created via API have plans auto-approved by default (unless `requirePlanApproval: true`)

## Base URL

`https://jules.googleapis.com`

## Authentication

All API requests require authentication via the `X-Goog-Api-Key` header:

`X-Goog-Api-Key: YOUR_API_KEY`

### Getting an API Key

1. Visit [Jules Settings](https://jules.google.com/settings#api)
2. Create a new API key (maximum 3 keys per account)
3. Keep keys secure - exposed keys are automatically disabled

## Core Concepts

### Source

An input source for the agent (e.g., a GitHub repository). Must be connected via
the Jules GitHub app before API usage.

### Session

A continuous unit of work within a specific context, similar to a chat session.
Initiated with a prompt and a source.

### Activity

A single unit of work within a Session. Contains multiple activities from both
user and agent (plans, messages, progress updates).

## REST Resources

The Jules API provides the following REST resources:

### v1alpha.sources

| Method | HTTP Request | Description |
|--------|--------------|-------------|
| `get` | `GET /v1alpha/{name=sources/**}` | Gets a single source |
| `list` | `GET /v1alpha/sources` | Lists sources |

### v1alpha.sessions

| Method | HTTP Request | Description |
|--------|--------------|-------------|
| `approvePlan` | `POST /v1alpha/{session=sessions/*}:approvePlan` | Approves a plan in a session |
| `create` | `POST /v1alpha/sessions` | Creates a new session |
| `get` | `GET /v1alpha/{name=sessions/*}` | Gets a single session |
| `list` | `GET /v1alpha/sessions` | Lists all sessions |
| `sendMessage` | `POST /v1alpha/{session=sessions/*}:sendMessage` | Sends a message from the user to a session |

### v1alpha.sessions.activities

| Method | HTTP Request | Description |
|--------|--------------|-------------|
| `get` | `GET /v1alpha/{name=sessions/*/activities/*}` | Gets a single activity |
| `list` | `GET /v1alpha/{parent=sessions/*}/activities` | Lists activities for a session |

## API Endpoints

### Sources

#### List Sources

```http
GET /v1alpha/sources?filter={string}&pageSize={number}&pageToken={token}
```

**Query Parameters:**

- `filter` (optional): Filter expression for listing sources (AIP-160).
  Example: `name=sources/source1 OR name=sources/source2`
- `pageSize` (optional): Number of sources to return. Must be 1-100 (default: 30, max: 100)
- `pageToken` (optional): Page token from previous `sources.list` call

**Response:**

```json
{
  "sources": [
    {
      "name": "sources/github/owner/repo",
      "id": "github/owner/repo",
      "githubRepo": {
        "owner": "owner",
        "repo": "repo",
        "isPrivate": false,
        "defaultBranch": {
          "displayName": "main"
        },
        "branches": [
          {
            "displayName": "main"
          },
          {
            "displayName": "develop"
          }
        ]
      }
    }
  ],
  "nextPageToken": "string"
}
```

#### Get Source

```http
GET /v1alpha/sources/{sourceName}
```

**Path Parameters:**

- `sourceName`: Full source name (e.g., `sources/github/owner/repo`)

**Response:** Single source object (same format as above)

### Sessions

#### List Sessions

```http
GET /v1alpha/sessions?pageSize={number}&pageToken={token}
```

**Query Parameters:**

- `pageSize` (optional): Number of sessions to return. Must be 1-100 (default: 30, max: 100)
- `pageToken` (optional): Page token from previous `sessions.list` call

**Response:**

```json
{
  "sessions": [
    {
      "name": "sessions/123456789",
      "id": "123456789",
      "title": "Session Title",
      "state": "COMPLETED|IN_PROGRESS|PLANNING|AWAITING_PLAN_APPROVAL|AWAITING_USER_FEEDBACK|QUEUED|PAUSED|FAILED",
      "createTime": "2025-01-01T00:00:00.000000Z",
      "updateTime": "2025-01-01T00:00:00.000000Z",
      "sourceContext": {
        "source": "sources/github/owner/repo",
        "githubRepoContext": {
          "startingBranch": "main"
        }
      },
      "prompt": "Initial prompt",
      "requirePlanApproval": false,
      "automationMode": "AUTO_CREATE_PR",
      "url": "https://jules.google.com/sessions/123456789",
      "outputs": [
        {
          "pullRequest": {
            "url": "https://github.com/owner/repo/pull/123",
            "title": "PR Title",
            "description": "PR Description"
          }
        }
      ]
    }
  ],
  "nextPageToken": "string"
}
```

#### Get Session

```http
GET /v1alpha/sessions/{sessionId}
```

**Response:** Single session object (same format as above)

#### Create Session

```http
POST /v1alpha/sessions
Content-Type: application/json
```

**Request Body:**

```json
{
  "prompt": "Create a boba app!",
  "sourceContext": {
    "source": "sources/github/owner/repo",
    "githubRepoContext": {
      "startingBranch": "main"
    }
  },
  "title": "Session Title (optional)",
  "automationMode": "AUTO_CREATE_PR (optional)",
  "requirePlanApproval": false
}
```

**Request Parameters:**

- `prompt` (required): The initial prompt for the session
- `sourceContext` (required): Source information
  - `source` (required): Source identifier (e.g., "sources/github/owner/repo")
  - `githubRepoContext` (optional): GitHub-specific context
    - `startingBranch` (required if githubRepoContext provided): Branch to work on
- `title` (optional): Human-readable title for the session
- `automationMode` (optional): Automation behavior
  - `"AUTO_CREATE_PR"`: Automatically create a pull request when complete
  - `"AUTOMATION_MODE_UNSPECIFIED"` or empty: No automatic PR creation
- `requirePlanApproval` (optional): Whether to require explicit plan approval (default: false)

**Response:** Session object (same format as above)

#### Send Message

```http
POST /v1alpha/sessions/{sessionId}:sendMessage
Content-Type: application/json
```

**Request Body:**

```json
{
  "prompt": "Can you make the app corgi themed?"
}
```

**Request Parameters:**

- `prompt` (required): The user prompt to send to the session

**Response:** Empty (agent response appears in next activity)

#### Approve Plan

```http
POST /v1alpha/sessions/{sessionId}:approvePlan
```

**Path Parameters:**

- `sessionId`: Session resource name (format: `sessions/{session}`)

**Request Body:** Empty

**Response:** Empty

### Activities

#### List Activities

```http
GET /v1alpha/sessions/{sessionId}/activities?pageSize={number}&pageToken={token}
```

**Path Parameters:**

- `sessionId`: Parent session resource name (format: `sessions/{session}`)

**Query Parameters:**

- `pageSize` (optional): Number of activities to return. Must be 1-100 (default: 50, max: 100)
- `pageToken` (optional): Page token from previous `activities.list` call

**Response:**

```json
{
  "activities": [
    {
      "name": "sessions/123/activities/abc123",
      "id": "abc123",
      "description": "Activity description",
      "createTime": "2025-01-01T00:00:00.000000Z",
      "originator": "agent|user|system",
      "agentMessaged": {
        "agentMessage": "Agent response message"
      },
      "userMessaged": {
        "userMessage": "User input message"
      },
      "planGenerated": {
        "plan": {
          "id": "plan123",
          "createTime": "2025-01-01T00:00:00.000000Z",
          "steps": [
            {
              "id": "step1",
              "title": "Setup the environment. I will install the dependencies to run the app.",
              "description": "Install dependencies",
              "index": 0
            },
            {
              "id": "step2",
              "title": "Modify src/App.js",
              "description": "Replace boilerplate with Boba-themed component",
              "index": 1
            }
          ]
        }
      },
      "planApproved": {
        "planId": "plan123"
      },
      "progressUpdated": {
        "title": "Ran bash command",
        "description": "Command: npm install\nOutput: added 1326 packages, and audited 1327 packages in 25s\n\n268 packages are looking for funding\nExit Code: 0"
      },
      "sessionCompleted": {},
      "sessionFailed": {
        "reason": "Error reason"
      },
      "artifacts": [
        {
          "bashOutput": {
            "command": "npm install",
            "output": "added 1326 packages, and audited 1327 packages in 25s\n\n268 packages are looking for funding",
            "exitCode": 0
          }
        },
        {
          "changeSet": {
            "source": "sources/github/owner/repo",
            "gitPatch": {
              "unidiffPatch": "diff --git a/src/App.js b/src/App.js\nindex 1234567..abcdef0 100644\n--- a/src/App.js\n+++ b/src/App.js\n@@ -1,1 +1,1 @@\n-Hello World\n+Boba App",
              "baseCommitId": "36ead0a4caefc451b9652ed926a15af9570f4f35",
              "suggestedCommitMessage": "feat: Create simple Boba App\n\nThis commit transforms the default Create React App boilerplate into a simple, visually appealing Boba-themed application."
            }
          }
        },
        {
          "media": {
            "data": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==",
            "mimeType": "image/png"
          }
        }
      ]
    }
  ],
  "nextPageToken": "string"
}
```

#### Get Activity

```http
GET /v1alpha/sessions/{sessionId}/activities/{activityId}
```

**Path Parameters:**

- `sessionId`: Session identifier
- `activityId`: Activity identifier (full format: `sessions/{session}/activities/{activity}`)

**Response:** Single activity object (same format as above)

## Session States

- `STATE_UNSPECIFIED`: The state is unspecified
- `QUEUED`: The session is queued
- `PLANNING`: The agent is planning
- `AWAITING_PLAN_APPROVAL`: The agent is waiting for plan approval
- `AWAITING_USER_FEEDBACK`: The agent is waiting for user feedback
- `IN_PROGRESS`: The session is in progress
- `PAUSED`: The session is paused
- `FAILED`: The session has failed
- `COMPLETED`: The session has completed

## Activity Types

Activities can contain one of the following event types:

### agentMessaged

The agent posted a message.

```json
{
  "agentMessaged": {
    "agentMessage": "Agent response message"
  }
}
```

### userMessaged

The user posted a message.

```json
{
  "userMessaged": {
    "userMessage": "User input message"
  }
}
```

### planGenerated

Agent has generated a plan for the session.

```json
{
  "planGenerated": {
    "plan": {
      "id": "plan123",
      "createTime": "2025-01-01T00:00:00.000000Z",
      "steps": [
        {
          "id": "step1",
          "title": "Step title",
          "description": "Step description",
          "index": 0
        }
      ]
    }
  }
}
```

### planApproved

User has approved a plan for execution.

```json
{
  "planApproved": {
    "planId": "plan123"
  }
}
```

### progressUpdated

Agent is providing progress updates during execution.

```json
{
  "progressUpdated": {
    "title": "Progress title",
    "description": "Detailed progress description"
  }
}
```

### sessionCompleted

Session has finished execution.

```json
{
  "sessionCompleted": {}
}
```

### sessionFailed

Session has failed.

```json
{
  "sessionFailed": {
    "reason": "Error reason"
  }
}
```

## Artifact Types

Activities can include artifacts that provide additional context:

### bashOutput

Results from terminal commands executed by the agent.

```json
{
  "bashOutput": {
    "command": "npm install",
    "output": "added 1326 packages, and audited 1327 packages in 25s\n\n268 packages are looking for funding",
    "exitCode": 0
  }
}
```

### changeSet

Code changes made during the session, including git patches.

```json
{
  "changeSet": {
    "source": "sources/github/owner/repo",
    "gitPatch": {
      "unidiffPatch": "diff --git a/src/App.js b/src/App.js\nindex 1234567..abcdef0 100644\n--- a/src/App.js\n+++ b/src/App.js\n@@ -1,1 +1,1 @@\n-Hello World\n+Boba App",
      "baseCommitId": "36ead0a4caefc451b9652ed926a15af9570f4f35",
      "suggestedCommitMessage": "feat: Create simple Boba App\n\nThis commit transforms the default Create React App boilerplate into a simple, visually appealing Boba-themed application."
    }
  }
}
```

### media

Screenshots, images, or other media generated during execution.

```json
{
  "media": {
    "data": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==",
    "mimeType": "image/png"
  }
}
```

## Error Handling

API returns standard HTTP status codes:

- `200`: Success
- `400`: Bad Request
- `401`: Unauthorized (invalid API key)
- `404`: Not Found
- `429`: Rate Limited
- `500`: Internal Server Error

Error responses include details in the response body.

## API Limitations

### Session Lifecycle Management

**Current Limitations (as of v1alpha):**

- ❌ **No Cancel Operation**: There is no API endpoint to cancel an in-progress session
  - Cancel is only available through the Jules web UI at `session.url`
  - Sessions must run to completion or failure

- ❌ **No Delete Operation**: There is no API endpoint to delete sessions
  - Delete is only available through the Jules web UI
  - Sessions remain accessible via API indefinitely

**Workarounds:**

- To stop a session: Use the web UI via the `url` field from session response
- Session states (`FAILED`, `COMPLETED`, `PAUSED`) indicate terminal states
- Use `sessions.list` with pagination to manage large session lists

**Note:** These limitations are expected in the alpha release. Future API versions may include
session lifecycle management endpoints.

## Rate Limits

Rate limits apply to API usage. Monitor for 429 responses and implement backoff.

## Examples

### Complete Workflow

```bash
# 1. List sources
curl 'https://jules.googleapis.com/v1alpha/sources' \
  -H 'X-Goog-Api-Key: YOUR_API_KEY'

# 2. Create session
curl 'https://jules.googleapis.com/v1alpha/sessions' \
  -X POST \
  -H 'Content-Type: application/json' \
  -H 'X-Goog-Api-Key: YOUR_API_KEY' \
  -d '{
    "prompt": "Create a React app",
    "sourceContext": {
      "source": "sources/github/owner/repo"
    },
    "automationMode": "AUTO_CREATE_PR"
  }'

# 3. List activities to see progress
curl 'https://jules.googleapis.com/v1alpha/sessions/SESSION_ID/activities' \
  -H 'X-Goog-Api-Key: YOUR_API_KEY'

# 4. Send follow-up message
curl 'https://jules.googleapis.com/v1alpha/sessions/SESSION_ID:sendMessage' \
  -X POST \
  -H 'Content-Type: application/json' \
  -H 'X-Goog-Api-Key: YOUR_API_KEY' \
  -d '{"prompt": "Make it blue themed"}'
```

## Official Documentation

For the complete and most up-to-date API reference, visit:

- Main Documentation: <https://developers.google.com/jules/api>
- REST API Reference: <https://developers.google.com/jules/api/reference/rest>

Last updated: 2025-11-01 UTC
