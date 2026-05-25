# Jules API Notes

Juleson targets the Jules API v1alpha endpoints under:

```text
https://jules.googleapis.com/v1alpha
```

The client is implemented under `internal/jules`.

## Resources Used

- Sources: list and get connected source repositories.
- Sessions: create, list, get, approve plans, and send messages.
- Activities: list and get activity details.
- Artifacts: download patch, output, and media artifacts from activities.

## Authentication

Set `JULES_API_KEY` or `jules.api_key` in `juleson.yaml`.

```bash
export JULES_API_KEY="..."
juleson sessions list
```

## Session Lifecycle Limits

The API operations used by Juleson do not expose session cancel or delete. Use the
Jules web UI for those operations when needed.

## Patch Workflow

Juleson can preview and apply patches from session artifacts:

```bash
juleson sessions preview SESSION_ID
juleson sessions download SESSION_ID ./artifacts
```

MCP tools also expose `preview_session_changes` and `apply_session_patches`.
