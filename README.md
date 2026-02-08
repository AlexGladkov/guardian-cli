# Guardian

**A constitutional engine for team agreements -- serverless, git-native, LLM-powered.**

---

## What is Guardian?

Guardian is **not** CODEOWNERS. It is not about who can merge what file.

Guardian is a **constitutional engine** that lets teams define code rules and govern changes to those rules through structured proposals and distributed voting -- all stored in git, with no server required.

Rules check code correctness: forbidden imports, dangerous patterns, required conventions. When someone wants to change a rule, they create a proposal. The team votes. The constitution defines the quorum. Git history provides the audit trail.

An LLM analyzes every check, explains violations in plain language, and helps draft proposals -- so the feedback loop is fast and human-friendly.

---

## Quick Start

```bash
# Install
go install github.com/AlexGladkov/guardian-cli/cmd/guardian@latest

# Initialize in your project
guardian init

# Check your changes against team rules
guardian check

# See what proposals need your vote
guardian inbox
```

After `guardian init`, you will find a `.agreements/` directory in your project containing:

- `constitution.yml` -- governance settings, roles, quorum rules, LLM config
- `rules.yml` -- the team rules that Guardian enforces
- `proposals/` -- rule change proposals
- `votes/` -- vote records
- `history/` -- finalized proposal history
- `exceptions/` -- temporary or permanent rule exceptions

---

## How It Works

1. **Rules live in `.agreements/rules.yml`.** Each rule defines a pattern to check against code diffs (forbidden imports, banned patterns, required tags).

2. **`guardian check` validates your changes.** It runs `git diff`, applies every rule, sends the results to the configured LLM for explanation, and reports violations.

3. **Rules are governed, not dictated.** To change a rule, a developer creates a proposal with `guardian propose`. Team members vote with `guardian vote`. The constitution defines how many votes are needed (majority, two-thirds, unanimous).

4. **`guardian tally` shows the current vote count.** When enough votes are in, `guardian finalize` accepts the proposal and records it in history.

5. **Everything is in git.** Proposals, votes, and history are YAML/Markdown files committed to the repository. No external service. No database. Git is the transport and the audit log.

---

## Commands

### `guardian init`

Creates the `.agreements/` directory with template configuration files.

```bash
# Initialize a new project
guardian init

# Overwrite existing configuration
guardian init --force
```

After initialization, Guardian will launch `guardian llm configure` if the LLM is not yet configured.

**Exit codes:** 0 success, 2 error.

---

### `guardian check [base..head]`

Checks code changes against all rules defined in `rules.yml`.

```bash
# Check against origin/main (default)
guardian check

# Check a specific range
guardian check main~5..HEAD

# Machine-readable output
guardian check --json
```

**Diff range resolution (priority order):**

1. Explicit argument: `guardian check main~5..HEAD`
2. CI auto-detect: reads `$GITHUB_BASE_REF` / `$CI_MERGE_REQUEST_TARGET_BRANCH_NAME`
3. Default: `origin/main..HEAD`
4. If diff is empty: shows a hint with an example and exits 0

**What it does:**

1. Collects changed files via `git diff --name-only`
2. Runs all rules from `rules.yml` (regex-based checkers)
3. Runs a meta-check: detects unauthorized changes to `.agreements/` files without a corresponding accepted proposal
4. Applies exceptions: skips violations for paths covered by non-expired exceptions
5. Sends diff + rule descriptions + violations to the LLM for analysis
6. Prints the report

**Exit codes:** 0 OK (or warnings only), 1 violations found, 2 config/runtime error.

---

### `guardian propose <rule_id>`

Creates a proposal to modify, add, or remove a rule.

```bash
# Interactive proposal creation
guardian propose domain_no_infra

# Use LLM to draft the proposal text
guardian propose domain_no_infra --llm
```

The command prompts for:

- **Proposal type:** modify, add, or remove
- **Change description:** what you want to change
- **Reason:** why the change is needed
- **Impact:** what code will be affected

