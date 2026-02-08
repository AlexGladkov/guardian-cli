# Guardian CLI — Technical Specification

## 1. Overview

**Guardian** is a CLI tool written in Go that implements a "constitutional engine" for development teams. It stores team agreements (rules) in a git repository, checks code changes against those rules, and manages rule changes through proposals and distributed voting governed by a constitution.

**Key distinction:** Guardian is NOT CODEOWNERS or file-permission management. Rules check code correctness (imports, patterns, conventions), and voting governs changes to rules — not changes to code.

### Core Principles

- **Serverless by default:** No server, no accounts, no tokens (beyond LLM API key).
- **Git as transport and audit:** Proposals and votes are files in the repository; git history provides audit trail.
- **Identity:** git `user.email` maps to roles. Enterprise can enhance with signed commits (future).
- **LLM is always on:** LLM analyzes every check, explains violations, drafts proposals. LLM provider/URL is configured in constitution.yml (shared); API key from environment variable.
- **Minimal code access:** Uses `git diff --name-only` and `git diff` by default. Full file reads only via explicit allowlist (TODO for future).

---

## 2. Technology Stack

| Component        | Choice                              |
|------------------|-------------------------------------|
| Language         | Go 1.22+                           |
| CLI framework    | cobra                               |
| Config format    | YAML (gopkg.in/yaml.v3)            |
| Tests            | standard `testing` + testify        |
| Lint/format      | gofmt, golangci-lint                |
| Build            | goreleaser, Makefile                |
| Module path      | `github.com/AlexGladkov/guardian-cli` |
| Binary name      | `guardian`                          |
| Distribution     | go install + GitHub releases + Homebrew + Docker |

---

## 3. Repository Structure

```
guardian-cli/
├── cmd/
│   └── guardian/
│       └── main.go
├── internal/
│   ├── cli/                    # Cobra commands
│   │   ├── root.go
│   │   ├── init.go
│   │   ├── check.go
│   │   ├── propose.go
│   │   ├── vote.go
│   │   ├── tally.go
│   │   ├── finalize.go
│   │   ├── withdraw.go
│   │   ├── inbox.go
│   │   ├── hooks.go
│   │   ├── history.go
│   │   ├── exception.go
│   │   └── llm_configure.go
│   ├── config/                 # YAML parsing and validation
│   │   ├── constitution.go
│   │   ├── rules.go
│   │   ├── proposal.go
│   │   ├── vote.go
│   │   ├── exception.go
│   │   └── validation.go
│   ├── engine/                 # Rule checking engine
│   │   ├── checker.go          # RuleChecker interface + registry
│   │   ├── imports_forbidden.go
│   │   ├── diff_pattern_forbidden.go
│   │   ├── diff_pattern_requires.go
│   │   ├── meta_check.go       # Detects unauthorized .agreements/ changes
│   │   └── engine.go           # Orchestrator
│   ├── governance/             # Voting and quorum logic
│   │   ├── tally.go
│   │   ├── quorum.go
│   │   └── roles.go
│   ├── git/                    # Git operations
│   │   ├── diff.go
│   │   ├── identity.go
│   │   ├── hooks.go
│   │   └── ci.go               # CI environment detection
│   ├── llm/                    # LLM integration
│   │   ├── client.go
│   │   ├── providers.go        # DeepSeek, OpenAI, Claude, Custom
│   │   ├── prompts.go          # Built-in prompts
│   │   └── configure.go
│   ├── inbox/                  # Inbox logic
│   │   ├── inbox.go
│   │   ├── notify.go           # OS notifications
│   │   └── state.go            # state.json management
│   ├── discovery/              # .agreements/ root discovery
│   │   └── discovery.go
│   └── output/                 # Output formatting
│       ├── human.go
│       └── json.go
├── examples/
│   └── .agreements/
│       ├── constitution.yml
│       ├── rules.yml
│       └── exceptions/
├── .github/
│   └── workflows/
│       └── ci.yml
├── .golangci.yml
├── .gitignore
├── .goreleaser.yml
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
├── README.md
├── LICENSE
└── SPEC.md
```

---

## 4. Data Files (.agreements/)

Guardian uses a `.agreements/` directory in the target project's repository.

