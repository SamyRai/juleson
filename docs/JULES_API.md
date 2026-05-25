# Jules API Notes

Juleson targets the Jules API v1alpha endpoints under:

```text
https://jules.googleapis.com/v1alpha
```

The client is implemented under `internal/jules`.

Official references:

- [Sessions](https://jules.google/docs/api/reference/sessions/)
- [Activities](https://jules.google/docs/api/reference/activities/)
- [Sources](https://jules.google/docs/api/reference/sources/)
- [January 26, 2026 changelog](https://jules.google/docs/changelog/2026-01-26-4/)

## Resources Used

- Sources: list and get connected source repositories.
- Sessions: create, list, get, delete, approve plans, and send messages.
- Activities: list and get activity details. The list call supports the official
  `createTime` filter.
- Artifacts: download patch, output, and media artifacts from activities.

## Authentication

Set `JULES_API_KEY` or `jules.api_key` in `juleson.yaml`.

```bash
export JULES_API_KEY="..."
juleson sessions list
```

## Session Creation

Source-backed sessions send `sourceContext`:

```bash
juleson sessions create sources/github/owner/repo "Fix failing tests"
```

Repoless sessions omit `sourceContext`:

```bash
juleson sessions create --no-source "Sketch an implementation plan"
```

Both forms call `POST /sessions`.

## Session Lifecycle

Juleson supports the documented session delete endpoint:

```bash
juleson sessions delete SESSION_ID --force
```

The MCP `delete_session` tool requires `confirm=true`.

The Jules API v1alpha reference used by this project does not expose a cancel
endpoint. Use the Jules web UI to cancel a running session.

## Activity Filtering

The official activity list endpoint supports pagination and `createTime`.
Legacy helpers such as type, status, plan, and artifact filters are applied
client-side after fetching activities; they are not sent as unsupported API
query parameters.

Full file outputs are mentioned in the upstream changelog, but the public
reference does not document a stable typed response shape in the pages above.
Juleson does not model a file output type until that schema is confirmed.

## Patch Workflow

Juleson can preview and apply patches from session artifacts:

```bash
juleson sessions preview SESSION_ID
juleson sessions download SESSION_ID ./artifacts
```

MCP tools also expose `preview_session_changes` and `apply_session_patches`.
