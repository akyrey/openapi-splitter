# openapi-splitter

A Go CLI tool that splits a single OpenAPI 3.0/3.1 specification file (JSON or YAML) into multiple organized files with proper `$ref` references.

Inspired by [alissonsleal/openapi-splitter](https://github.com/alissonsleal/openapi-splitter), rewritten in Go with OpenAPI 3.1 support.

## How it works

1. Reads and validates the input OpenAPI specification (JSON or YAML, 3.0.x or 3.1.x)
2. Splits every component section into individual files under named subdirectories
3. Rewrites all internal `$ref` values (e.g. `#/components/schemas/Pet`) to relative file paths (e.g. `../schemas/Pet.json`)
4. Generates a root `openapi.{ext}` that references all split files
5. Removes empty subdirectories (sections absent from the spec are not created)

## Supported component types

| Directory | Source |
|---|---|
| `paths/` | Top-level `paths` entries |
| `schemas/` | `components/schemas` |
| `parameters/` | `components/parameters` |
| `responses/` | `components/responses` |
| `requestBodies/` | `components/requestBodies` |
| `headers/` | `components/headers` |
| `securitySchemes/` | `components/securitySchemes` |
| `links/` | `components/links` |
| `callbacks/` | `components/callbacks` |
| `pathItems/` | `components/pathItems` (3.1 only) |
| `examples/` | `components/examples` |
| `tags/` | Top-level `tags` array |
| `webhooks/` | Top-level `webhooks` (3.1 only) |

## Prerequisites

- Go 1.22+

## Installation

```bash
go install github.com/akyrey/openapi-splitter@latest
```

Or build from source:

```bash
git clone https://github.com/akyrey/openapi-splitter.git
cd openapi-splitter
make build
```

## Usage

```bash
# Split a JSON spec, write output to ./openapi-split (default)
openapi-splitter api.json

# Split a YAML spec to a custom output directory
openapi-splitter api.yaml -o ./split-output

# Convert to YAML output regardless of input format
openapi-splitter api.json -f yaml

# Convert to JSON output regardless of input format
openapi-splitter api.yaml -f json

# Enable debug logging
openapi-splitter api.json -d

# Disable _index files (root references individual component files directly)
openapi-splitter api.json -n

# Use 4-space indentation
openapi-splitter api.json --indent 4

# Use indentation from .editorconfig in the current directory
openapi-splitter api.json --editorconfig

# Print version
openapi-splitter --version
```

### Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--output` | `-o` | `./openapi-split` | Output directory |
| `--format` | `-f` | _(from input)_ | Output format: `json` or `yaml` |
| `--debug` | `-d` | `false` | Enable debug logging |
| `--no-index` | `-n` | `false` | Disable `_index` files; root references individual component files directly |
| `--indent` | `-i` | `2` | Number of spaces per indentation level |
| `--editorconfig` | `-e` | `false` | Read `indent_size` from `.editorconfig` in the current directory (overridden by `--indent`) |
| `--version` | `-v` | — | Print version and exit |

## Output layout

Given an input file with paths, schemas, and parameters, the default output looks like:

```
openapi-split/
├── openapi.json          # root file — references everything below
├── paths/
│   ├── pets.json
│   └── pets_petId.json
├── schemas/
│   ├── _index.json       # component index, referenced by root
│   ├── Pet.json
│   └── Error.json
└── parameters/
    ├── _index.json
    └── limitParam.json
```

Each component directory gets an `_index.{ext}` file that maps component names to `$ref`s pointing at the individual files. The root `openapi.json` references the index files and path files directly.

### With `--no-index`

When `--no-index` / `-n` is passed, no `_index.{ext}` files are created. The root file instead references each component file directly:

```
openapi-split/
├── openapi.json          # root file — components listed individually
├── paths/
│   ├── pets.json
│   └── pets_petId.json
├── schemas/
│   ├── Pet.json
│   └── Error.json
└── parameters/
    └── limitParam.json
```

The root `openapi.json` components section becomes:

```json
"components": {
  "schemas": {
    "Pet":   { "$ref": "./schemas/Pet.json" },
    "Error": { "$ref": "./schemas/Error.json" }
  },
  "parameters": {
    "limitParam": { "$ref": "./parameters/limitParam.json" }
  }
}
```

## $ref rewriting

Internal refs are rewritten to relative paths at the correct depth:

| Location | Original ref | Rewritten ref (default) | Rewritten ref (`--no-index`) |
|---|---|---|---|
| Root `openapi.json` | — | `./schemas/_index.json` | `./schemas/Pet.json` (per name) |
| Inside `paths/pets.json` | `#/components/schemas/Pet` | `../schemas/Pet.json` | `../schemas/Pet.json` |
| Inside `schemas/Pet.json` | `#/components/schemas/Error` | `../schemas/Error.json` | `../schemas/Error.json` |

## Project structure

```
openapi-splitter/
├── main.go
├── cmd/
│   └── root.go                  # Cobra CLI: flags, arg validation, wiring
├── internal/
│   ├── parser/
│   │   └── parser.go            # Read, detect format, validate with libopenapi
│   ├── splitter/
│   │   ├── splitter.go          # Split() orchestrator
│   │   ├── context.go           # Options and Context structs
│   │   ├── mainfile.go          # CreateMainFile() — root openapi.{ext}
│   │   ├── references.go        # RewriteRefs() — $ref rewriting
│   │   ├── components.go        # splitComponentMap() generic helper
│   │   ├── schemas.go
│   │   ├── parameters.go
│   │   ├── responses.go
│   │   ├── request_bodies.go
│   │   ├── headers.go
│   │   ├── security_schemes.go
│   │   ├── links.go
│   │   ├── callbacks.go
│   │   ├── path_items.go
│   │   ├── examples.go
│   │   ├── paths.go
│   │   ├── tags.go
│   │   └── webhooks.go
│   ├── util/
│   │   └── normalize.go         # NormalizePathForFileName()
│   └── writer/
│       └── writer.go            # WriteFile() — JSON + YAML atomic write
├── testdata/
│   ├── petstore-3.0.json
│   ├── petstore-3.0.yaml
│   ├── petstore-with-tags.json
│   └── petstore-3.1.json
├── Makefile
└── .goreleaser.yaml
```

## Testing

```bash
make test       # run all tests
```

Or directly:

```bash
go test ./... -count=1
```

## License

GNU General Public License
