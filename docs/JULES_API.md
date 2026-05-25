# Jules API Notes

Juleson targets the Jules API v1alpha endpoints under:

```text
https://jules.googleapis.com/v1alpha
```

The public Go SDK is implemented under `pkg/jules` and can be imported as
`github.com/SamyRai/juleson/pkg/jules`. Local filesystem and `git apply`
operations live in `internal/julesops` so the SDK remains reusable without app
side effects.

Official references:

- [Sessions](https://jules.google/docs/api/reference/sessions/)
- [Activities](https://jules.google/docs/api/reference/activities/)
- [Sources](https://jules.google/docs/api/reference/sources/)
- [Types](https://jules.google/docs/api/reference/types/)
- [January 26, 2026 changelog](https://jules.google/docs/changelog/2026-01-26-4/)

## Resources Used

- Sources: list and get connected source repositories.
- Sessions: create, list, get, delete, approve plans, and send messages.
- Activities: list and get activity details. The list call supports the official
  `createTime` filter.
- Artifacts: read embedded `changeSet`, `bashOutput`, and base64 `media`
  payloads from activities.

## Go SDK

```go
client := jules.NewClient(
	"api-key",
	jules.WithBaseURL("https://jules.googleapis.com/v1alpha"),
	jules.WithTimeout(30*time.Second),
	jules.WithRetryAttempts(3),
	jules.WithRetryBackoff(time.Second),
	jules.WithUserAgent("my-tool/1.0"),
)
```

The SDK uses typed session states, automation modes, activity originators, and
`time.Time` values for documented RFC3339 timestamps. Methods accepting resource
names normalize both bare IDs and full names, so `123` and `sessions/123` are
equivalent for sessions, and slash-containing source names are path-escaped for
API calls.

The SDK exposes documented Jules API resources plus pure helpers for embedded
artifact payloads. It does not call undocumented artifact `/content`,
`/download`, `/analyze`, or activity `/search` endpoints. Juleson CLI and MCP
commands layer local download, preview, backup, and patch application behavior on
top of the SDK through `internal/julesops`.

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

For immutable activity streams, SDK callers can use `ListActivitiesSince` with a
stored `createTime` cursor and `ActivityCursor` to compute the next cursor from a
batch.

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
