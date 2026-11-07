# Scripts

This directory contains helper scripts for managing the GitHub backlog and Project settings for the database monitoring v2 work. They are designed to be safe, idempotent, and easy to run on Windows PowerShell with the GitHub CLI (`gh`).

All scripts assume you have `gh` installed and authenticated.

- Install: https://cli.github.com/
- Authenticate: `gh auth login`
- Ensure required scopes: `gh auth refresh -s repo -s project -s read:project`


## 1) seed_github_v2.ps1 — Seed or reconcile the v2 backlog
Creates labels, milestones, Project custom fields, and the full set of v2 Issues, then links each Issue to your GitHub Project and sets fields (Status/Area/Size). Safe to re‑run; will not duplicate issues and will skip closed ones.

Usage (PowerShell):

```
# Pilot mode (creates just A1 and A2) — recommended first run
./scripts/seed_github_v2.ps1 -Owner "<owner>" -Repo "<repo>" -ProjectNumber 1 -PilotOnly

# Full backlog (creates/ensures all issues and links to Project)
./scripts/seed_github_v2.ps1 -Owner "<owner>" -Repo "<repo>" -ProjectNumber 1
```

Parameters:
- `-Owner` (string): Repository owner (user or org). Example: `guilhermearpassos`
- `-Repo` (string): Repository name. Example: `database_monitoring`
- `-ProjectNumber` (int): Projects v2 project number (see the URL). Example: `1`
- `-PilotOnly` (switch): When present, only seeds two pilot issues (A1, A2) for quick verification.

Behavior:
- Ensures labels: `dbmon-v2`, `area:*`, `complexity:*`, `kind:feature` (idempotent)
- Ensures milestones: Milestones A..J (idempotent)
- Ensures Project single‑select fields: `Area`, `Size` (idempotent)
- Creates or updates Issues per milestone with clear acceptance criteria and DMV/SQL examples where relevant
- Links Issues to the specified Project, sets `Status`, `Area`, and `Size`
- Idempotency rules:
  - If an issue with the same Title exists and is open → it is reconciled (labels/milestone/Project fields ensured)
  - If it exists and is closed → it is skipped (not modified)
  - If it does not exist → it is created

Troubleshooting:
- Missing scopes → run: `gh auth refresh -s repo -s project -s read:project`
- Permission errors on Projects → ensure the Project is owned by `-Owner` and you have write access
- Labels already exist → script ignores and continues


## 2) setup_github_project_views.ps1 — Quick helper for creating useful Project views
A minimal helper that validates access to your user‑owned Project and prints exact UI steps to create the recommended views. GitHub’s public GraphQL for Projects v2 view creation is not consistently available; these instructions are guaranteed.

Usage:
```
./scripts/setup_github_project_views.ps1 -Owner "<owner>" -ProjectNumber 1
```

It will verify the project and then instruct you to create (via UI):
- Engineering Board (Board) — Filter: `label:dbmon-v2`, Group by: `Status`
- Area Board (Board) — Filter: `label:dbmon-v2`, Group by: `Area`
- Current Iteration (Board) — Filter: `label:dbmon-v2 iteration:@current`, Group by: `Status`
- Milestone Progress (Table) — Filter: `label:dbmon-v2`

Optional (UI):
- Roadmap view — Filter `label:dbmon-v2`
- Charts under “Milestone Progress”: 
  - Progress by Milestone (Bar) — Group by: Milestone, Stack by: Status
  - Weekly Throughput (Line) — Filter includes `is:closed`, Group by: Week


## 3) gh_issue_tool.ps1 — One‑off GitHub Issue and Project operations
A Swiss‑army‑knife for working with a single issue at a time. Supports read/create/update/comment and Project linking/field updates. Idempotently ensures labels and avoids duplicate Project items. Guards against editing closed issues unless you explicitly reopen or force.

Common prerequisites:
- `gh` authenticated and scoped: `gh auth refresh -s repo -s project -s read:project`
- Optionally set default repo context: `$env:GH_REPO = "<owner>/<repo>"`

Actions:
- `read` — read and print basic issue details (number/title/state/url/labels)
- `create` — create a new issue with title/body, labels, and milestone
- `update` — edit labels, milestone; append/replace body; close/reopen
- `comment` — add a comment from inline text or a file
- `project-link` — add the issue to a Project and set fields (Status/Area/Size)
- `set-fields` — update Project fields for an already‑linked issue

Examples:
```
# Read by URL
./scripts/gh_issue_tool.ps1 -Action read -Issue https://github.com/<owner>/<repo>/issues/123

# Create (labels ensured automatically)
./scripts/gh_issue_tool.ps1 -Action create -Owner "<owner>" -Repo "<repo>" -Title "Spike: QS feature detection" -BodyPath .\ticket.md -AddLabels area:metrics,dbmon-v2 -Milestone "Milestone E - Metrics v1"

# Update (append to body, label changes, set milestone)
./scripts/gh_issue_tool.ps1 -Action update -Issue 123 -Owner "<owner>" -Repo "<repo>" -AppendBody -Body "New findings from test env." -AddLabels area:metrics -RemoveLabels area:bootstrap -Milestone "Milestone E - Metrics v1"

# Reopen or close
./scripts/gh_issue_tool.ps1 -Action update -Issue 123 -Owner "<owner>" -Repo "<repo>" -Reopen
./scripts/gh_issue_tool.ps1 -Action update -Issue 123 -Owner "<owner>" -Repo "<repo>" -Close

# Comment
./scripts/gh_issue_tool.ps1 -Action comment -Issue 123 -Owner "<owner>" -Repo "<repo>" -Body "LGTM, proceeding to implement."

# Link to user Project 1 and set fields
./scripts/gh_issue_tool.ps1 -Action project-link -Issue 123 -Owner "<owner>" -Repo "<repo>" -ProjectNumber 1 -ProjectStatus "Todo" -ProjectArea "Metrics" -ProjectSize "M"

# Later: update only the fields (already linked)
./scripts/gh_issue_tool.ps1 -Action set-fields -Issue https://github.com/<owner>/<repo>/issues/123 -Owner "<owner>" -ProjectNumber 1 -ProjectStatus "In Progress"
```

Key behaviors and flags:
- Closed issues are not modified on `update` unless you pass `-Reopen` or `-Force`
- Label ensuring is idempotent; missing labels are created automatically when possible
- Project linking is idempotent; if already linked, fields are updated only
- `-DryRun` can be used to print intended operations without making changes
- Owner/Repo inference: if `-Owner`/`-Repo` not provided, the tool attempts to use `$env:GH_REPO`


## Tips & FAQ
- “Field/command not found” errors often indicate an older `gh` version. Upgrade `gh` and retry.
- For org‑owned Projects, use the org name in `-Owner` and make sure your token has org project access.
- To rollback mass‑created issues: filter by label `dbmon-v2` in Issues UI, bulk close or edit as needed.
- Iterations field is created in the Project UI. Filters like `iteration:@current` work only after an active iteration exists.


## License
These scripts are part of the repository and follow the repository’s LICENSE.
