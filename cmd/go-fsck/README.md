# go-fsck

The code introspection tooling for your package layout.

```
Usage: go-fsck <command> help
Available commands: coverage, docs, extract, query, report, restore, search, sqlite, stats
```

## Use cases

While the tool outgrew its use quite quickly, I use it today to cover
various software development life cycle concerns. You could use it for
any of the following, and I do, somewhere.

The root of the `go-fsck` tool is its data model. Extract will scan a
codebase and produce a data model in .json, with source-accurate detail. It balances
the complexity of the AST against a typed representation of its entities.

```
$ go-fsck extract --help
Usage of go-fsck:
      --include-sources      include sources
      --include-tests        include test files
  -o, --output-file string   output file (default "go-fsck.json")
      --pretty-json          print pretty json
  -r, --recursive            recurse packages
  -i, --source-path string   source path (default ".")
  -v, --verbose              verbose output
```

The data model has rich traversal opportunities, as well as gives
accessibility to the data. This has proven to be valuable for:

- proto UML generation from data model, database schema model
- source code generation with naming by policy
- markdown documentation with godoc API
- linting of compliance, naming, structure
- extended cognitive complexity metrics combined with coverage

It's something to build upon. The feature existed first, while others
have been added or abandoned over time.

- `coverage`: print a coverage report, per function, per package, markdown
- `diff`: compare the exported API and the go.mod of two models, report what a release takes away
- `docs`: print markdown docs with package godoc, render plantuml diagrams
- `query`: a half-hearted attempt at interface discovery
- `report`: reporting test naming conventions to match symbols
- `restore`: the opinionated file grouping (symbol should match filename)
- `search`: symbol lookup, takes a reference symbol as `oas.OAS`, also with name.
- `sqlite`: it may scan go-fsck.json into a sqlite database for further querying
- `stats`: various code coupling stats, imports, reverse symbol usage, docs compliance, package stats, etc.

The errata over time is as follows:

## API reference with `docs`

The `docs` command renders the model as markdown: the package godoc, the
types, consts and vars it declares, and the signature of every exported
function. Function bodies are not printed, whether or not the model carries
sources.

The exception is godoc examples, which are printed whole, under an
`## Examples` heading, each wrapped in a `<section>` named after the function.
An example is its body, and it is compiled and run by `go test`, so it is
usage that can't go stale. They live in a test package, so the model needs
both flags to hold them:

```shell
go-fsck extract -i ./migrate --include-sources --include-tests
go-fsck docs > docs/api.md
```

Without `--include-sources` there is nothing to print and the heading is left
out.

## Comparing releases with `diff`

The `diff` command reads two models and reports what happened between them: to
the exported API, to the data model behind it, and to the go.mod the module is
built from. It answers the question a release has to answer before it is
tagged: does this take anything away?

```shell
git archive v1.2.0 | tar -x -C /tmp/old
go-fsck extract -i /tmp/old -r -o /tmp/old.json
go-fsck extract -i . -r -o /tmp/new.json
go-fsck diff --old /tmp/old.json --new /tmp/new.json
```

```
- github.com/go-bridget/mig/migrate.RunWithFS
~ github.com/go-bridget/mig/migrate.Print
+ github.com/go-bridget/mig/migrate.NewManager
~ github.com/go-bridget/mig/migrate.Config.Driver
~ go 1.23.0 -> 1.24.0
~ require github.com/jmoiron/sqlx v1.3.5 -> v1.4.0
+ require golang.org/x/sync v0.22.0
1 removed, 1 changed, 1 added, 1 fields, 2 requires, breaking
```

A symbol is keyed by import path, receiver type and name, so moving a
declaration to another file, or adding a sibling to its `const (...)` block, is
not a difference. Functions also carry their signature, with parameter names
removed: renaming a parameter is not a change, changing its type is. `--verbose`
prints both signatures under a changed symbol.

