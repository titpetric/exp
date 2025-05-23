# Cmd runbook

- Each cmd/ has a Taskfile.
- The cmd/Taskfile.yml executes the default target on all taskfiles.
- Cmds usually `go install .`, but sometimes more.

## Typical maintenance tasks

### Update go version

```bash
task api:list | xargs -I{} echo 'cd {} ; go mod edit -go=1.24.3 ; cd ..' | sh -x
```

### Update go modules

```bash
task api:list | xargs -I{} echo 'cd {} ; go get -u ./... ; go mod tidy ; cd ..' | sh -x
```