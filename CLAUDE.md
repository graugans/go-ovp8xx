# CLAUDE.md

Go client library and CLI (`ovp8xx`) for the ifm OVP8xx device series. Personal project, not an official ifm product.
Module path: `github.com/graugans/go-ovp8xx/v2`.

## Layout

- `cmd/ovp8xx/` – Cobra based CLI (entry point `ovp8xx.go`)
- `pkg/ovp8xx/` – XML-RPC client for the device configuration
- `pkg/pcic/` – PCIC protocol (ifm proprietary TCP result stream), chunks and frames
- `pkg/swupdater/` – firmware update via websocket
- `internal/versioninfo/` – version info injected at build time
- `doc/` – user documentation (e.g. config filtering)

## Commands

```sh
go build ./...
go test -timeout=1m ./...
golangci-lint run ./...
go build cmd/ovp8xx/ovp8xx.go   # build the CLI
```

## Notes

- PCIC test data in `pkg/pcic/testdata/*.bz2` is stored in **Git LFS**. If the PCIC tests fail with read
  errors, run `git lfs pull` first (pointer files instead of real data).
- Go version is defined in `go.mod` only; CI (`.github/workflows/go.yml`) reads it via `go-version-file`.
  The dev container image tag in `.devcontainer/devcontainer.json` should match.
- Releases are cut automatically by go-semantic-release + goreleaser, so commit messages must follow
  Conventional Commits (`feat:`, `fix:`, `doc:`, `build(deps):`, …).
- pre-commit hooks: gofmt, goimports, golangci-lint, unit tests, shfmt (`-i 4`), shellcheck.
- The API is still allowed to break; keep the README in sync with CLI changes.
