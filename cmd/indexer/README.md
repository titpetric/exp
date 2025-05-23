# Indexer

This tool generates an `index.html` file to be used in cases where:

- you don't have nginx `autoindex on;`
- github pages directory indexes
- ...

## Install

```
go install github.com/titpetric/exp/cmd/indexer@main
```

Just run `indexer` in victim folder.

## Flags

```
# indexer --help
Usage of indexer:
  -i string
    	Input folder (default: current folder) (default ".")
  -o string
    	Output filename or absolute filepath (default "index.html")
  -t string
    	Title of index page to render
  -template string
    	Template to render (index.tpl is bundled) (default "index.tpl")
```