Only one active proposal per rule is allowed. The proposal file is created at `.agreements/proposals/<date>-<rule_id>.yml`. Guardian does not auto-commit; it shows a `git add` / `git commit` hint.

---

### `guardian vote <proposal_id> --yes|--no`

Records your vote on a proposal.

```bash
# Vote yes
guardian vote 2024-01-15-domain_no_infra --yes

# Vote no with a comment
guardian vote 2024-01-15-domain_no_infra --no --comment "This would break our layered architecture"
```

**Validation:**

- Voter must belong to a role listed in `governance.voters`
- If `forbid_self_approval` is true, the proposal author cannot vote yes on their own proposal
- If `allow_vote_change` is false, voting again produces an error

Vote files are created at `.agreements/votes/<proposal_id>/<voter_email>.yml`.

---

### `guardian tally <proposal_id>`

Calculates and displays the current voting results for a proposal.

```bash
# Human-readable tally
guardian tally 2024-01-15-domain_no_infra

# Machine-readable
guardian tally 2024-01-15-domain_no_infra --json
```

**Result states:**

| State      | Meaning                                             |
|------------|-----------------------------------------------------|
| `ACCEPTED` | Quorum reached with sufficient yes votes            |
| `REJECTED` | Enough no votes that yes quorum is impossible       |
| `PENDING`  | Voting still in progress                            |
| `EXPIRED`  | Proposal TTL exceeded                               |

---

### `guardian finalize <proposal_id>`

Finalizes an accepted proposal.

```bash
guardian finalize 2024-01-15-domain_no_infra
```

**Who can finalize:** any person with a role in `governance.voters`.

This command:

1. Runs `tally` internally -- only proceeds if the result is ACCEPTED
2. Updates the proposal status to `accepted`
3. Creates a history file at `.agreements/history/<proposal_id>.md`
4. Prints instructions for what to change in `rules.yml`

Guardian does **not** auto-modify `rules.yml`. The developer applies the change and commits it.

**Exit codes:** 0 success, 1 not accepted or error.

---

### `guardian withdraw <proposal_id>`

Withdraws a proposal. Only the proposal author can withdraw.

```bash
guardian withdraw 2024-01-15-domain_no_infra
```

Sets the proposal status to `withdrawn`.

---

### `guardian inbox`

Shows proposals that need the current user's vote.

```bash
# Default: fetch and show pending proposals
guardian inbox

# Send OS notification for pending votes
guardian inbox --notify

# Only show proposals since last check
guardian inbox --since-last-check

# Minimal output for hooks
guardian inbox --quiet

# Skip git fetch
guardian inbox --no-fetch

# Machine-readable
guardian inbox --json
```

The `--notify` flag triggers OS notifications:

- **macOS:** uses `terminal-notifier` if available, otherwise stdout
- **Linux:** uses `notify-send` if available, otherwise stdout
- **Windows:** stdout (PowerShell toast planned for future)

---

### `guardian hooks install|uninstall`

Manages git hooks that run `guardian inbox` automatically after git operations.

```bash
# Install post-merge and post-checkout hooks
guardian hooks install

# Remove guardian-managed hooks
guardian hooks uninstall
```

Installed hooks run `guardian inbox --notify --since-last-check --quiet` in the background so they do not block git operations. Only hooks with the `# GUARDIAN-MANAGED-HOOK` marker are affected by uninstall.

---

### `guardian history`

Displays the history of accepted and finalized proposals.

```bash
# Human-readable history
guardian history

# Machine-readable
guardian history --json
```

---

### `guardian exception create <rule_id>`

Creates a temporary or permanent exception for a rule.

```bash
guardian exception create domain_no_infra
```

The command prompts for:

- **Paths:** glob patterns or specific files to exempt
- **Reason:** why the exception is needed
- **Expires at:** optional date (empty for permanent)

If `governance.exceptions.require_approval` is true, the exception will need at least one voter's approval.

