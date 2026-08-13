# Issue Tracker

**Type:** GitHub Issues  
**Repository:** `jackjustus/CreditTracker`

## How agents should interact

- **Read issues:** `gh issue list` / `gh issue view <number>`
- **Create issues:** `gh issue create --title "..." --body "..."`
- **Update issues:** `gh issue edit <number> --add-label "..."` etc.
- **Close issues:** `gh issue close <number>`

## Pull requests as a request surface

Off — agents do not treat external PRs as a triage input.

## Notes

The `gh` CLI must be authenticated (`gh auth status`) before any issue commands will work.
