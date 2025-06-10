# Dumb K/V type blob storage

This serves as a distribution system for large files, e.g. zipped
archives, video, node_modules.

It's a functional KV store for files.

## Notes

In this case, `curl` clearly loses out over wget. Somehow curl truncates
the PUT request body to ~135MB consistently. Wget works first try. Shame
on curl for letting me question the quality of MY code.

```yaml
- wget \
    --method=PUT \
    --header='Content-Type: application/octet-stream' \
    --body-file=data.zip \
    http://localhost:8080/data.zip -O /dev/stdout | jq .
- rm -f tyk-copy.tar
- wget http://localhost:8080/data.zip -O tyk-copy.tar
- sha256sum tyk-*.tar
- curl http://localhost:8080 | jq .
```