```
.agreements/
├── constitution.yml
├── rules.yml
├── proposals/
│   └── <date>-<rule_id>.yml
├── votes/
│   └── <proposal_id>/
│       └── <voter_email>.yml
├── history/
│   └── <proposal_id>.md
└── exceptions/
    └── <exception_id>.yml
```

### 4.1. constitution.yml

```yaml
governance:
  voters:
    - role: techlead
    - role: architect
    - role: product
  quorum:
    type: two_thirds        # majority | two_thirds | unanimous | custom
    threshold: 0.66         # used for custom, or as override
  forbid_self_approval: true  # configurable: true = author cannot vote yes on own proposal
  allow_vote_change: false    # configurable: whether voters can change their vote before finalize
  proposal_ttl_days: 30       # proposals expire after N days; status becomes "expired"
  per_rule_overrides:
    payment_state_machine:
      quorum:
        type: unanimous
  exceptions:
    require_approval: false   # configurable: if true, exception requires mini-proposal (1 approve)
    # if false, exception is created by anyone and goes through code review

identity:
  allowed_domains: ["company.com"]   # optional
  require_signed_commits: false      # MVP: false

roles:
  techlead:
    members:
      - email: ivan@company.com
  architect:
    members:
      - email: maria@company.com
  product:
    members:
      - email: alex@company.com

llm:
  provider: deepseek          # deepseek | openai | claude | custom
  endpoint: ""                # required for custom; auto-set for known providers
  model: ""                   # optional model override
  # API key is NEVER stored here; taken from env: GUARDIAN_LLM_API_KEY
  prompts:
    check_system: ""          # optional override for built-in check prompt
    propose_system: ""        # optional override for built-in propose prompt
```

#### Quorum Types

| Type        | Logic                                                |
|-------------|------------------------------------------------------|
| `majority`  | > 50% of eligible voters vote yes                    |
| `two_thirds`| >= 66.7% of eligible voters vote yes                 |
| `unanimous` | 100% of eligible voters vote yes                     |
| `custom`    | >= `threshold` fraction of eligible voters vote yes  |

**Quorum base:** By unique people (emails), NOT by roles. If a person has multiple roles, they still count as one voter (deduplicated by email).

**Abstain:** Removed. Only `yes` or `no` votes are allowed.

**Self-vote:** Configurable via `forbid_self_approval`:
- `true` — author cannot vote `yes` on their own proposal (can vote `no`)
- `false` — author can vote freely

**Vote mutability:** Configurable via `allow_vote_change`:
- `true` — repeated `guardian vote` overwrites the vote file; git history preserves audit
- `false` — error if vote file already exists

### 4.2. rules.yml

```yaml
rules:
  - id: domain_no_infra
    description: Domain layer must not depend on infra
    type: imports_forbidden
    config:
      from_globs: ["domain/**"]
      forbid_globs: ["infra/**"]
    severity: error

  - id: money_minor_units
    description: Money must use int minor units, not float/double
    type: diff_pattern_forbidden
    config:
      forbidden_regexes:
        - "\\bDouble\\b"
        - "\\bfloat\\b"
      only_in_paths: ["**/*.kt", "**/*.java", "**/*.ts"]
    severity: warning

  - id: public_api_stability
    description: Public API changes require RFC tag in proposal
    type: diff_pattern_requires
    config:
      required_regexes:
        - "RFC:"
      only_in_paths: ["sdk/public/**"]
    severity: error
```

Rules are extensible via `RuleChecker` interface + registry by `type`.

### 4.3. Proposal File

`.agreements/proposals/<date>-<rule_id>.yml`

```yaml
id: "2024-01-15-domain_no_infra"
rule_id: domain_no_infra
proposal_type: modify          # modify | add | remove
change:
  description: "Allow infra imports in domain/adapters/"
  details: "..."               # freeform text
reason: "Domain adapters need to implement infra interfaces"
impact: "domain/adapters/ files can now import from infra/"
created_by: ivan@company.com
created_at: "2024-01-15T10:30:00Z"
status: proposed               # proposed | accepted | rejected | withdrawn | expired
```

**Constraints:**
- Only one active proposal per rule_id at a time. Creating a second proposal for the same rule while one is active (status: proposed) is blocked with an error.
- Proposal types: `modify` (change existing rule), `add` (new rule), `remove` (delete rule).

