# Contributing to untis-notifier

Thanks for considering a contribution! Here's how to get started.

## Reporting bugs

Please open an issue and include:

- Your Docker / Go version
- Relevant log output (`docker compose logs untis-notifier`)
- What you expected vs. what happened
- **Do not include your credentials or school name**

## Suggesting features

Open an issue with the `enhancement` label. Describe the use case — not just the feature — so we can understand the problem you're solving.

## Pull requests

1. **Open an issue first** for anything beyond small fixes, so we can agree on the approach before you invest time writing code
2. Fork the repo and create a branch: `git checkout -b my-feature`
3. Write tests for any new logic (see `diff/diff_test.go` and `notifier/ntfy_test.go` for examples)
4. Make sure `go vet ./...` and `go test ./...` pass
5. Open a PR with a clear description of what changed and why

## Development setup

```bash
git clone https://github.com/crwntec/untis-notifier
cd untis-notifier
cp .env.example .env   # fill in your credentials
go build -o untis-notifier . && ./untis-notifier
```

For live reload:

```bash
go install github.com/air-verse/air@latest
air
```

## Code style

- Standard `gofmt` formatting (enforced by CI via `go vet`)
- Keep packages focused: `untis` for API interaction, `diff` for change detection, `notifier` for delivery
- Prefer explicit error wrapping with `fmt.Errorf("context: %w", err)`