A type carries the shape it is declared with, which is `struct`, `interface`, or
the type it is defined as. A type that changes shape, `type ID string` becoming
`type ID int`, is a changed symbol. A type that keeps its shape is compared
field by field instead, and the fields it gained, lost or reshaped are reported
under `types` as the data model the release changes.

Only exported fields are compared, since an unexported one is not something
another module can reach. A struct tag is compared along with the field type: it
is what a document decodes through, so renaming a `json` or `yaml` key breaks
every document already written even though the code reading it still compiles.
An embedded field is reached by the last identifier of the type it embeds, and
counts because it promotes that type's method set.

What a data model change costs depends on the shape:

| Shape | field added | field reshaped | field removed |
|-----------|--------------|----------------|---------------|
| struct | not breaking | breaking | breaking |
| interface | breaking | breaking | breaking |

Adding a method to an interface stops every implementor compiling, where adding
a field to a struct costs a consumer nothing.

The go.mod of each module the two models were extracted from is compared as
well, and reported under `modules`. A release changes what it costs to depend
on a module as much as it changes what the module offers: a requirement moved
to another major version, a dependency dropped or taken on, a replace directive
pointing the build somewhere else, or a raised `go` directive that leaves older
toolchains behind.

```json
{
  "path": "github.com/go-bridget/mig",
  "go": { "old": "1.23.0", "new": "1.24.0" },
  "requires": [
    {
      "path": "github.com/jmoiron/sqlx",
      "change": "changed",
      "old": { "path": "github.com/jmoiron/sqlx", "version": "v1.3.5" },
      "new": { "path": "github.com/jmoiron/sqlx", "version": "v1.4.0" }
    },
    {
      "path": "golang.org/x/sync",
      "change": "added",
      "new": { "path": "golang.org/x/sync", "version": "v0.22.0" }
    }
  ]
}
```

Indirect requirements are left out. They are written by `go mod tidy` rather
than decided on, and a routine tidy rewrites dozens of them, which buries the
handful of direct ones that are the actual release note. `--include-indirect`
puts them back. A requirement direct on either side counts as direct, so one
that changed from indirect to direct is reported either way.

No module change sets `breaking`. A dependency is not API, and moving one takes
nothing away that a compiler will complain about.

`--json` writes the same result as `{"removed": [], "added": [], "changed": [],
"types": [], "modules": [], "breaking": false}`, where `breaking` covers removed
symbols, changed signatures and the data model changes that cost something, but
not added ones. That is the semantic version question: a breaking difference
earns at least a minor release, anything else a patch. Each entry of `types`
holds the type's `key`, `underlying` shape and its `fields`, each naming the
`change` (`added`, `changed` or `removed`) and the field on either side of it.
Each entry of `modules` holds the module `path`, the `go` and `toolchain`
directives when they moved, and its `requires` and `replaces`, which name the
same three changes and the entry on either side of them. An added type carries
its exported `fields` on the symbol itself, so a reader sees the shape without
reading the source.

Test packages, commands and internal packages are left out, since none of them
are API another module can depend on; `--include-internal` puts internal
packages back. Note that the loader skips files carrying build constraints, so
symbols behind a `//go:build` line are invisible to both sides of the
comparison.

Extraction works on an unbuilt source tree, which is what makes the `git
archive` above viable: the model is read from the AST and package load errors
are discarded, so no module cache or build is needed.

## Linting, and where the model lives now