### 4.4. Vote File

`.agreements/votes/<proposal_id>/<voter_email>.yml`

```yaml
proposal_id: "2024-01-15-domain_no_infra"
voter_email: maria@company.com
decision: yes                  # yes | no
comment: "Makes sense for adapter pattern"
voted_at: "2024-01-16T14:00:00Z"
```

### 4.5. Exception File

`.agreements/exceptions/<exception_id>.yml`

```yaml
id: "exc-2024-01-20-domain_no_infra"
rule_id: domain_no_infra
paths:
  - "domain/legacy/old_adapter.kt"
reason: "Legacy code, will be migrated in Q2"
created_by: ivan@company.com
created_at: "2024-01-20T09:00:00Z"
expires_at: "2024-06-30T00:00:00Z"   # optional; if omitted, exception is permanent
```

**Exception ACL:** Configurable in `constitution.yml` via `governance.exceptions.require_approval`:
- `false` — anyone creates, goes through normal code review
- `true` — requires simplified approval (at least 1 voter approves)

**Expiry:** Expired exceptions are ignored by `guardian check`.

### 4.6. History File

`.agreements/history/<proposal_id>.md`

Generated by `guardian finalize`. Contains:
- Proposal summary (id, rule_id, proposal_type, change description)
- Vote tally (who voted what, timestamps)
- Result: ACCEPTED
- Finalized by (email) and timestamp

---

## 5. CLI Commands

### 5.1. `guardian init`

Creates `.agreements/` directory with template files.

- Creates: `constitution.yml`, `rules.yml`, `proposals/`, `votes/`, `history/`, `exceptions/`
- Does NOT overwrite existing files unless `--force` flag is used
- After creating files, automatically launches `guardian llm configure` if LLM is not yet configured
- Exit codes: 0 success, 2 error

### 5.2. `guardian check [<base>..<head>]`

Checks code changes against all rules.

**Diff range resolution (priority order):**
1. Explicit argument: `guardian check main~5..HEAD`
2. CI auto-detect: reads `$GITHUB_BASE_REF` / `$CI_MERGE_REQUEST_TARGET_BRANCH_NAME` etc.
3. Default: `origin/main..HEAD`
4. If diff is empty: show hint message with example `guardian check HEAD~3..HEAD` and exit 0

**Process:**
1. Collect changed files via `git diff --name-only <range>`
2. Collect diff content via `git diff <range>`
3. Run all rules from `rules.yml` (regex-based checkers)
4. **Meta-check:** detect unauthorized changes to `.agreements/` files (constitution.yml, rules.yml) without a corresponding accepted proposal — this is a violation
5. Apply exceptions: skip violations for paths covered by non-expired exceptions
6. Send diff + rule descriptions + violations to LLM for analysis and explanation
7. Print report

**Output format:**
- Human-friendly by default (rule id, severity, explanation, path, short diff snippet — first N lines, NOT full diff)
- `--json` flag for machine-readable output

**Exit codes:**
- 0: OK (or only warnings)
- 1: violations found
- 2: config error / runtime error (including LLM unavailable)

**LLM behavior:**
- LLM is called on every check
- If LLM is unavailable (no key, timeout, error): exit code 2 (fail)
- Regex rules always run; LLM adds explanations and recommendations to violations
- LLM does NOT make decisions — regex rules determine pass/fail; LLM provides context

### 5.3. `guardian propose <rule_id>`

Creates a proposal file.

- `rule_id` is required. For `add` type, user provides a new rule_id
- Validates that no other active proposal exists for the same rule_id (for `modify` and `remove`)
- Interactive mode: prompts in terminal (standard input, no external libs)
  - Proposal type: modify / add / remove
  - Change description
  - Reason
  - Impact
- `--llm` flag: uses LLM to generate draft text based on rule and context
- Creates file: `.agreements/proposals/<date>-<rule_id>.yml`
- Does NOT auto-commit; shows `git add` / `git commit` hint

### 5.4. `guardian vote <proposal_id> --yes|--no [--comment "..."]`

Records a vote.

- Creates file: `.agreements/votes/<proposal_id>/<voter_email>.yml`
- Voter email determined from `git config user.email`
- **Validation:**
  - Voter must belong to one of the roles in `governance.voters`
  - `forbid_self_approval`: if proposal `created_by` == voter email AND vote is `yes` — error
  - Vote already exists: check `allow_vote_change` in constitution
    - `true`: overwrite file
    - `false`: error "You have already voted"
