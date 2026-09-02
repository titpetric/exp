# go-ddd-stats

The tool reports the size of every `.go` file below a path, per file and per
package, with a histogram over them. The path defaults to the current
directory, and vendored files are left out.

| Flag | What it writes |
|---|---|
| none, `-json` | the whole collection: every file, every package, the histogram |
| `-stats` | the histogram and the counts, without the files |
| `-d2` | the histogram as a d2 diagram, for the d2 binary to render |

```
go-ddd-stats -d2 | d2 --layout elk - docs/assets/size.svg
```

Example:

```
# go-ddd-stats -stats
{
  "Sizes": [
    {
      "Size": "< 1 KB",
      "Count": 0
    },
    {
      "Size": "< 2 KB",
      "Count": 2
    },
    {
      "Size": "< 4 KB",
      "Count": 2
    }
  ],
  "Packages": 2,
  "Files": 4
}
```

Check out `examples/` for a visualization of the data.

![](examples/gateway.png)
![](examples/dashboard.png)
