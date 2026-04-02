# CLAUDE.md

Project context for Claude Code sessions.

## What is this project?

`openapi-splitter` is a Go CLI tool that takes a single OpenAPI 3.0/3.1 specification file (JSON or YAML) and splits it into multiple organized files with proper `$ref` references. It is the Go equivalent of [alissonsleal/openapi-splitter](https://github.com/alissonsleal/openapi-splitter), extended to support OpenAPI 3.1.

## Tech stack

- **Language**: Go 1.22+
- **Dependencies**: `github.com/pb33f/libopenapi` + `github.com/pb33f/libopenapi-validator` (validation), `github.com/spf13/cobra` (CLI), `gopkg.in/yaml.v3`
- **Build**: `make build` (outputs `./openapi-splitter`) or `go build -o openapi-splitter .`
- **Tests**: `go test ./...`

## Project structure

```
main.go                              # entrypoint — delegates to cmd.Execute()
cmd/root.go                          # Cobra root command: flags, arg validation, wiring
internal/
  parser/parser.go                   # Read file, detect format (JSON/YAML), decode to map, validate with libopenapi
  splitter/
    splitter.go                      # Split() orchestrator: clean output dir, run all splitters, create main file, remove empty dirs
    context.go                       # Options + Context structs; Log/Logf debug helpers
    mainfile.go                      # CreateMainFile() — generates root openapi.{ext} referencing all split files
    references.go                    # RewriteRefs() — rewrites #/components/... $refs to relative file paths
    components.go                    # splitComponentMap() — generic helper used by all component splitters
    schemas.go / parameters.go / responses.go / request_bodies.go / headers.go /
    security_schemes.go / links.go / callbacks.go / path_items.go / examples.go /
    paths.go / webhooks.go            # one file per component type (tags are preserved inline)
  util/normalize.go                  # NormalizePathForFileName() — converts /pets/{petId} → pets_petId
  writer/writer.go                   # WriteFile() — JSON (pretty) or YAML, atomic write (temp + rename)
testdata/
  petstore-3.0.json                  # OpenAPI 3.0.3 fixture
  petstore-3.0.yaml                  # Same in YAML
  petstore-with-tags.json            # 3.0 fixture with top-level tags array
  petstore-3.1.json                  # OpenAPI 3.1.0 fixture with webhooks, requestBodies, headers, examples, securitySchemes
```

## Key design decisions

- **Single positional arg**: the input file path. All other options are flags (`-o`, `-f`, `-d`).
- **Format detection**: by file extension (`.json`, `.yaml`, `.yml`), falling back to content sniffing (`{` prefix → JSON). Output format defaults to the input format; `-f` overrides.
- **Generic map representation**: the document is parsed into `map[string]interface{}` (not into typed structs) so the splitter can freely manipulate any field without being constrained by the 3.0 type model.
- **libopenapi validation**: used for schema validation only. Natively supports OpenAPI 3.0, 3.1, and 3.2 — no workarounds needed for `jsonSchemaDialect` or `webhooks` fields.
- **Clean output on each run**: the output directory is fully removed and recreated. No incremental/merge behaviour.
- **$ref depth**: files at the root (`openapi.{ext}`) use `./dir/Name.ext` refs; files inside a subdirectory (e.g. `paths/pets.json`) use `../dir/Name.ext`. Controlled by the `depth` parameter in `RewriteRefs`.
- **Component index files**: each non-empty component directory gets an `_index.{ext}` file mapping component names to `$ref`s for the individual files. The root file references `./schemas/_index.json` rather than listing every schema inline. Disabled by `--no-index`.
- **Indentation**: output files default to 2-space indentation. Controlled by `--indent` (explicit) or `--editorconfig` (reads `indent_size` for `*.json`/`*.yaml` from CWD). Priority: `--indent` > editorconfig > default (2). Zero/negative `IndentSize` in `Options` defaults to 2 inside `Split()`.
- **Atomic write**: `writer.WriteFile` writes to a temp file then renames. Falls back to direct write on cross-device rename errors.
- **Path normalization**: `NormalizePathForFileName` strips leading slash, replaces `/` with `_`, strips `{}`  from path parameters, and trims trailing underscores. `/pets/{petId}` → `pets_petId`.
- **Version injection**: `version` constant in `cmd/root.go` is a compile-time constant; goreleaser sets it via `-ldflags "-X main.version={{.Version}}"` (note: the variable lives in `cmd`, so the ldflag target is `github.com/akyrey/openapi-splitter/cmd.version`).

## CLI flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--output` | `-o` | `./openapi-split` | Output directory |
| `--format` | `-f` | _(from input)_ | Output format: `json` or `yaml` |
| `--debug` | `-d` | `false` | Enable debug logging |
| `--no-index` | `-n` | `false` | Disable `_index` files; root references individual component files directly |
| `--indent` | `-i` | `2` | Spaces per indentation level in output files |
| `--editorconfig` | `-e` | `false` | Read `indent_size` from `.editorconfig` in CWD; overridden by `--indent` |
| `--version` | `-v` | — | Print version and exit |

## Commands

```bash
# Build
make build

# Split a spec
./openapi-splitter api.json
./openapi-splitter api.yaml -o ./output -f json
./openapi-splitter api.json -d   # debug mode
./openapi-splitter api.json -n   # no index files

# Test all packages
make test          # or: go test ./... -count=1

# Lint (requires golangci-lint)
make lint

# Install to $GOPATH/bin
make install

# Remove build artifact
make clean
```

## Testing conventions

- All test packages use `package <name>_test` (black-box), except where package-internal access is required.
- Table-driven tests with `t.Run`.
- Integration tests in `internal/splitter/splitter_test.go` use `t.TempDir()` as the output directory and inspect the filesystem directly.
- Test fixtures live in `testdata/` at the repo root and are referenced by relative path from test files via `../../testdata/`.

## OpenAPI 3.1 note

`libopenapi` natively supports OpenAPI 3.0, 3.1, and 3.2. No workarounds are needed for `jsonSchemaDialect`, `webhooks`, or any other 3.1-specific fields — they are modeled directly. Validation uses `libopenapi.NewDocument(data)` → `validator.NewValidator(document)` → `v.ValidateDocument()` which returns `(bool, []*errors.ValidationError)`.

## Release

Releases are automated via GoReleaser (`.goreleaser.yaml`) and triggered by pushing a `v*` tag. The `release.yml` GitHub Actions workflow runs tests first, then GoReleaser.

Secrets required in the GitHub repository:
- `CODECOV_TOKEN`: for coverage uploads in CI
- `HOMEBREW_TAP_GITHUB_TOKEN`: personal access token with `repo` scope, needed for GoReleaser to push the Homebrew formula to `akyrey/homebrew-tap`