- Does NOT auto-commit; shows `git add` / `git commit` hint

### 5.5. `guardian tally <proposal_id>`

Calculates and displays voting results.

- Reads proposal + all vote files
- Computes quorum based on constitution (with per-rule override support)
- Checks proposal TTL (if `proposal_ttl_days` set and exceeded — status: expired)
- Displays: required roles, eligible voters (unique emails), current votes, result

**Result states:**
- `ACCEPTED`: quorum reached with sufficient yes votes
- `REJECTED`: enough no votes that quorum for yes is impossible
- `PENDING`: voting still in progress
- `EXPIRED`: TTL exceeded

**Flags:** `--json`

### 5.6. `guardian finalize <proposal_id>`

Finalizes an accepted proposal.

- **Who can finalize:** any person with a role in `governance.voters`
- Runs `tally` internally — only proceeds if result is ACCEPTED
- Actions:
  1. Updates proposal status to `accepted`
  2. Creates history file `.agreements/history/<proposal_id>.md`
  3. Prints instructions to user: what to change in `rules.yml` (based on proposal change description)
  4. Does NOT auto-modify rules.yml
- Exit codes: 0 success, 1 not accepted / error

### 5.7. `guardian withdraw <proposal_id>`

Withdraws a proposal (author only).

- Only the proposal author (`created_by`) can withdraw
- Sets status to `withdrawn`
- Moves/marks proposal as closed
- Does NOT auto-commit; shows hint

### 5.8. `guardian inbox [--notify] [--since-last-check] [--quiet] [--no-fetch]`

Shows proposals that need the current user's vote.

**Process:**
1. `git fetch` (unless `--no-fetch`)
2. Find proposals with status `proposed` (not expired)
3. Determine current user via `git config user.email`
4. Determine user's roles
5. Filter: proposals where user is eligible to vote and hasn't voted yet
6. Display list (with proposal age highlighted for old proposals)

**Flags:**
- `--notify`: OS notification
  - macOS: `terminal-notifier` if available, else stdout
  - Linux: `notify-send` if available, else stdout
  - Windows: stdout (TODO: PowerShell toast)
- `--since-last-check`: use `.guardian/state.json` (local, NOT committed) to filter new proposals
- `--quiet`: minimal output (for hooks)
- `--no-fetch`: skip git fetch
- `--json`: machine-readable output

### 5.9. `guardian hooks install|uninstall`

Manages git hooks.

**install:**
- Creates `.git/hooks/post-merge` and `.git/hooks/post-checkout`
- Hook content: launches `guardian inbox --notify --since-last-check --quiet` **in background** (`& disown`), does not block git operations
- Each hook file includes a marker comment: `# GUARDIAN-MANAGED-HOOK`

**uninstall:**
- Removes only hooks with the `# GUARDIAN-MANAGED-HOOK` marker

### 5.10. `guardian llm configure`

Interactive LLM setup.

- Creates/updates LLM config in `constitution.yml` (shared) and stores API key guidance
- Interactive flow:
  1. Select provider: DeepSeek / OpenAI / Claude (Anthropic) / Custom (OpenAI-compatible)
  2. For custom: enter endpoint URL
  3. Show instruction: "Set environment variable GUARDIAN_LLM_API_KEY=<your-key>"
  4. If cloud provider selected: show privacy warning + suggest custom/local endpoint for confidential code
- API key is NEVER stored in files — only via env var `GUARDIAN_LLM_API_KEY`

### 5.11. `guardian history`

Displays history of accepted/finalized proposals.

- Reads all files from `.agreements/history/`
- Displays: proposal_id, rule_id, proposal_type, date finalized, summary
- `--json` flag for machine output

### 5.12. `guardian exception create <rule_id>`

Creates an exception file.

- Interactive prompts:
  - Paths (glob patterns or specific files)
  - Reason
  - Expires at (optional date, or empty for permanent)
- Creates file: `.agreements/exceptions/<exception_id>.yml`
- If `governance.exceptions.require_approval` is true: shows warning that exception needs approval
- Does NOT auto-commit; shows hint

---

