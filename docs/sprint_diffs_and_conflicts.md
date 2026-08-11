# Sprint Plan: Juleson CLI UX Improvements for Diffs and Conflicts

## Overview

This sprint focuses on improving the developer experience for diffs and merge
conflicts in the `juleson` CLI. Instead of relying only on manual 3-way merges
or raw text diffs, the CLI should help the developer guide Jules through an
interactive terminal workflow.

## Goals

1. **Enhanced Diff Viewing:** provide structural, syntax-highlighted diffs using
   external tools such as `difftastic` or `delta`, with native Go fallbacks.
2. **Context Gathering TUI:** provide an interactive terminal wizard that bundles
   local state, patch content, and compiler output when a conflict occurs.
3. **Agentic Dispatch:** route the bundled context back to the Jules session and
   surface a detailed `resolution_report.md`.

## Epic 1: Better Structured & Formatted Diffs

**Objective:** Upgrade `juleson sessions preview <ID>` to produce readable, syntax-highlighted diffs.

- **Task 1.1: Shell out wrapper for modern diff tools**
  - Check for external diff pagers in the user's `$PATH`, specifically
    `difftastic` or `delta`.
  - If present, pipe patch output through the external tool when running
    `juleson sessions preview`.
- **Task 1.2: Native Go Fallback implementation**
  - Add `github.com/bluekeyes/go-gitdiff` or `github.com/sourcegraph/go-diff`
    to parse raw unified diff patches.
  - Integrate `github.com/alecthomas/chroma` for code syntax highlighting.
  - Use `charmbracelet/lipgloss` to render highlighted diffs in the terminal.
- **Task 1.3: Add configuration flags**
  - Add flags or global config settings to force the native fallback or set a
    preferred diff tool path.

---

## Epic 2: Conflict Context-Gathering UI

**Objective:** When `juleson sessions apply <ID>` encounters a conflict, launch
an interactive wizard to gather resolution context.

- **Task 2.1: Implement TUI using `charmbracelet/huh`**
  - Create a new package `internal/tui/conflict` to house the form application
    logic.
  - Use `charmbracelet/huh` to implement a multi-select context form:
    - `[x]` Current state of the local file
    - `[x]` The failing patch diff
    - `[ ]` Recent compiler/linter errors
    - `[ ]` Additional related files
- **Task 2.2: Context Builder utility**
  - Collect selected information from the local file system or `go build` output.
  - Package this context into a structured JSON/Markdown payload.
- **Task 2.3: Developer Guidance Input**
  - Add a free-form text input so developers can leave specific instructions for
    Jules.

---

## Epic 3: Dispatch & Agentic Resolution

**Objective:** Send the bundled context to the agent and display the results securely and transparently.

- **Task 3.1: Implement `juleson sessions resolve <ID>`**
  - Create a CLI command to trigger the resolution workflow manually, or hook it
    into the end of `apply` on failure.
  - Dispatch the payload generated from Task 2.2 to the Jules API.
- **Task 3.2: Async waiting UI**
  - Use `charmbracelet/bubbles/spinner` while waiting for Jules to resolve the
    conflict.
- **Task 3.3: Handle the Resolution Response**
  - Parse the returned, clean patch from the agent.
  - Display the `resolution_report.md` artifact so the developer can review how
    the conflict was reconciled.
  - Prompt the developer to preview the new patch or apply it immediately.

---

## Verification & Testing

- **Testing TUI State:** write unit tests for Bubble Tea models using simulated
  key presses.
- **Mocking External Pagers:** verify fallback behavior when `$PATH` does not
  contain `delta`.
- **E2E Conflict Scenario:** create a mock repository conflict, run
  `juleson sessions apply`, trigger the TUI, mock the API response, and verify
  the patch is applied correctly.
