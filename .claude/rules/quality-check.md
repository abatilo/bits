# Quality Checks

Run quality checks after making code changes using mise.

## Commands

| Task | Command | When to Run |
|------|---------|-------------|
| Format | `mise run fmt` | After editing Go files |
| Lint | `mise run lint` | Before committing |
| Lint + fix | `mise run lint:fix` | To auto-fix lint issues |
| Test | `mise run test` | After changes that affect behavior |
| Vet | `mise run vet` | Included in lint, rarely needed standalone |
| Build | `mise run build` | To verify compilation |

## Workflow

1. After editing Go code: `mise run fmt`
2. Before committing: `mise run lint && mise run test`
3. If lint fails with fixable issues: `mise run lint:fix`

## Notes

- Always fix lint errors before committing
- Run tests after any behavioral changes
- The `lint` task includes `go vet` checks