Exception files are created at `.agreements/exceptions/<exception_id>.yml`.

---

### `guardian llm configure`

Interactive LLM setup. Configures the LLM provider in `constitution.yml`.

```bash
guardian llm configure
```

The command prompts for:

1. **Provider:** DeepSeek, OpenAI, Claude (Anthropic), or Custom (OpenAI-compatible)
2. **Endpoint:** required for custom provider
3. **API key guidance:** instructs you to set `GUARDIAN_LLM_API_KEY` environment variable

The API key is **never** stored in files.

---

## Example Scenario

Here is a complete walkthrough of Guardian in action.

### 1. A developer adds code that violates a rule

A developer working on the domain layer adds an import from the infrastructure package:

```kotlin
// domain/service/UserService.kt
import com.myapp.infra.database.UserRepository  // violates domain_no_infra rule
```

### 2. Guardian check catches the violation

```bash
$ guardian check

Guardian Check Report
=====================

VIOLATION [error] domain_no_infra
  Domain layer must not depend on infra
  File: domain/service/UserService.kt
  Diff:
    + import com.myapp.infra.database.UserRepository
  AI: This import creates a direct dependency from domain to infrastructure layer.
      Consider using a domain interface instead.

Result: 1 violation(s), 0 warning(s)
```

The developer realizes the rule is too strict for the adapter pattern they are implementing.

### 3. Developer proposes a rule change

```bash
$ guardian propose domain_no_infra

Proposal type (modify/add/remove): modify
Change description: Allow infra imports in domain/adapters/
Reason: Domain adapters need to implement infra interfaces for the adapter pattern
Impact: domain/adapters/ files can now import from infra/

Proposal created: .agreements/proposals/2024-01-15-domain_no_infra.yml
Hint: git add .agreements/proposals/2024-01-15-domain_no_infra.yml && git commit -m "propose: relax domain_no_infra for adapters"
```

### 4. Team members vote

```bash
# Maria (architect) votes yes
$ guardian vote 2024-01-15-domain_no_infra --yes --comment "Makes sense for adapter pattern"

# Alex (product) votes yes
$ guardian vote 2024-01-15-domain_no_infra --yes
```

### 5. Check the tally

```bash
$ guardian tally 2024-01-15-domain_no_infra

Proposal: 2024-01-15-domain_no_infra
Rule: domain_no_infra
Type: modify

Eligible voters: 3 (ivan@example.com, maria@example.com, alex@example.com)
Quorum: two_thirds (need 2 of 3)

Votes:
  maria@example.com  YES  "Makes sense for adapter pattern"
  alex@example.com   YES

Result: ACCEPTED (2/3 yes, quorum met)
```

### 6. Finalize the proposal

```bash
$ guardian finalize 2024-01-15-domain_no_infra

Proposal 2024-01-15-domain_no_infra ACCEPTED and finalized.
History recorded: .agreements/history/2024-01-15-domain_no_infra.md

Next steps:
  1. Update .agreements/rules.yml according to the accepted change
  2. Commit the updated rules.yml
```

The developer updates `rules.yml` to add the exception for `domain/adapters/` and commits the change. Because the proposal was accepted, `guardian check` will not flag the `rules.yml` modification as unauthorized.

---

## Rule Types

### `imports_forbidden`

Prevents files matching certain globs from importing/depending on files matching other globs.

```yaml
- id: domain_no_infra
  description: Domain layer must not depend on infra
  type: imports_forbidden
  config:
    from_globs: ["domain/**"]
    forbid_globs: ["infra/**"]
  severity: error
```

**How it works:** For files matching `from_globs` that appear in the diff, Guardian scans added lines (`+` lines) for path segments matching `forbid_globs`. The LLM provides additional context and reduces false positives in its explanation.

---

### `diff_pattern_forbidden`

Forbids specific regex patterns from appearing in added lines of the diff.