## 6. Rule Engine

### 6.1. RuleChecker Interface

```go
type RuleChecker interface {
    Check(ctx CheckContext) ([]Violation, error)
    Type() string
}

type CheckContext struct {
    ChangedFiles []string
    DiffContent  string       // full unified diff
    RuleConfig   map[string]interface{}
    Severity     string
    RuleID       string
}

type Violation struct {
    RuleID      string
    Severity    string   // error | warning
    Description string
    FilePath    string
    DiffSnippet string   // first N lines of relevant diff
    LLMExplanation string // populated after LLM analysis
}
```

Checkers are registered in a registry by `type` string.

### 6.2. imports_forbidden

- Matches changed files against `from_globs`
- For matching files, searches `+` lines in diff for patterns from `forbid_globs`
- Simple pattern match: checks if forbidden glob path segments appear in added lines
- LLM provides additional context and reduces false positives in its explanation

### 6.3. diff_pattern_forbidden

- Optionally filters by `only_in_paths` (glob match on changed files)
- Applies `forbidden_regexes` to `+` lines in diff
- Reports each match as a violation

### 6.4. diff_pattern_requires

- If changed files match `only_in_paths`
- Requires at least one `required_regexes` pattern to be present somewhere in the diff
- If not found — violation

### 6.5. meta_check (built-in, always active)

- Detects changes to `.agreements/constitution.yml` or `.agreements/rules.yml` in the diff
- If changes found — checks if there's a corresponding accepted proposal
- If no accepted proposal — violation (severity: error)

---

## 7. LLM Integration

### 7.1. Providers

| Provider | Endpoint                            | Env var             |
|----------|-------------------------------------|---------------------|
| deepseek | `https://api.deepseek.com/v1`      | GUARDIAN_LLM_API_KEY |
| openai   | `https://api.openai.com/v1`        | GUARDIAN_LLM_API_KEY |
| claude   | `https://api.anthropic.com/v1`     | GUARDIAN_LLM_API_KEY |
| custom   | User-provided URL (OpenAI-compatible) | GUARDIAN_LLM_API_KEY |

All providers use OpenAI-compatible API format, except Claude which uses Anthropic Messages API.

### 7.2. Usage Points

1. **`guardian check`** — every run:
   - Sends: diff content + rule descriptions + regex-detected violations
   - Receives: explanation of violations, recommendations, false-positive assessment
   - If LLM unavailable: exit code 2

2. **`guardian propose --llm`** — optional:
   - Sends: rule description + context
   - Receives: draft proposal text (change, reason, impact)

### 7.3. Prompts

Built-in prompts are hardcoded in `internal/llm/prompts.go`.
Can be overridden via `llm.prompts.check_system` and `llm.prompts.propose_system` in `constitution.yml`.

### 7.4. Privacy

- When user selects a cloud provider (deepseek/openai/claude): show warning that code diffs will be sent to external API
- Suggest using `custom` provider with local endpoint (e.g., Ollama) for confidential code
- Warning shown during `guardian llm configure`

---

## 8. Root Discovery

Guardian searches for `.agreements/` directory starting from CWD upward to filesystem root (similar to how git searches for `.git/`). This ensures it works when invoked from any subdirectory.

---

## 9. CI Integration

### 9.1. Auto-detection

`guardian check` auto-detects CI environments and extracts base/head refs:

| CI System       | Base Branch Variable           |
|-----------------|--------------------------------|
| GitHub Actions  | `GITHUB_BASE_REF`             |
| GitLab CI       | `CI_MERGE_REQUEST_TARGET_BRANCH_NAME` |
| Other           | Falls back to `origin/main..HEAD` |

### 9.2. GitHub Actions Workflow (for guardian itself)

```yaml
name: CI
on: [push, pull_request]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: make lint
      - run: make test
      - run: make build
```

---

## 10. Local State

`.guardian/state.json` — stored locally, NOT committed.

```json
{
  "last_inbox_check": "2024-01-16T14:00:00Z"
}
```

Used by `guardian inbox --since-last-check` to filter only new proposals.

---

## 11. Testing Strategy

**Coverage: all components** (as specified).

