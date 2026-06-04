# Jules API Notes

Juleson targets the Jules API v1alpha endpoints under:

```text
https://jules.googleapis.com/v1alpha
```

The public Go SDK is published as `github.com/SamyRai/go-jules`. Local
filesystem and `git apply` operations live in Juleson's internal workspace code so
the SDK remains reusable without app side effects.

Official references:

- [Sessions](https://jules.google/docs/api/reference/sessions/)
- [Activities](https://jules.google/docs/api/reference/activities/)
- [Sources](https://jules.google/docs/api/reference/sources/)
- [Types](https://jules.google/docs/api/reference/types/)
- [January 26, 2026 changelog](https://jules.google/docs/changelog/2026-01-26-4/)

## Resources Used

- Sources: list and get connected source repositories.
- Sessions: create, list, get, delete, approve plans, and send messages.
- Activities: list and get activity details. Activity cursor filtering is
  applied client-side for compatibility with live API behavior.
- Artifacts: read embedded `changeSet`, `bashOutput`, and base64 `media`
  payloads from activities.
- Outputs: surface documented session outputs such as pull requests from session
  responses.

## Go SDK

```go
import jules "github.com/SamyRai/go-jules"

client := jules.NewClient(
    "api-key",
    jules.WithBaseURL("https://jules.googleapis.com/v1alpha"),
    jules.WithTimeout(30*time.Second),
    jules.WithRetryAttempts(3),
    jules.WithRetryBackoff(time.Second),
    jules.WithUserAgent("my-tool/1.0"),
    jules.WithLogger(slog.Default()),
    jules.WithDebugLog(true),
)
```

Debug logging for HTTP requests can be enabled via `jules.WithDebugLog(true)`
and requires setting `jules.WithLogger`. When enabled, it logs HTTP method,
URL, duration, status code, and errors at the `Debug` level, while redacting any
values resembling keys or tokens in query parameters to ensure secrets are not
exposed.

The SDK uses typed session states, automation modes, activity originators, and
`time.Time` values for documented RFC3339 timestamps. Methods accepting resource
names normalize both bare IDs and full names, so `123` and `sessions/123` are
equivalent for sessions. Slash-containing source names such as
`sources/github/owner/repo` are preserved as path segments for source get calls.

The SDK exposes documented Jules API resources plus pure helpers for embedded
artifact payloads. It does not call undocumented artifact `/content`,
`/download`, `/analyze`, or activity `/search` endpoints. Juleson CLI and MCP
commands layer local download, preview, backup, and patch application behavior on
top of the SDK through `internal/jules/workspace`.

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
juleson sessions create . --prompt-file task.md
```

Repoless sessions omit `sourceContext`:

```bash
juleson sessions create --no-source "Sketch an implementation plan"
```

Both forms call `POST /sessions`.

For GitHub-backed sources, Juleson includes `githubRepoContext.startingBranch`.
If a caller provides only the source name, the SDK reads the source metadata and
uses the connected repository's default branch before creating the session. This
avoids the Jules API's generic `INVALID_ARGUMENT` response for source-backed
session creates that omit branch context.

For CLI source inference, `.` is resolved from the local git `origin` remote to a
connected Jules source. If multiple connected sources match the same
owner/repository, Juleson fails with the candidates instead of guessing. Batch
creation keeps a local `batch_id`/`group_title` prompt convention and loops over
`POST /sessions`; the REST reference does not expose a documented bulk-create
endpoint.

## Session Lifecycle

Juleson supports the documented session delete endpoint:

```bash
juleson sessions delete SESSION_ID --force
```

The MCP `delete_session` tool requires `confirm=true`.

The Jules API v1alpha reference used by this project does not expose a cancel
endpoint. Use the Jules web UI to cancel a running session.

## Activity Filtering

The activity list endpoint supports pagination. Although some Jules docs show a
`createTime` query parameter, live API responses can reject it as an unknown
field. Juleson therefore performs `createTime`, type, status, plan, and artifact
filtering client-side after fetching activities.

For immutable activity streams, SDK callers can use `ListActivitiesSince` with a
stored `createTime` cursor and `ActivityCursor` to compute the next cursor from a
batch.

Plan inspection is built from documented activity payloads. `sessions plans` and
the MCP `get_session_plans` tool read activities with `planGenerated` and
`planApproved`, then summarize plan IDs, activity IDs/resource names, creation
times, approval state, and every step title and description. The MCP response
keeps the existing raw `activities` field and adds a structured `plans` field.

Full file outputs are mentioned in the upstream changelog, but the public
reference does not document a stable typed response shape in the pages above.
Juleson does not model a file output type until that schema is confirmed.

## Patch Workflow

Juleson can preview and apply patches from session artifacts:

```bash
juleson sessions plans SESSION_ID
juleson sessions review SESSION_ID ./repo
juleson sessions artifacts list SESSION_ID
juleson sessions apply SESSION_ID ./repo --activity-id ACTIVITY_ID --artifact-index 0
juleson sessions apply SESSION_ID ./repo --confirm
juleson sessions download SESSION_ID ./artifacts
```

MCP tools also expose `review_session`, `list_session_artifacts`, and
`get_session_outputs`.

`sessions review` and MCP `review_session` are read-only operator snapshots.
They combine session state, latest plan, documented outputs, artifact manifests,
patch dry-run preview, base commit mismatch warnings, dirty-worktree blockers,
verification suggestions, and safe next actions. They do not approve plans,
send messages, mutate worktrees, or apply patches; when safe they only
recommend the exact `sessions apply ... --confirm` command.

Preview and apply can be scoped to one activity and artifact index. `git apply
--check` remains the source of truth for whether a patch can apply. Juleson
parses patch metadata for changed files, deletes, renames, binary markers, and
paths with spaces only for previews and manifests. When `gitPatch.baseCommitId`
is present, dry-run reports mismatches and mutation blocks unless the caller
sets `--allow-base-mismatch` or `allow_base_mismatch=true`.

## Verification And PR Outputs

`verify_session_changes` detects Go (`go test`), Node/Yarn (`yarn test`),
Python/uv (`uv run pytest`), and Rust (`cargo test`) from project files. The
explicit command escape hatch is opt-in and only runs when supplied by the user
or caller.

`sessions outputs` and `get_session_outputs` surface documented pull request
outputs. Juleson reports PR URLs and leaves general GitHub and Actions handling
to `gh`, GitHub's CLI, or the official GitHub MCP server.

## Unsupported Or Deferred Jules Features

Session cancel remains unavailable in the referenced Jules REST v1alpha API.
Web-only suggested performance tasks and any unstable output fields are not
modeled until they appear in documented REST responses.
