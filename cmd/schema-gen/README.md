# schema-gen

The `schema-gen` tool analyzes a single go source code package and dumps
the data model into a json file. The ultimate goal of the tool is to
extract a data model from any package enough to:

- create a new package with only the data model types,
- enable generating documentation based on source code fields,
- enable further rich outputs outputs like jsonschema.

To install:

`go install github.com/titpetric/exp/cmd/schema-gen@main`

Invoke `schema-gen` without arguments to print usage help.

```
./schema-gen
Usage: schema-gen <command> help
Available commands: extract, lint, markdown, restore
```

- `extract` dumps go type declarations into a .json document,
- `lint` takes a .json document and applies linter rules,
- `markdown` takes a .json document and renders a markdown doc,
- `render` takes a .json document and renders it to .go source code.

The tool is ready for general use.

Usage:

- `schema-gen`
- `schema-gen extract help`
- `schema-gen extract -i _example/ -o _example/model.json`
- `schema-gen restore -i _example/model.json -o _example/model.go.txt`
- ...

Example:

See the `example/` subfolder.

## Random facts

- we exclude `_` fields,
- we exclude unexposed fields,
- run `task` to run everything

## People to talk to

Slack #team-ext-manage-squad, Tit Petric
