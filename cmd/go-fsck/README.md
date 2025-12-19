# go-fsck

The code introspection tooling for your package layout.

```
Usage: go-fsck <command> help
Available commands: coverage, docs, extract, lint, query, report, restore, search, sqlite, stats
```

## Use cases

While the tool outgrew it's use quite quickly, I use it today to cover
various software development life cycle concerns. You could use it for
any of the following, and I do, somewhere.

The root of the `go-fsck` tool is it's data model. Extract will scan a
codebase and produce a data model in .json, with source-accurate detail. It balances
the complexity of the AST against a typed representation of it's entities.

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

- jsonschema generation from data model structs
- proto UML generation from data model, database schema model
- source code generation with naming by policy
- markdown documentation with godoc API
- linting of compliance, naming, structure
- extended cognitive complexity metrics combined with coverage

It's something to build upon. The feature existed first, while others
have been added or abandoned over time.

- `coverage`: print a coverage report, per function, per package, markdown
- `docs`: print markdown docs with package godoc, render plantuml diagrams
- `lint`: test that no package name in a project repeats, fight ambigous short imports
- `query`: a half-hearted attempt at interface discovery
- `report`: reporting test naming conventions to match symbols
- `restore`: the opinionated file grouping (symbol should match filename)
- `search`: symbol lookup, takes a reference symbol as `oas.OAS`, also with name.
- `sqlite`: it may scan go-fsck.json into a sqlite database for further querying
- `stats`: various code coupling stats, imports, reverse symbol usage, docs compliance, package stats, etc.

The errata over time is as follows:

## Linting with `lint`

The `lint` tool has limited use. Arguably it can be replaced with a `go
test -c ./...`, which will let you know the error concisely:

```shell
$ go test -c ./...
cannot write test binary loader.test for multiple packages:
github.com/titpetric/exp/cmd/go-fsck/model/loader
github.com/titpetric/exp/cmd/go-fsck/query/loader
```

The linter protects against ambigous imports, e.g. repetition of `model`
folders or similar. It's sort of hard to enforce on a repository basis,
and there is a sweet spot where it's reasonable.

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

You can try the linter:

```go
go install github.com/titpetric/tools/gofsck
```

## Schema

Running `go-fsck extract --pretty-json` will render the schema for a
package into a local `go-fsck.json` file.

Using `go-fsck restore -p package (--save)` will render the schema into a
package on disk. This package groups structs to 1 per file, keeping
grouped var declarations scoped together.

It's intent is mostly as a research tool, and it's not guaranteed to
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
the package scope. This also means it can be moved out to it's own
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
inside the package scope is a drastic constraint of feasability. You
would not be able to use this process at any kind of package scale.

Go is a package driven language - the main intent of the tool is to
organize the code in such a way where we're able to address moving code
into new packages in multiple projects that have grown too big and make
it extra difficult to maintain due to that shared package scope, design
issues and things like global shared state in tests.

Using `go-fsck` acomplishes this by enabling local behaviour tests,
essentially having the coupling / failure information as a measurable
data point for each of the types. We get to calculate impact of
refactorings in many dimensions.

## Initial design notes

The `go/ast` package is essentially very simple. There are only a few
declaration types in the language, `var`, `const`, `type` and `func`,
and that's about it for possible global symbols an application developer
cares about. A special case is the package level documentation, a
comment. There are a few other edge cases where the declaration may not
make sense, but for the most part, this encompases the go type system.

### Naming conventions

- group all `var` declarations into `vars.go`,
  - optional: group `var Err...` into `errors.go`.
  - any good convention to follow to know ErrSomething belongs to Something{} struct?
- group all `const` declarations into `const.go`,
- group all functions without receivers into `funcs.go`,
  - classify if there's a pattern we can follow to see if some of it belogs to struct internals.
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