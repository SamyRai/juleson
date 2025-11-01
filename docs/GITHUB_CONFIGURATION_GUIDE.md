# GitHub Repository Configuration Guide 101

> **Complete guide to configuring and automating your GitHub repository**
> Last Updated: November 1, 2025

## 📚 Table of Contents

- [Introduction](#introduction)
- [The `.github` Folder Structure](#the-github-folder-structure)
- [Community Health Files](#community-health-files)
- [Issue & Pull Request Templates](#issue--pull-request-templates)
- [GitHub Actions Workflows](#github-actions-workflows)
- [Repository Automation](#repository-automation)
- [Repository Settings](#repository-settings)
- [Branch Protection Rules](#branch-protection-rules)
- [Labels & Milestones](#labels--milestones)
- [Advanced Features](#advanced-features)
- [Best Practices](#best-practices)
- [Quick Reference](#quick-reference)

---

## Introduction

The `.github` folder is a special directory in your repository that GitHub recognizes and uses to configure various aspects of your project. This guide covers everything you can configure to create a professional, automated, and contributor-friendly repository.

### Why Configure Your Repository?

- ✅ **Professionalism** - Show your project is well-maintained
- ✅ **Automation** - Reduce manual work with CI/CD and bots
- ✅ **Collaboration** - Guide contributors with clear templates
- ✅ **Quality** - Enforce standards through branch protection
- ✅ **Community** - Build a welcoming open-source community

---

## The `.github` Folder Structure

```
.github/
├── CODEOWNERS                    # Auto-assign code reviewers
├── FUNDING.yml                   # Sponsor button configuration
├── dependabot.yml                # Automated dependency updates
├── ISSUE_TEMPLATE/               # Issue templates directory
│   ├── bug_report.md            # Bug report template
│   ├── feature_request.md       # Feature request template
│   ├── custom_template.md       # Custom templates
│   └── config.yml               # Issue template chooser config
├── PULL_REQUEST_TEMPLATE.md      # PR template
├── workflows/                    # GitHub Actions workflows
│   ├── ci.yml                   # Continuous Integration
│   ├── codeql.yml               # Security scanning
│   ├── release.yml              # Release automation
│   ├── dependabot-auto-merge.yml # Auto-merge dependencies
│   ├── stale.yml                # Close stale issues
│   └── label-sync.yml           # Sync labels
└── scripts/                      # Custom automation scripts
```

### File Location Options

Many files can be placed in three locations (GitHub checks in this order):

1. `.github/` folder (recommended)
2. Repository root
3. `docs/` folder

---

## Community Health Files

These files help build a healthy, welcoming community around your project.

### 1. CODE_OF_CONDUCT.md

**Purpose:** Define standards for community engagement

**Location:** `.github/CODE_OF_CONDUCT.md` or repository root

**Example:**

```markdown
# Code of Conduct

## Our Pledge

We pledge to make participation in our project a harassment-free experience for everyone.

## Our Standards

Examples of behavior that contributes to a positive environment:
- Using welcoming and inclusive language
- Being respectful of differing viewpoints
- Gracefully accepting constructive criticism

Examples of unacceptable behavior:
- Trolling, insulting/derogatory comments, and personal attacks
- Public or private harassment
- Publishing others' private information

## Enforcement

Instances of abusive behavior may be reported to [email@example.com]
```

**Best Practices:**

- Use the Contributor Covenant template
- Provide clear contact information
- Define consequences for violations

---

### 2. CONTRIBUTING.md

**Purpose:** Guide contributors on how to contribute

**Location:** `.github/CONTRIBUTING.md` or repository root

**Should Include:**

- Getting started (setup, installation)
- Development workflow
- Code style and standards
- Testing requirements
- Pull request process
- Commit message conventions
- Where to ask questions

**Example Structure:**

```markdown
# Contributing Guide

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/repo.git`
3. Install dependencies: `npm install` or `go mod download`
4. Create a branch: `git checkout -b feature/my-feature`

## Development Standards

- Follow existing code style
- Write tests for new features
- Update documentation
- Use conventional commits: `feat:`, `fix:`, `docs:`, `chore:`

## Pull Request Process

1. Update the README.md with details of changes
2. Ensure all tests pass
3. Request review from maintainers
4. Wait for approval before merging

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md)
```

---

### 3. SECURITY.md

**Purpose:** Provide security vulnerability reporting instructions

**Location:** `.github/SECURITY.md` or repository root

**Should Include:**

- Supported versions
- How to report vulnerabilities (email, not public issues)
- Response timeline
- Disclosure policy
- Security best practices

**Example:**

```markdown
# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | ✅                 |
| < 1.0   | ❌                 |

## Reporting a Vulnerability

**Do not** open public issues for security vulnerabilities.

Email: security@example.com

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

## Response Timeline

- Initial response: 48 hours
- Status update: 7 days
- Fix timeline: Varies by severity
```

---

### 4. SUPPORT.md

**Purpose:** Help users get support

**Location:** `.github/SUPPORT.md` or repository root

**Example:**

```markdown
# Getting Help

## Resources

- 📖 [Documentation](https://docs.example.com)
- 💬 [Discussions](https://github.com/user/repo/discussions)
- 🐛 [Report a Bug](https://github.com/user/repo/issues/new?template=bug_report.md)
- 💡 [Request a Feature](https://github.com/user/repo/issues/new?template=feature_request.md)

## Questions?

- Check [existing issues](https://github.com/user/repo/issues)
- Search [discussions](https://github.com/user/repo/discussions)
- Join our [Discord/Slack](https://discord.gg/...)

## Commercial Support

For commercial support, contact: support@example.com
```

---

### 5. FUNDING.yml

**Purpose:** Display sponsor button on repository

**Location:** `.github/FUNDING.yml`

**Example:**

```yaml
# Sponsorship configuration
github: [username1, username2]  # GitHub Sponsors
patreon: username               # Patreon
open_collective: projectname    # Open Collective
ko_fi: username                 # Ko-fi
tidelift: npm/package-name      # Tidelift
custom: ["https://example.com"] # Custom URLs
```

---

### 6. GOVERNANCE.md

**Purpose:** Explain project governance structure

**Location:** `.github/GOVERNANCE.md` or repository root

**Example:**

```markdown
# Project Governance

## Maintainers

- @username1 - Lead Maintainer
- @username2 - Core Contributor

## Decision Making

- Minor changes: Any maintainer can approve
- Major changes: Requires consensus from 2+ maintainers
- Breaking changes: Community RFC process

## Becoming a Maintainer

Consistent high-quality contributions over 6+ months
```

---

## Issue & Pull Request Templates

Templates help contributors provide necessary information and maintain consistency.

### Issue Templates

**Location:** `.github/ISSUE_TEMPLATE/`

#### Bug Report Template

**File:** `.github/ISSUE_TEMPLATE/bug_report.md`

```yaml
---
name: Bug Report
about: Report a bug to help us improve
title: '[BUG] '
labels: bug
assignees: ''
---

## Bug Description
A clear and concise description of what the bug is.

## Steps to Reproduce
1. Go to '...'
2. Click on '...'
3. See error

## Expected Behavior
What you expected to happen.

## Actual Behavior
What actually happened.

## Screenshots
If applicable, add screenshots.

## Environment
- OS: [e.g. macOS, Windows, Linux]
- Version: [e.g. 1.2.3]
- Browser: [e.g. Chrome, Safari]

## Additional Context
Any other context about the problem.
```

#### Feature Request Template

**File:** `.github/ISSUE_TEMPLATE/feature_request.md`

```yaml
---
name: Feature Request
about: Suggest an idea for this project
title: '[FEATURE] '
labels: enhancement
assignees: ''
---

## Problem Statement
What problem does this feature solve?

## Proposed Solution
Describe the solution you'd like.

## Alternatives Considered
What alternatives have you considered?

## Additional Context
Any other context or screenshots.
```

#### Issue Template Chooser Config

**File:** `.github/ISSUE_TEMPLATE/config.yml`

```yaml
blank_issues_enabled: true
contact_links:
  - name: 💬 Discussions
    url: https://github.com/user/repo/discussions
    about: Ask questions and discuss ideas
  - name: 📖 Documentation
    url: https://docs.example.com
    about: Read the documentation
  - name: 💼 Commercial Support
    url: https://example.com/support
    about: Get commercial support
```

---

### Pull Request Template

**Location:** `.github/PULL_REQUEST_TEMPLATE.md`

```markdown
## Description
<!-- Describe your changes in detail -->

## Related Issue
<!-- Link to related issue: Fixes #123 -->

## Type of Change
<!-- Mark with 'x' -->
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## How Has This Been Tested?
<!-- Describe the tests you ran -->

## Checklist
- [ ] My code follows the project's code style
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published

## Screenshots (if applicable)
<!-- Add screenshots to demonstrate changes -->
```

---

## GitHub Actions Workflows

Automate your development workflow with GitHub Actions.

**Location:** `.github/workflows/`

### 1. Continuous Integration (CI)

**File:** `.github/workflows/ci.yml`

```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    name: Test
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go-version: ['1.21', '1.22', '1.23']

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}

      - name: Install dependencies
        run: go mod download

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.txt ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.txt

  lint:
    name: Lint
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [test, lint]

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Build
        run: go build -v ./...
```

---

### 2. Security Scanning (CodeQL)

**File:** `.github/workflows/codeql.yml`

```yaml
name: CodeQL

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 0 * * 1'  # Weekly on Monday

jobs:
  analyze:
    name: Analyze
    runs-on: ubuntu-latest
    permissions:
      actions: read
      contents: read
      security-events: write

    strategy:
      fail-fast: false
      matrix:
        language: ['go', 'javascript']

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Initialize CodeQL
        uses: github/codeql-action/init@v3
        with:
          languages: ${{ matrix.language }}

      - name: Autobuild
        uses: github/codeql-action/autobuild@v3

      - name: Perform CodeQL Analysis
        uses: github/codeql-action/analyze@v3
```

---

### 3. Release Automation

**File:** `.github/workflows/release.yml`

```yaml
name: Release

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  release:
    name: Create Release
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

### 4. Dependabot Auto-Merge

**File:** `.github/workflows/dependabot-auto-merge.yml`

```yaml
name: Dependabot Auto-Merge

on:
  pull_request:
    types: [opened, synchronize, reopened]

permissions:
  contents: write
  pull-requests: write

jobs:
  auto-merge:
    name: Auto-merge Dependabot PRs
    runs-on: ubuntu-latest
    if: github.actor == 'dependabot[bot]'

    steps:
      - name: Dependabot metadata
        id: metadata
        uses: dependabot/fetch-metadata@v2
        with:
          github-token: "${{ secrets.GITHUB_TOKEN }}"

      - name: Auto-merge minor updates
        if: steps.metadata.outputs.update-type == 'version-update:semver-minor' || steps.metadata.outputs.update-type == 'version-update:semver-patch'
        run: gh pr merge --auto --merge "$PR_URL"
        env:
          PR_URL: ${{ github.event.pull_request.html_url }}
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

### 5. Stale Issues/PRs Management

**File:** `.github/workflows/stale.yml`

```yaml
name: Stale Issues

on:
  schedule:
    - cron: '0 0 * * *'  # Daily
  workflow_dispatch:

jobs:
  stale:
    name: Close stale issues
    runs-on: ubuntu-latest

    steps:
      - uses: actions/stale@v9
        with:
          repo-token: ${{ secrets.GITHUB_TOKEN }}
          stale-issue-message: 'This issue is stale because it has been open 60 days with no activity. Remove stale label or comment or this will be closed in 7 days.'
          stale-pr-message: 'This PR is stale because it has been open 45 days with no activity. Remove stale label or comment or this will be closed in 7 days.'
          close-issue-message: 'This issue was closed because it has been stalled for 7 days with no activity.'
          close-pr-message: 'This PR was closed because it has been stalled for 7 days with no activity.'
          days-before-stale: 60
          days-before-close: 7
          stale-issue-label: 'stale'
          stale-pr-label: 'stale'
          exempt-issue-labels: 'pinned,security,roadmap'
          exempt-pr-labels: 'pinned,security'
```

---

## Repository Automation

### CODEOWNERS

**Purpose:** Auto-assign reviewers for specific files/directories

**Location:** `.github/CODEOWNERS`

**Example:**

```
# Default owners for everything
* @org/core-team

# Backend code
/internal/      @org/backend-team
/cmd/           @org/backend-team

# Frontend code
/web/           @org/frontend-team
/ui/            @org/frontend-team

# Documentation
*.md            @org/docs-team
/docs/          @org/docs-team

# CI/CD
/.github/       @org/devops-team
Makefile        @org/devops-team

# Security
SECURITY.md     @org/security-team
```

---

### Dependabot Configuration

**Purpose:** Automated dependency updates

**Location:** `.github/dependabot.yml`

```yaml
version: 2

updates:
  # Go modules
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "09:00"
    open-pull-requests-limit: 10
    reviewers:
      - "org/backend-team"
    labels:
      - "dependencies"
      - "go"
    commit-message:
      prefix: "chore"
      include: "scope"

  # GitHub Actions
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
    reviewers:
      - "org/devops-team"
    labels:
      - "dependencies"
      - "github-actions"

  # npm (if you have frontend)
  - package-ecosystem: "npm"
    directory: "/web"
    schedule:
      interval: "weekly"
    reviewers:
      - "org/frontend-team"
    labels:
      - "dependencies"
      - "npm"
```

---

## Repository Settings

These are configured through the GitHub UI but documented here for reference.

### General Settings

**Navigation:** Settings → General

- **Repository name** - Clear, descriptive name
- **Description** - Short description (shown in search results)
- **Website** - Documentation or project homepage
- **Topics** - Tags for discoverability (e.g., `golang`, `cli`, `automation`)
- **Include in GitHub Explore** - Make discoverable
- **Visibility** - Public or Private
- **Features**:
  - ✅ Issues
  - ✅ Discussions (for Q&A and community)
  - ✅ Projects (for roadmaps)
  - ✅ Wiki (optional)
  - ✅ Preserve this repository (archive on delete)

### Pull Requests

**Navigation:** Settings → Pull Requests

- ✅ **Allow merge commits** - For preserving history
- ✅ **Allow squash merging** - For clean history (recommended)
- ✅ **Allow rebase merging** - For linear history
- ✅ **Automatically delete head branches** - Clean up after merge
- ✅ **Allow auto-merge** - Enable auto-merge when checks pass
- ✅ **Suggest updating pull request branches** - Keep PRs up-to-date

---

## Branch Protection Rules

**Purpose:** Enforce quality standards before merging

**Navigation:** Settings → Branches → Add rule

### Recommended Protection for `main` Branch

```
Branch name pattern: main

✅ Require pull request before merging
   ✅ Require approvals: 1 (or 2 for critical projects)
   ✅ Dismiss stale pull request approvals when new commits are pushed
   ✅ Require review from Code Owners

✅ Require status checks to pass before merging
   ✅ Require branches to be up to date before merging
   Status checks required:
   - test / Test (ubuntu-latest, 1.23)
   - test / Test (macos-latest, 1.23)
   - lint / Lint
   - build / Build

✅ Require conversation resolution before merging

✅ Require signed commits (optional, for extra security)

✅ Require linear history (recommended for clean history)

✅ Do not allow bypassing the above settings
   (Applies to administrators too)

❌ Allow force pushes (disabled)

❌ Allow deletions (disabled)
```

### Branch Protection Levels

**Level 1 - Basic (Solo/Small Projects):**

- Require pull requests
- Require 1 approval
- Require CI to pass

**Level 2 - Standard (Team Projects):**

- Everything in Level 1
- Require Code Owner review
- Require conversation resolution
- Require linear history

**Level 3 - Strict (Critical Projects):**

- Everything in Level 2
- Require 2+ approvals
- Require signed commits
- Dismiss stale reviews
- Apply to administrators
- Require deployment to staging

---

## Labels & Milestones

### Default GitHub Labels

GitHub provides these default labels:

| Label | Description |
|-------|-------------|
| `bug` | Something isn't working |
| `documentation` | Improvements to docs |
| `duplicate` | Duplicate issue |
| `enhancement` | New feature or request |
| `good first issue` | Good for newcomers |
| `help wanted` | Extra attention needed |
| `invalid` | Doesn't seem right |
| `question` | Further information requested |
| `wontfix` | Won't be worked on |

### Recommended Additional Labels

**Type Labels:**

- `type: bug` 🐛 - Bug reports
- `type: feature` ✨ - New features
- `type: refactor` 🔨 - Code refactoring
- `type: docs` 📚 - Documentation
- `type: test` 🧪 - Testing
- `type: chore` 🧹 - Maintenance tasks

**Priority Labels:**

- `priority: critical` 🔴 - Needs immediate attention
- `priority: high` 🟠 - Important
- `priority: medium` 🟡 - Normal priority
- `priority: low` 🟢 - Nice to have

**Status Labels:**

- `status: needs-triage` - Needs review
- `status: blocked` - Blocked by dependency
- `status: in-progress` - Being worked on
- `status: needs-review` - Awaiting code review
- `status: ready-to-merge` - Approved and ready

**Size Labels:**

- `size: XS` - < 10 lines changed
- `size: S` - 10-50 lines
- `size: M` - 50-200 lines
- `size: L` - 200-500 lines
- `size: XL` - > 500 lines

### Creating Labels via Script

```bash
#!/bin/bash
# create-labels.sh

# Type labels
gh label create "type: bug" --color d73a4a --description "Bug reports"
gh label create "type: feature" --color 0e8a16 --description "New features"
gh label create "type: refactor" --color fbca04 --description "Code refactoring"

# Priority labels
gh label create "priority: critical" --color b60205 --description "Needs immediate attention"
gh label create "priority: high" --color ff9800 --description "Important"
gh label create "priority: medium" --color ffc107 --description "Normal priority"
gh label create "priority: low" --color 4caf50 --description "Nice to have"
```

---

## Advanced Features

### 1. GitHub Discussions

**Purpose:** Community Q&A, announcements, and conversations

**Enable:** Settings → Features → Discussions

**Categories to Create:**

- 📣 **Announcements** - Project updates
- 💡 **Ideas** - Feature discussions
- 🙏 **Q&A** - Community questions
- 💬 **General** - General discussions
- 🎉 **Show and Tell** - Community showcases

---

### 2. GitHub Projects

**Purpose:** Project management and roadmaps

**Create:** Projects tab → New project

**Views:**

- **Roadmap** - Timeline view
- **Board** - Kanban board
- **Table** - Spreadsheet view

**Automation:**

- Auto-add items
- Auto-move based on status
- Close linked issues

---

### 3. Repository Insights

**Navigation:** Insights tab

**Monitor:**

- **Pulse** - Recent activity
- **Contributors** - Contribution stats
- **Traffic** - Views and clones
- **Commits** - Commit activity
- **Code frequency** - Additions/deletions
- **Dependency graph** - Dependencies
- **Network** - Forks and branches

---

### 4. Social Preview Image

**Purpose:** Custom image for social media shares

**Configure:** Settings → Social preview → Upload image

**Recommendations:**

- Size: 1280x640 px
- Format: PNG or JPG
- Include: Logo + project name
- Keep text readable

---

### 5. Citation File

**Purpose:** Make your project citable in academic work

**Location:** `CITATION.cff` (repository root)

```yaml
cff-version: 1.2.0
message: "If you use this software, please cite it as below."
authors:
  - family-names: "Doe"
    given-names: "John"
    orcid: "https://orcid.org/0000-0000-0000-0000"
title: "My Project"
version: 1.0.0
doi: 10.5281/zenodo.1234567
date-released: 2025-11-01
url: "https://github.com/user/project"
```

---

## Best Practices

### Documentation

✅ **Do:**

- Keep README concise but comprehensive
- Include quick start guide
- Add code examples
- Link to detailed documentation
- Keep documentation up-to-date

❌ **Don't:**

- Write walls of text
- Assume prior knowledge
- Leave broken links
- Skip examples

---

### Automation

✅ **Do:**

- Automate repetitive tasks
- Use CI/CD for testing
- Auto-merge safe dependencies
- Close stale issues automatically
- Use bots for labeling

❌ **Don't:**

- Over-automate (keep human oversight)
- Auto-merge without tests
- Close issues too aggressively
- Spam contributors with bot messages

---

### Community

✅ **Do:**

- Respond to issues promptly
- Welcome new contributors
- Acknowledge contributions
- Maintain Code of Conduct
- Celebrate milestones

❌ **Don't:**

- Ignore community input
- Be dismissive of contributions
- Let issues pile up unanswered
- Tolerate harassment

---

### Security

✅ **Do:**

- Enable Dependabot
- Run CodeQL scans
- Require signed commits (critical projects)
- Review dependency updates
- Keep secrets out of repository

❌ **Don't:**

- Commit API keys or passwords
- Ignore security advisories
- Skip security reviews
- Auto-merge major updates blindly

---

## Quick Reference

### Essential Files Checklist

```
Repository Root:
✅ README.md
✅ LICENSE
✅ .gitignore
✅ CHANGELOG.md

.github/:
✅ CODE_OF_CONDUCT.md
✅ CONTRIBUTING.md
✅ SECURITY.md
✅ SUPPORT.md (optional)
✅ FUNDING.yml (optional)
✅ CODEOWNERS
✅ dependabot.yml

.github/ISSUE_TEMPLATE/:
✅ bug_report.md
✅ feature_request.md
✅ config.yml

.github/:
✅ PULL_REQUEST_TEMPLATE.md

.github/workflows/:
✅ ci.yml
✅ codeql.yml
✅ release.yml (optional)
✅ dependabot-auto-merge.yml (optional)
✅ stale.yml (optional)
```

---

### Common Workflow Triggers

```yaml
# On push to specific branches
on:
  push:
    branches: [main, develop]

# On pull request
on:
  pull_request:
    branches: [main]

# On tag creation
on:
  push:
    tags:
      - 'v*.*.*'

# On schedule (cron)
on:
  schedule:
    - cron: '0 0 * * 1'  # Weekly on Monday

# Manual trigger
on:
  workflow_dispatch:

# Multiple events
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 0 * * 1'
```

---

### GitHub CLI Commands

```bash
# Create labels
gh label create "bug" --color d73a4a --description "Bug reports"

# List labels
gh label list

# Create issue
gh issue create --title "Bug report" --body "Description" --label bug

# List pull requests
gh pr list

# Create release
gh release create v1.0.0 --title "Version 1.0.0" --notes "Release notes"

# Enable discussions
gh repo edit --enable-discussions

# Enable vulnerability alerts
gh repo edit --enable-vulnerability-alerts
```

---

### Useful GitHub Actions

```yaml
# Popular Actions to Use

# Checkout code
- uses: actions/checkout@v4

# Setup programming languages
- uses: actions/setup-go@v5
- uses: actions/setup-node@v4
- uses: actions/setup-python@v5

# Caching
- uses: actions/cache@v4

# Code coverage
- uses: codecov/codecov-action@v4

# Security scanning
- uses: github/codeql-action/init@v3

# Release automation
- uses: goreleaser/goreleaser-action@v5

# Dependabot
- uses: dependabot/fetch-metadata@v2
```

---

## Resources

### Official Documentation

- [GitHub Docs](https://docs.github.com)
- [GitHub Actions](https://docs.github.com/en/actions)
- [Community Health Files](https://docs.github.com/en/communities/setting-up-your-project-for-healthy-contributions)
- [Branch Protection](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches)

### Tools

- [GitHub CLI](https://cli.github.com/)
- [Dependabot](https://github.com/dependabot)
- [CodeQL](https://codeql.github.com/)
- [Probot](https://probot.github.io/) - GitHub App framework

### Templates & Examples

- [Awesome README](https://github.com/matiassingers/awesome-readme)
- [Contributor Covenant](https://www.contributor-covenant.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Semantic Versioning](https://semver.org/)

---

## Conclusion

A well-configured GitHub repository:

- 🎯 Attracts contributors
- 🤖 Automates workflows
- 🔒 Maintains quality
- 🚀 Accelerates development
- 🌟 Builds community

Start with the essentials (README, LICENSE, CONTRIBUTING) and gradually add automation and advanced features as your project grows.

---

**Last Updated:** November 1, 2025
**License:** MIT
**Maintained By:** Juleson Team