```yaml
- id: money_minor_units
  description: Money must use int minor units, not float/double
  type: diff_pattern_forbidden
  config:
    forbidden_regexes:
      - "\\bDouble\\b"
      - "\\bfloat\\b"
    only_in_paths: ["**/*.kt", "**/*.java", "**/*.ts"]
  severity: warning
```

**How it works:** Optionally filters changed files by `only_in_paths`. Then applies each `forbidden_regexes` pattern to added lines. Each match is reported as a violation.

---

### `diff_pattern_requires`

Requires that at least one of the specified regex patterns is present somewhere in the diff when certain files are changed.

```yaml
- id: public_api_stability
  description: Public API changes require RFC tag in proposal
  type: diff_pattern_requires
  config:
    required_regexes:
      - "RFC:"
    only_in_paths: ["sdk/public/**"]
  severity: error
```

**How it works:** If changed files match `only_in_paths`, Guardian checks the entire diff for at least one match of `required_regexes`. If none is found, a violation is reported.

---

## Configuration

### constitution.yml

The constitution defines governance rules, roles, and LLM configuration. It is shared across the team and committed to git.

```yaml
governance:
  voters:
    - role: techlead
    - role: architect
  quorum:
    type: two_thirds
    threshold: 0.66
  forbid_self_approval: true
  allow_vote_change: false
  proposal_ttl_days: 30
  per_rule_overrides:
    critical_rule:
      quorum:
        type: unanimous
  exceptions:
    require_approval: false

identity:
  allowed_domains: ["company.com"]
  require_signed_commits: false

roles:
  techlead:
    members:
      - email: lead@company.com
  architect:
    members:
      - email: architect@company.com

llm:
  provider: deepseek
  endpoint: ""
  model: ""
  prompts:
    check_system: ""
    propose_system: ""
```

**Quorum types:**

| Type          | Logic                                              |
|---------------|----------------------------------------------------|
| `majority`    | More than 50% of eligible voters vote yes          |
| `two_thirds`  | At least 66.7% of eligible voters vote yes         |
| `unanimous`   | 100% of eligible voters vote yes                   |
| `custom`      | At least `threshold` fraction of voters vote yes   |

Voters are deduplicated by email. If a person has multiple roles, they count as one voter.

---

### rules.yml

Defines the rules Guardian checks code against. See the Rule Types section above for details on each rule type.

---

### Exceptions

Exception files live in `.agreements/exceptions/` and allow specific paths to bypass a rule temporarily or permanently.

```yaml
id: "exc-2024-01-20-domain_no_infra"
rule_id: domain_no_infra
paths:
  - "domain/legacy/old_adapter.kt"
reason: "Legacy code, will be migrated in Q2"
created_by: ivan@company.com
created_at: "2024-01-20T09:00:00Z"
expires_at: "2024-06-30T00:00:00Z"  # optional; omit for permanent exception
```

Expired exceptions are ignored by `guardian check`.

---

## Hooks and Notifications

### Installing hooks

```bash
guardian hooks install
```

This installs `post-merge` and `post-checkout` git hooks that run `guardian inbox --notify --since-last-check --quiet` in the background after every merge or checkout. The hooks do not block git operations.

### OS Notifications

When using `guardian inbox --notify`:

- **macOS:** Uses `terminal-notifier` if installed, otherwise prints to stdout
- **Linux:** Uses `notify-send` if available, otherwise prints to stdout
- **Windows:** Prints to stdout (PowerShell toast notifications planned for future)

### Removing hooks

```bash
guardian hooks uninstall
```

Only removes hooks with the `# GUARDIAN-MANAGED-HOOK` marker. Hooks managed by other tools are left untouched.

---

## LLM Integration

Guardian uses an LLM on every `guardian check` run. The LLM does **not** make pass/fail decisions -- regex rules determine that. The LLM provides:

- Plain-language explanations of violations
- Recommendations for fixing issues
- False-positive assessment and context

### Configuration