| Component                  | Test Type        | Key Scenarios                                        |
|----------------------------|------------------|------------------------------------------------------|
| governance/tally           | Unit             | majority, two_thirds, unanimous, custom threshold    |
| governance/quorum          | Unit             | deduplication by email, per-rule overrides            |
| engine/imports_forbidden   | Unit             | matching, non-matching, edge cases                   |
| engine/diff_pattern_*      | Unit             | regex matching, path filtering                       |
| engine/meta_check          | Unit             | unauthorized .agreements/ changes                    |
| inbox/inbox                | Unit             | needs-my-vote filtering, role matching               |
| config/validation          | Unit             | valid/invalid constitution, rules, proposals, votes  |
| config/constitution        | Unit             | parsing, defaults, per-rule overrides                |
| discovery/discovery        | Unit             | traversal up directory tree                          |
| git/ci                     | Unit             | CI env detection for GitHub, GitLab                  |
| governance/roles           | Unit             | multi-role dedup, email matching                     |
| cli/*                      | Integration      | Command execution with test fixtures                 |
| output/*                   | Unit             | Human and JSON formatting                            |
| inbox/state                | Unit             | state.json read/write, since-last-check              |
| exception handling         | Unit             | expiry logic, path matching                          |
| proposal lifecycle         | Integration      | propose -> vote -> tally -> finalize                 |
| proposal TTL               | Unit             | expiry calculation                                   |
| withdraw                   | Unit             | author-only check                                    |

---

## 12. Output Format

### 12.1. Human-readable (default)

```
Guardian Check Report
=====================

VIOLATION [error] domain_no_infra
  Domain layer must not depend on infra
  File: domain/service/UserService.kt
  Diff:
    + import com.myapp.infra.database.UserRepository
  AI: This import creates a direct dependency from domain to infrastructure layer.
      Consider using a domain interface instead.

WARNING [warning] money_minor_units
  Money must use int minor units, not float/double
  File: domain/model/Price.kt
  Diff:
    + val amount: Double
  AI: Using Double for monetary values can cause precision issues.
      Use Long with minor units (cents) instead.

Result: 1 violation(s), 1 warning(s)
```

### 12.2. JSON (`--json`)

```json
{
  "violations": [
    {
      "rule_id": "domain_no_infra",
      "severity": "error",
      "description": "Domain layer must not depend on infra",
      "file_path": "domain/service/UserService.kt",
      "diff_snippet": "+ import com.myapp.infra.database.UserRepository",
      "llm_explanation": "..."
    }
  ],
  "summary": {
    "errors": 1,
    "warnings": 1,
    "passed": true
  }
}
```

---

## 13. Build & Distribution

### 13.1. Makefile

```makefile
build      # go build ./cmd/guardian
test       # go test ./...
lint       # golangci-lint run
fmt        # gofmt
release    # goreleaser release
docker     # docker build
```

### 13.2. Distribution Channels

1. `go install github.com/AlexGladkov/guardian-cli/cmd/guardian@latest`
2. GitHub Releases (goreleaser — binaries for linux/darwin/windows, amd64/arm64)
3. Homebrew tap
4. Docker image

### 13.3. Dockerfile

Multi-stage build: Go builder → minimal runtime image with `guardian` binary.

---

## 14. Security & Privacy

- LLM is always on; code diffs are sent to configured LLM endpoint
- Privacy warning shown when configuring cloud LLM providers
- Local LLM (custom endpoint) recommended for confidential code
- API key stored ONLY in environment variable, never in files
- OS notifications work locally
- Hooks installed explicitly by user
- All governance artifacts (proposals, votes, history) stored in git and subject to code review
- `.guardian/` directory (state.json) is local-only, not committed

---

## 15. Exit Codes (all commands)

| Code | Meaning                              |
|------|--------------------------------------|
| 0    | Success (or warnings only for check) |
| 1    | Violations found / operation failed  |
| 2    | Config error / runtime error         |

---

## 16. Future Work (out of MVP scope)

- AST-based import analysis (instead of regex heuristics)
- Signed commits verification (`require_signed_commits: true`)
- File content allowlist for deeper analysis
- SARIF output format for GitHub Security integration
- Markdown output for PR comments
- `guardian vote --commit --push` with automatic rebase
- Windows PowerShell toast notifications
- Web UI for governance dashboard
- Webhook notifications (Slack, email)
- Monorepo support (multiple .agreements/ directories)
