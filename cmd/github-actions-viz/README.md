# GitHub Actions Visualizer

The tool, if ran in a github actions folder, reads in the `*.{yml,yaml}`
files and produces a mermaidjs diagram for each workflow.

To install:

`go install github.com/titpetric/exp/cmd/github-actions-viz@main`

Flags:

- `-i <folder>` - input folder (defaults to .),
- `--format <md|mermaid>` - define the output format,
- `-w` writes out files to disk.

The files written out are `action.yml.md` and `action.yml.mermaid`
respectively. The .mermaid file could be used by further tooling. See
references as the end of readme.


## Example

```mermaid
stateDiagram-v2
    workflow : tyk-docs.yml - Tyk Docs sync for release
    state workflow {
        configs: Configuration docs
        state configs {
            [*] --> step0configs
            step0configs : Checkout Config Generator Repo
            step0configs --> step1configs
            step1configs : Generate gateway config
            step1configs --> step2configs
            step2configs : Generate dashboard config
            step2configs --> step3configs
            step3configs : Generate mdcb/pump config docs
            step3configs --> step4configs
            step4configs : Store docs
            step4configs --> finish
        }

        dashboard: Dashboard docs
        state dashboard {
            [*] --> step0dashboard
            step0dashboard : Checkout Dashboard
            step0dashboard --> step1dashboard
            step1dashboard : Generate docs
            step1dashboard --> step2dashboard
            step2dashboard : Store docs
            step2dashboard --> finish
        }

        gateway: Gateway docs
        state gateway {
            [*] --> step0gateway
            step0gateway : Checkout Gateway
            step0gateway --> step1gateway
            step1gateway : Generate docs
            step1gateway --> step2gateway
            step2gateway : Store docs
            step2gateway --> finish
        }

        finish: Open PR against tyk-docs
        state finish {
            [*] --> step0finish
            step0finish : Restore artifacts
            step0finish --> step1finish
            step1finish : Print artifacts
            step1finish --> step2finish
            step2finish : Checkout Docs
            step2finish --> step3finish
            step3finish : Write out swagger schema
            step3finish --> step4finish
            step4finish : Write out markdown docs
            step4finish --> step5finish
            step5finish : Sanitize JIRA input
            step5finish --> step6finish
            step6finish : Raise docs changes PR
        }
    }
```

## Other/misc

- [Mermaid online editor](https://mermaid.live)
- [Mermaid CLI](https://github.com/mermaid-js/mermaid-cli) - Cli tooling with markdown/svg rendering
- [Docker Compose graph visualization](https://github.com/pmsipilot/docker-compose-viz)
- [GitHub PRs / Markdown](https://github.blog/2022-02-14-include-diagrams-markdown-files-mermaid/)
- [Hugo docs / Markdown](https://discourse.gohugo.io/t/correct-way-to-embed-mermaid-js/43491/3)

## Status

- Mostly responds well to a filled out `name` in github actions,
- Resolves `needs` at least on the root level correctly

Fidelity notes:

- It may have problem with deeper nesting correctness (fidelity)
- It doesn't render/note dependent workflows [stackoverflow, workflow_run](https://stackoverflow.com/questions/58457140/dependencies-between-workflows-on-github-actions)
- It doesn't render dependent steps with `id` and GITHUB_OUTPUT's

This project uses [nektos/act
pkg/model](https://pkg.go.dev/github.com/nektos/act@v0.2.49/pkg/model#Workflow)
to read the github actions workflow files.

## Usage

Used for [/exp/docs/github-actions](https://github.com/titpetric/exp/tree/main/docs/github-actions).