Run `guardian llm configure` to set up the LLM provider interactively. The provider and endpoint are stored in `constitution.yml` (shared with the team). The API key is stored **only** in the `GUARDIAN_LLM_API_KEY` environment variable.

### Supported Providers

| Provider   | Endpoint                         | API Format                |
|------------|----------------------------------|---------------------------|
| `deepseek` | `https://api.deepseek.com/v1`   | OpenAI-compatible         |
| `openai`   | `https://api.openai.com/v1`     | OpenAI-compatible         |
| `claude`   | `https://api.anthropic.com/v1`  | Anthropic Messages API    |
| `custom`   | User-provided URL               | OpenAI-compatible         |

### Using `--llm` with propose

The `--llm` flag on `guardian propose` sends the rule description and context to the LLM to generate a draft proposal. You can then edit the draft before confirming.

### Privacy Considerations

- When a cloud provider is selected, code diffs are sent to an external API
- For confidential or proprietary code, use the `custom` provider with a local endpoint (for example, Ollama)
- The privacy warning is shown during `guardian llm configure`
- No data is sent anywhere unless LLM is configured

---

## CI Integration

### GitHub Actions

Add Guardian to your CI pipeline:

```yaml
name: Guardian Check
on:
  pull_request:
    branches: [main]

jobs:
  guardian:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Install Guardian
        run: go install github.com/AlexGladkov/guardian-cli/cmd/guardian@latest

      - name: Run Guardian Check
        env:
          GUARDIAN_LLM_API_KEY: ${{ secrets.GUARDIAN_LLM_API_KEY }}
        run: guardian check
```

Guardian auto-detects GitHub Actions and reads `$GITHUB_BASE_REF` for the base branch.

### GitLab CI

```yaml
guardian-check:
  image: golang:1.22
  script:
    - go install github.com/AlexGladkov/guardian-cli/cmd/guardian@latest
    - guardian check
  variables:
    GUARDIAN_LLM_API_KEY: $GUARDIAN_LLM_API_KEY
  only:
    - merge_requests
```

Guardian auto-detects GitLab CI and reads `$CI_MERGE_REQUEST_TARGET_BRANCH_NAME`.

### Exit Codes in CI

| Code | Meaning                              | CI Action            |
|------|--------------------------------------|----------------------|
| 0    | All checks pass (or warnings only)   | Pipeline passes      |
| 1    | Violations found                     | Pipeline fails       |
| 2    | Configuration or runtime error       | Pipeline fails       |

---

## Security and Privacy

- **No data sent without LLM config.** Guardian does not phone home or collect telemetry. Code diffs are only sent to the LLM endpoint you configure.
- **API key in environment variable only.** The `GUARDIAN_LLM_API_KEY` is never stored in configuration files. It is read exclusively from the environment.
- **Privacy warning for cloud providers.** When you select DeepSeek, OpenAI, or Claude as the provider, Guardian shows a warning that code diffs will be sent to an external service.
- **Local LLM recommendation.** For confidential or proprietary codebases, use the `custom` provider with a local endpoint such as Ollama. This keeps all data on your machine.
- **Governance artifacts are in git.** Proposals, votes, and history files are committed to the repository and go through normal code review. The audit trail is the git log.
- **Local state is not committed.** The `.guardian/` directory (containing `state.json`) is local-only and listed in `.gitignore`.

---

## Installation

### go install

```bash
go install github.com/AlexGladkov/guardian-cli/cmd/guardian@latest
```

Requires Go 1.22 or later.

### GitHub Releases

Download pre-built binaries for your platform from the [Releases](https://github.com/AlexGladkov/guardian-cli/releases) page.

### Homebrew

```bash
brew tap AlexGladkov/tap
brew install guardian
```

### Docker

```bash
docker run --rm -v "$(pwd):/repo" -w /repo guardian:latest check
```

Or build the image yourself:

```bash
docker build -t guardian:latest .
```

---

## License

MIT License. See [LICENSE](LICENSE) for details.
