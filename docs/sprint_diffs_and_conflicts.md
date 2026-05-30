# Sprint Plan: Juleson CLI UX Improvements for Diffs and Conflicts

## Overview
This sprint focuses on overhauling the developer experience when dealing with diffs and merge conflicts in the `juleson` CLI. Instead of relying on traditional human-centric tools (like manual 3-way merges or raw text diffs), we are building an agentic-first experience where the developer guides the LLM to resolve conflicts via an interactive TUI.

## Goals
1. **Enhanced Diff Viewing:** Provide structural, syntax-highlighted diffs using external tools like `difftastic`/`delta` or native Go fallbacks.
2. **Context Gathering TUI:** Introduce a sleek interactive terminal wizard to bundle context (local state, patch, compiler output) when a conflict occurs.
3. **Agentic Dispatch:** Route the bundled context back to the `jules` remote session for automatic resolution and surface a detailed `resolution_report.md`.

---

## Epic 1: Better Structured & Formatted Diffs

**Objective:** Upgrade `juleson sessions preview <ID>` to produce readable, syntax-highlighted diffs.

- **Task 1.1: Shell out wrapper for modern diff tools**
  - Implement a check for external diff pagers in the user's `$PATH` (specifically `difftastic` or `delta`).
  - If present, pipe the patch output through the external tool when running `juleson sessions preview`.
- **Task 1.2: Native Go Fallback implementation**
  - Add `github.com/sergi/go-diff` to parse raw diff patches.
  - Integrate `github.com/alecthomas/chroma` for code syntax highlighting.
  - Use `charmbracelet/lipgloss` to render the syntax-highlighted diff gracefully in the terminal.
- **Task 1.3: Add configuration flags**
  - Add flags or global config settings (e.g. `~/.juleson/config.yaml`) to allow users to force the native fallback or specify their preferred diff tool path.

---

## Epic 2: Conflict Context-Gathering UI

**Objective:** When `juleson sessions apply <ID>` encounters a conflict, launch an interactive wizard to gather necessary resolution context.

- **Task 2.1: Implement TUI using `charmbracelet/bubbletea`**
  - Create a new package `internal/tui/conflict` to house the Bubble Tea application logic.
  - Implement a checkbox list allowing the developer to select which context to include:
    - `[x]` Current state of the local file
    - `[x]` The failing patch diff
    - `[ ]` Recent compiler/linter errors
    - `[ ]` Additional related files
- **Task 2.2: Context Builder utility**
  - Write a utility to collect the selected information from the local file system or `go build` output.
  - Package this context into a structured JSON/Markdown payload.
- **Task 2.3: Developer Guidance Input**
  - Add a free-form text input field in the TUI allowing developers to leave specific instructions for the agent (e.g., "Keep my local changes on line 42, but apply the rest").

---

## Epic 3: Dispatch & Agentic Resolution

**Objective:** Send the bundled context to the agent and display the results securely and transparently.

- **Task 3.1: Implement `juleson sessions resolve <ID>`**
  - Create the new CLI command to trigger the resolution workflow manually (or hook it into the end of the `apply` command on failure).
  - Use the payload generated from Task 2.2 and dispatch it to the `jules` remote API.
- **Task 3.2: Async waiting UI**
  - Use `charmbracelet/bubbles/spinner` to show a "Waiting for Jules to resolve conflict..." loading indicator.
- **Task 3.3: Handle the Resolution Response**
  - Parse the returned, clean patch from the agent.
  - Retrieve and display the `resolution_report.md` artifact from the payload to explain to the developer exactly how the conflict was logically reconciled.
  - Prompt the developer to preview the new patch or apply it immediately.

---

## Verification & Testing

- **Testing TUI State:** Write unit tests for the Bubble Tea models using simulated key presses (`tea.KeyMsg`) to ensure selection and inputs work properly.
- **Mocking External Pagers:** Ensure the fallback logic works when `$PATH` doesn't contain `delta`.
- **E2E Conflict Scenario:** Create a mock repository state with an intentional merge conflict, run `juleson sessions apply`, trigger the TUI, mock the API response, and verify the patch is applied correctly.
