# Scan type and field usage

In large package scopes and large codebases it's hard to reason about
which fields are in use where. This tool approaches this with AST,
scanning the code base to find a particular type name, and a field
that's being referenced. This has a few benefits:

- We can show exactly which config fields would get evaluated where (impact)
- We can show exactly where a typed field is used (coupling)
- We can detect if a typed field is unused (dead code)

## Install

Install the tool like so:

```bash
go get github.com/titpetric/exp/cmd/discover@main
```

Usage is: `discover <StructName> <FieldName>`.

## Examples

- Gateway API definition defines an AuthConfig type.
- That type has a Signature field.

To answer where the type is used, run:

```bash
discover AuthConfig Struct
```

The program outputs `file.go <line> <trimmed-source>` as output, showing
detected references to the requested symbol.

```text
$ discover AuthConfig Signature
apidef/oas/authentication.go 442 signature := authConfig.Signature
gateway/mw_auth_key.go 208 if authConfig.Signature.ErrorCode != 0 {
gateway/mw_auth_key.go 209 errorCode = authConfig.Signature.ErrorCode
gateway/mw_auth_key.go 213 if authConfig.Signature.ErrorMessage != "" {
gateway/mw_auth_key.go 214 errorMessage = authConfig.Signature.ErrorMessage
gateway/mw_auth_key.go 218 if err := validator.Init(authConfig.Signature.Algorithm); err != nil {
gateway/mw_auth_key.go 223 signature := r.Header.Get(authConfig.Signature.Header)
gateway/mw_auth_key.go 225 paramName := authConfig.Signature.ParamName
gateway/mw_auth_key.go 226 if authConfig.Signature.UseParam || paramName != "" {
gateway/mw_auth_key.go 228 paramName = authConfig.Signature.Header
gateway/mw_auth_key.go 244 secret := k.Gw.ReplaceTykVariables(r, authConfig.Signature.Secret, false)
gateway/mw_auth_key.go 251 if err := validator.Validate(signature, key, secret, authConfig.Signature.AllowedClockSkew); err != nil {
```

## Authors note

This is mainly usable for large, hard to navigate codebases with
extensive data models like configuration and API definitions.

Known limitations:

Technically the first argument should include the package name, as I'm
particularly interested in `apidef.AuthConfig` use. It's expected that
different packages declare their own `Config` types, and other type
names could repeat. Currently this is not supported.
