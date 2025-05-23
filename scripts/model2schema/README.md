# Model verification against swagger schema

The motivation behind this is to validate the data model against the
swagger specification, validating that all fields are covered in schema.

> This tool has been deprecated, kept around for generic purposes.
> Go to cmd/schema-gen and look at the jsonschema command.

## Development usage

To use the task stand alone, install and run it like so:

```
go install github.com/titpetric/exp/scripts/model2schema@main
```

In the folder where you want to get the definitions:

```
schema-gen extract
model2schema schema.json <names...>
```

If no names are provided, all the symbols are output.


# How is this tested?

Manual tests exists in this folder. They validate the definitions as
configured in the taskfile. The validations may be extended to support
work in progress.

Specifically what it does is:

- using `yq`, extracts existing definitions from data/gateway-swagger.json
- using `schema-gen` produced `data/gateway-user.json` and `data/gateway-apidef.json`
- implements rendering an openapi definition from the schema-gen outputs
- using `dyff` to compare yaml outputs and note differences

To see the report, just run `task`. You may use `task install` to install
dyff. On occasion, you may need to run `task update` to ensure the latest
dependencies are in use.

## Example output

The following shows issues with:

- not keeping naming up to date
- invalid types, number when only integers are valid
- naming, consistency, outdated docs issues

```
Policy.properties._id.$ref
  ± value change
    - #/components/schemas/ObjectId
    + #/components/schemas/ObjectID

Policy.properties.throttle_retry_limit.type
  ± value change
    - number
    + integer

Policy.properties.max_query_depth.type
  ± value change
    - number
    + integer

Policy.properties.key_expires_in.type
  ± value change
    - number
    + integer

Policy.properties.last_updated.x-go-name
  ± value change
    - LastUpdates
    + LastUpdated

Policy.properties.graphql_access_rights
  - one map entry removed:                               + three map entries added:
    $ref: #/components/schemas/GraphAccessDefinition       type: object
                                                           x-go-name: GraphQL
                                                           additionalProperties:
                                                             $ref: #/components/schemas/GraphAccessDefinition

SessionState.properties
  + six map entries added:
    :
      type: string
      x-go-name: KeyID
    date_created:
      type: string
      format: date-time
      x-go-name: DateCreated
    enable_detailed_recording:
      type: boolean
      x-go-name: EnableDetailedRecording
    enable_http_signature_validation:
      type: boolean
      x-go-name: EnableHTTPSignatureValidation
    max_query_depth:
      type: integer
      x-go-name: MaxQueryDepth
    rsa_certificate_id:
      type: string
      x-go-name: RSACertificateId

SessionState.properties.basic_auth_data
  - one map entry removed:                      + one map entry added:
    properties:                                   $ref: #/components/schemas/BasicAuthData
      hash_type:
        $ref: #/components/schemas/HashType
      password:
        type: string
        x-go-name: Password

SessionState.properties.enable_detail_recording
  + one map entry added:
    description: |
      Deprecated: EnableDetailRecording is deprecated. Use EnableDetailedRecording
      going forward instead

SessionState.properties.enable_detail_recording.x-go-name
  ± value change
    - EnableDetailedRecording
    + EnableDetailRecording

SessionState.properties.jwt_data
  - one map entry removed:     + one map entry added:
    properties:                  $ref: #/components/schemas/JWTData
      secret:
        type: string
        x-go-name: Secret

SessionState.properties.meta_data
  - one map entry removed:
    additionalProperties:
      type: object

SessionState.properties.monitor
  - one map entry removed:           + one map entry added:
    properties:                        $ref: #/components/schemas/Monitor
      trigger_limits:
        type: array
        x-go-name: TriggerLimits
        items:
          type: number
          format: double

SessionState.properties.throttle_retry_limit
  - one map entry removed:
    format: int64
```