The model, the extractor behind it and the linters over it are
[splint](https://github.com/titpetric/tools/tree/main/splint), a module of its
own. go-fsck depends on it and no longer carries any of it: every subcommand
here used to repeat the same twenty lines to build the model, and there is one
call now.

```go
options := internal.Options(".", recursive, includeTests, includeSources, verbose)
defs, err := internal.Definitions(options)
```

`go-fsck lint` went with it. What it did, and three rules besides, is what
`splint` does:

```shell
go install github.com/titpetric/tools/splint/cmd/splint@latest
splint ./...
```

`go-fsck jsonschema` went the same way, and reads better for it: `splint
--schema` renders every package of a tree rather than the first one, and takes
a `go-fsck.json` as readily as a source path.

```shell
splint --schema ./... > schema.json
splint --schema --input go-fsck.json > schema.json
```

splint carries a second parser as well, one that reads the source without
building a syntax tree. It produces the same document, reads source that does
not compile, and is twenty times quicker. Anything reading a `go-fsck.json`
reads what it writes: the schema is the same one, and splint compares the two
value by value over sixteen repositories to keep it that way.

## Interface discovery with `query`

With new codebases, it's almost inevitable that I need to inspect the largest
package scope. There's usually one or more implementations that are modular
in some way, if that's a `http.HandlerFunc`, or something else.

The tool has `--show-handlers` and `--middleware` flags that search for
some particular function signatures.

The attempt is to find code that is grouped by function signatures,
type returns or otherwise common function API. It's more than common
that these should be decomposed into a new package.

## Restoring a codebase with `restore`

This is the missing part to `go fmt` for the codebase. The restore rules
aren't defined well enough, and without deterministic rules, the
restored code may only be partially usable.

I used the feature exactly once, [forking a rate limiter project](https://github.com/TykTechnologies/exp/tree/main/pkg/limiters).

I've accepted linting may be the only approach to the issue, even if
fixing could be made deterministic. I tend to follow code grouping
naturally, but wouldn't mind a sanity check in the pipeline.

You can try the linters:

```shell
go install github.com/titpetric/tools/gofsck@latest
go install github.com/titpetric/tools/splint/cmd/splint@latest
```

## Schema

Running `go-fsck extract --pretty-json` will render the schema for a
package into a local `go-fsck.json` file.

Each package in the model carries a `Module` block, read from the go.mod that
governs it. The go.mod is found by searching upward from the source path, so
`go-fsck extract -i model/` reaches the one a level or more above it, and a
recursive extract records each package under the go.mod of its own module:

```json
{
  "Module": {
    "Path": "github.com/titpetric/exp/cmd/go-fsck",
    "GoVersion": "1.27.0",
    "Toolchain": "go1.27.0",
    "Requires": [
      { "Path": "github.com/spf13/pflag", "Version": "v1.0.10" },
      { "Path": "golang.org/x/sys", "Version": "v0.47.0", "Indirect": true }
    ],
    "Replaces": [
      { "Path": "github.com/spf13/pflag", "NewPath": "../pflag" }
    ]
  }
}
```

The model is a flat list of packages with nowhere else to hold module level
facts, so the block repeats on every package of the same module. A package
extracted from a tree holding no go.mod carries none. `Requires` and `Replaces`
are sorted by module path, so reordering a require block does not change the
model.

Using `go-fsck restore -p package (--save)` will render the schema into a
package on disk. This package groups structs to 1 per file, keeping
grouped var declarations scoped together.

Its intent is mostly as a research tool, and it's not guaranteed to
handle every possible edge case in terms of how people structure their
code.

Generally the tool requires `goimports -w .` to fix the imports, as
it does not handle those in a fine grained way (yet). Improvement is
possible, but also, there are tools like goimports that implement this
logic and we depend on that functionality as a development shortcut.

## Current state

I define local behaviour as the completeness of the implementation, by
invoking `go test file*.go` it reduces the scope only to these files. If
the package only imports other packages, the behaviour of the
implementation and the tests is local - does not need other symbols in
the package scope. This also means it can be moved out to its own
package and make other code have local behaviour.

This is in effect a black box test, if there is no shared package scope.
Test utilities are a common coupling that belongs in a separate package.

- The tool implements --save, but two different models emerge, this tool
  is aimed for DDD schema, mainly grouping by structs. Packages that
  provide a package-level API need to be structured by functions.
  How do we better handle the case of conventions for something
  similar to "strings" package?

- To get real use of the tool we need to build a test harness that would
  run the isolation tests against individual file groups in a restored
  package. This way we can figure out offline which types and functions
  can be extracted into subpackages, and what kind of % of the package
  that extraction represent (how much smaller it gets).

- Restoring with -p allows us to restore blackbox tests separately.
  We mostly have tests in the same scope. Unit tests are not a thing,
  and we know that tests with StartTest() are expensive. There's an
  extreme solution: move StartTest behind an `e2e` tag, and instantly
  move all the tests that require it behind the same e2e tag. This
  does a few things:

  1. it splits the already ~4 minute running test for the package
     into two pipelines running in parallel. Unit tests do not
     depend on storage, are cheap to run, but need writing in
     the first place.

  2. supposedly leaves just to add an `integration` tag for actual
     integration points, like testing the 'storage' package, giving
     us a third parallel pipeline.

  3. Code and tests are inherently coupled. The biggest effort is
     keeping TestA in the scope of A struct, or A function. But
     some packages are function oriented, other more struct and
     interface. This tool is firstly aimed at the struct case.

- Restore needs work (sorting symbols is a big chunk).

In a single package, when a struct A depends on struct B and C, then the
behaviour of A is not local. However, if B and C are imported from
packages, then the behaviour of A is local. Another way to remove the
dependency is to update A to use interfaces, which are satisfied by both
B and C, and then behaviour becomes local.

To really get advantage of the tool, using `type ( ... )` groups is
encouraged. If you have multiple declaration types in a single type
group, the tool will keep these together and group the code into the file
corresponding to the *shortest* of the type names. The following code
would be a red flag:

```go
type FieldName struct {}
type FieldKind struct {}
type Field struct {
       Name FieldName
       Kind FieldKind
}
```

In order to hint the types are depending on each other, the
correct way to implement that is:

```go
type (
     FieldName struct {}
     FieldKind struct {}
     Field struct {
       Name FieldName
       Kind FieldKind
     }
)
```

And this should live in `field.go`.

This mostly applies to investigate cases of service structs, and not data models.

By default, `go-fsck` should be really good at taking a data model package
and laying it out in go files that are named by the types. It makes a
flat 1-1 file structure for types, with the grouping behaviour above.

## Run it on your project

If you want to run it on your project, which is highly not recommended for
anything resembling production use, you can use this taskfile:

```yaml
---
version: "3"

desc:
  default:
    desc: "Run go-fsck and restore the package"
    cmds:
      - go-fsck extract .
      - go-fsck restore -p folder --save
```

I often use `go-fsck extract ./...` to inspect the complete source tree.

By default, go-fsck should leave `pkg.go` alone, but I have
no idea if it's implemented correctly (QA: none). There are
implementation gaps and some things are not handled. Mileage may vary.
Data loss is expected so small packages fit best.

## Future

The actual granularity between packages with 1, 10, 100 or 1000 types
inside the package scope is a drastic constraint of feasibility. You
would not be able to use this process at any kind of package scale.

Go is a package driven language - the main intent of the tool is to
organize the code in such a way where we're able to address moving code
into new packages in multiple projects that have grown too big and make
it extra difficult to maintain due to that shared package scope, design
issues and things like global shared state in tests.

Using `go-fsck` accomplishes this by enabling local behaviour tests,
essentially having the coupling / failure information as a measurable
data point for each of the types. We get to calculate impact of
refactorings in many dimensions.

## Initial design notes

The `go/ast` package is essentially very simple. There are only a few
declaration types in the language, `var`, `const`, `type` and `func`,
and that's about it for possible global symbols an application developer
cares about. A special case is the package level documentation, a
comment. There are a few other edge cases where the declaration may not
make sense, but for the most part, this encompasses the Go type system.

### Naming conventions

- group all `var` declarations into `vars.go`,
  - optional: group `var Err...` into `errors.go`.
  - any good convention to follow to know ErrSomething belongs to Something{} struct?
- group all `const` declarations into `const.go`,
- group all functions without receivers into `funcs.go`,
  - classify if there's a pattern we can follow to see if some of it belongs to struct internals.
- group all types into `<name>.go`,
- group all `Test<Name>*` functions into `<name>_test.go`,
- group remaining functions into `funcs_test.go`,
- group all the interfaces into `interfaces.go`,
- store package doc in `doc.go`.

### Non-goals:

- build tags?
- dot imports
- multiple `init` functions per package
- unnamed `_` vars?
- supporting `./...` to reformat the world (do we need it?)

Things that are enabled by this:

Restructuring the package to above conventions would let us surface
bounded contexts for individual declarations. Surfacing bounded context
for declarations uses `go test` to reduce scope only to particular files.
Code may not be coupled to anything in the package (strict) and if we can
test for that, we can move it out. Moving things out lets us test better.

For each resulting declaration, we can surface bounded contexts like so:

- strict: `go test <name>*.go const.go`
- with vars: `go test <name>*.go const.go vars.go`
- with funcs: `go test <name>*.go const.go funcs.go`
- with funcs and vars: `go test <name>*.go const.go vars.go funcs.go`
- additional cases for all with `interfaces.go`.

Now, code, with small adjustments, may be possible to become strictly
bounded. For example, it may implement an internal function that landed
in `funcs`, but is not used otherwise. Running the strict check will
surface these explicit couplings and let us know which declarations
depend on others, and what the coupling level inside the package is.

Anything that's not a public declaration inside `vars.go` is a code
smell, hinting at global singletons. It takes additional conventions to
make singletons safe (e.g. interfaces, mutexes, pointer swaps, etc.).
Having those grouped in a nice little `vars.go` file is nice. Globals
need to be understood and protected and testing with t.Parallel is going
to be a pain if the data is shared. Even reusing global loggers is a
code smell, because you can never move that file out without changes.

## Summary

This tool will let us pick code apart more safely. We can see what's
already implemented in ways that let us extract it from large package
scope. The benefit of smaller package is focus when addressing defects,
and this is the main goal of the tool, to enable that analysis and act on
the data. We often don't know how large problems are due to large package
scopes and couplings, this gives us data.

## Fidelity

As it may produce unwanted results, the way to use the tool is to
generate it from a package, and output to a new package. Using it
is expected to have bugs (I am my own QA), but - here's a few caveats:

- the premise is simple: the package would compile if we had all the
  symbols in one file, or if we had them scattered in a thousand,
- when we essentially restructure the package, this is a significant
  automated change. The change will be attributed to the commiter,
- if you'd use it, i'd suggest a git hook to check it on pre_commit,
  or even better, run it by hand in `task fmt` or something,
- it may not work in various use cases, things like go version may be
  problematic, generally we build it on a recent one and see,
- just consider it an academics tool, rather than a CI one. I don't
  expect this to be stable, so control the invocation.
- i mean, it's in the experimental repo...

## Aggregations

A few aggregations of symbols are available below. Using `jq`
lets us transform our schema into either an array of key value pairs,
or an object. Jq examples filter the count and allow some
degree of customization to quickly adjust the json schema in order
to inspect it with various code pens.

Example code pen:

- https://codepen.io/kendsnyder/pen/vPmQbY
- https://codepen.io/thecraftycoderpdx/pen/jZyzKo

---

Use case: number of symbols in files as an array of {name, value}:

```
go-fsck restore -p gateway --stats-files --remove-tests | \
    jq -s '.[] | select( all(.Count; . > 10) ) | {"name": .File, "value": .Count}' | \
    jq -s
```

Example:

```
[
  {
    "name": "api_definition_loader.go",
    "value": 36
  },
  {
    "name": "api_spec.go",
    "value": 21
  },
```

---

Use case: number of symbols in files as an object with key/value:

```
go-fsck restore -p gateway --stats-files --remove-tests | \
    jq -s '.[] | select( all(.Count; . > 10) )' | \
    jq -s 'to_entries | map( {(.value.File) : (.value.Count)} ) | add'
```

Example:

```json
{
  "api_definition_loader.go": 36,
  "api_spec.go": 21,
  "base_middleware.go": 19,
...
```

---