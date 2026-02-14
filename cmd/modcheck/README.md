# modcheck

This tool will go through the imports in go.mod and check with the
official go proxy to get a list of versions for each of the imports.

## Install

```bash
go install github.com/TykTechnologies/exp/cmd/modcheck@main
```

## Usage

```
Usage of modcheck:
      --for-upgrade    only list packages for upgrade
      --json           output as JSON
      --skip strings   skip packages
      --suggest        print go get commands to update dependencies
```

Run `modcheck` in your repo where `go.mod` exists, or pass a module
path as the first argument to check a remote module from the Go proxy:

```
modcheck github.com/titpetric/atkins@latest
```

The report is provided in markdown output, suitable for github issues.

## Comparison: atkins vs task

The following compares the direct dependencies of
[atkins](https://github.com/titpetric/atkins) and
[task](https://github.com/go-task/task) (v3).

### Summary

| Metric | atkins | task |
|:---|:---|:---|
| Direct dependencies | 9 | 29 |
| Total files | 515 | 2,282 |
| Total size | 7.2 MB | 21.4 MB |
| Shared dependencies | 4 | 4 |

### atkins dependencies (9 deps, 7.2 MB)

| Import | Size | Files |
|:---|:---|:---|
| creack/pty | 66.3 KB | 65 |
| expr-lang/expr | 6.0 MB | 217 |
| oklog/ulid/v2 | 68.1 KB | 12 |
| spf13/pflag | 357.3 KB | 88 |
| stretchr/testify | 654.8 KB | 63 |
| titpetric/cli | 22.0 KB | 10 |
| golang.org/x/sync | 58.5 KB | 19 |
| golang.org/x/term | 52.6 KB | 17 |
| gopkg.in/yaml.v3 | 452.1 KB | 24 |

### task dependencies (29 deps, 21.4 MB)

| Import | Size | Files |
|:---|:---|:---|
| charm.land/bubbles/v2 | 378.5 KB | 97 |
| charm.land/bubbletea/v2 | 169.4 KB | 74 |
| charm.land/lipgloss/v2 | 508.7 KB | 257 |
| Ladicle/tabwriter | 36.4 KB | 7 |
| Masterminds/semver/v3 | 133.7 KB | 20 |
| alecthomas/chroma/v2 | 8.4 MB | 1,082 |
| chainguard-dev/git-urls | 13.9 KB | 11 |
| davecgh/go-spew | 207.0 KB | 24 |
| dominikbraun/graph | 443.2 KB | 47 |
| elliotchance/orderedmap/v3 | 37.5 KB | 6 |
| fatih/color | 47.1 KB | 10 |
| fsnotify/fsnotify | 248.5 KB | 119 |
| go-task/slim-sprig/v3 | 144.7 KB | 61 |
| go-task/template | 290.2 KB | 26 |
| google/uuid | 76.4 KB | 31 |
| hashicorp/go-getter | 424.9 KB | 220 |
| joho/godotenv | 40.3 KB | 20 |
| mitchellh/hashstructure/v2 | 30.6 KB | 9 |
| puzpuzpuz/xsync/v4 | 153.9 KB | 25 |
| sajari/fuzzy | 6.2 MB | 8 |
| sebdah/goldie/v2 | 51.5 KB | 13 |
| spf13/pflag | 357.3 KB | 88 |
| stretchr/testify | 654.8 KB | 63 |
| zeebo/xxh3 | 614.4 KB | 28 |
| go.yaml.in/yaml/v4 | 584.2 KB | 50 |
| golang.org/x/sync | 58.5 KB | 19 |
| golang.org/x/term | 52.6 KB | 17 |
| mvdan.cc/sh/moreinterp | 8.9 KB | 6 |
| mvdan.cc/sh/v3 | 966.4 KB | 100 |

### Shared dependencies

Both projects share 4 direct dependencies:

- `spf13/pflag` (357.3 KB)
- `stretchr/testify` (654.8 KB)
- `golang.org/x/sync` (58.5 KB)
- `golang.org/x/term` (52.6 KB)

### Key takeaways

- **atkins has ~3x fewer dependencies** (9 vs 29) and **~3x less dependency weight** (7.2 MB vs 21.4 MB).
- **task's largest deps are chroma (8.4 MB)** for syntax highlighting and **sajari/fuzzy (6.2 MB)** for fuzzy matching — these two alone account for ~68% of task's total dependency size.
- **atkins' largest dep is expr-lang/expr (6.0 MB)** which accounts for ~83% of its total.
- task depends on several **pre-release/rc versions** (charm.land libs, go.yaml.in/yaml/v4) and **unreleased commits** (mvdan.cc/sh, davecgh/go-spew).
