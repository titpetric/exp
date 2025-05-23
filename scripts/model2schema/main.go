package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/titpetric/exp/cmd/schema-gen/model"
)

func main() {
	if err := start(); err != nil {
		log.Fatal(err)
	}
}

func start() error {
	schema, err := model.Load(os.Args[1])
	if err != nil {
		return err
	}

	models := schema[0].Declarations.Find(os.Args[2:])

	for _, model := range models {
		def := Convert("github.com/TykTechnologies/tyk/"+schema[0].Name, model)
		b, _ := yaml.Marshal(def)

		lines := strings.Split(string(b), "\n")
		for _, line := range lines {
			line = strings.ReplaceAll(line, "  ", " ")
			line = line

			fmt.Println(line)
		}
	}

	return nil
}

func Convert(goPackage string, in *model.TypeInfo) DefinitionMap {
	props := make(Properties)

	for _, field := range in.Fields {
		//Debug input:
		//spew.Dump(field)

		builtin := false
		fieldType := field.Type
		fieldFormat := ""

		p, ok := TypeAlias(fieldType)
		fieldType = p.Type
		fieldFormat = p.Format
		builtin = ok

		if fieldType == fieldFormat {
			fieldFormat = ""
		}

		name := field.JSONName
		if strings.Contains(field.JSONName, ",") {
			name = name[:strings.Index(name, ",")]
		}

		props[name] = &Property{
			Type:        fieldType,
			Format:      fieldFormat,
			GoName:      field.Name,
			Description: field.Doc,
			BuiltIn:     builtin,
		}
	}

	props.Sanitize()

	return DefinitionMap{
		in.Name: &Definition{
			Title:       title(in.Doc),
			Description: description(in.Doc),
			Properties:  props,
			Type:        "object",
			GoPackage:   goPackage,
		},
	}
}

type Definition struct {
	Properties  Properties `json:"properties" yaml:"properties"`
	Title       string     `json:"title" yaml:"title"`
	Description string     `json:"description,omitempty" yaml:"description,omitempty"`
	Type        string     `json:"type" yaml:"type"`
	GoPackage   string     `json:"x-go-package" yaml:"x-go-package"`
}

type DefinitionMap map[string]*Definition

type Properties map[string]*Property

func (p Properties) Sanitize() {
	for _, i := range p {
		i.Sanitize()
	}
}

type Property struct {
	Type        string `json:"type,omitempty" yaml:"type,omitempty"`
	Items       *Items `json:"items,omitempty" yaml:"items,omitempty"`
	Ref         string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Format      string `json:"format,omitempty" yaml:"format,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	GoName      string `json:"x-go-name" yaml:"x-go-name"`

	AdditionalProperties *Items `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`

	BuiltIn bool `json:"-" yaml:"-"`
}

type Items struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	Ref  string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

func (p *Property) Sanitize() {
	switch p.Type {
	case "map":
		format := p.Format
		if strings.Contains(format, ".") {
			format = strings.Split(format, ".")[1]
		}

		format, custom := toType(format)

		p.AdditionalProperties = &Items{
			Type: format,
		}

		if custom {
			p.AdditionalProperties = &Items{
				Ref: "#/components/schemas/" + format,
			}
		}
		p.Format = ""
		p.Type = "object"

	case "object":
		format := p.Format
		if strings.Contains(format, ".") {
			format = strings.Split(format, ".")[1]
		}
		if !p.BuiltIn {
			p.Ref = "#/components/schemas/" + format
			p.Format = ""
		}
	case "array":
		format := p.Format
		if strings.Contains(format, ".") {
			format = strings.Split(format, ".")[1]
		}
		if !p.BuiltIn {
			p.Items = &Items{
				Ref: "#/components/schemas/" + format,
			}
		} else {
			p.Items = &Items{
				Type: format,
			}
		}
		p.Format = ""
	default:
		if !p.BuiltIn {
			format := p.Type
			if strings.Contains(format, ".") {
				format = strings.Split(format, ".")[1]
			}
			p.Ref = "#/components/schemas/" + format
			p.Type = "object"
			p.Format = ""
		}
	}

}

/*
   Policy:
     properties:
       _id:
         $ref: '#/components/schemas/ObjectId'
         type: object
         x-go-name: ID
       id:
         type: string
         x-go-name: ID
       name:
         type: string
         x-go-name: Name
       org_id:
         type: string
         x-go-name: OrgID
       rate:
         format: double
         type: number
         x-go-name: Rate
       per:
         format: double
         type: number
         x-go-name: Per
       quota_max:
         format: int64
         type: integer
         x-go-name: QuotaMax
       quota_renewal_rate:
         format: int64
         type: integer
         x-go-name: QuotaRenewalRate
       throttle_interval:
         format: double
         type: number
         x-go-name: ThrottleInterval
       throttle_retry_limit:
         type: number
         x-go-name: ThrottleRetryLimit
       max_query_depth:
         type: number
         x-go-name: MaxQueryDepth
       access_rights:
         $ref: '#/components/schemas/AccessDefinition'
         type: object
         x-go-name: AccessRights
       hmac_enabled:
         type: boolean
         x-go-name: HMACEnabled
       enable_http_signature_validation:
         type: boolean
         x-go-name: EnableHTTPSignatureValidation
       active:
         type: boolean
         x-go-name: Active
       is_inactive:
         type: boolean
         x-go-name: IsInactive
       tags:
         type: array
         items:
           type: string
         x-go-name: Tags
       key_expires_in:
         format: int64
         type: number
         x-go-name: KeyExpiresIn
       partitions:
         $ref: '#/components/schemas/PolicyPartitions'
         type: object
         x-go-name: Partitions
       last_updated:
         type: string
         x-go-name: LastUpdates
       meta_data:
         type: object
         x-go-name: MetaData
       graphql_access_rights:
         $ref: '#/components/schemas/GraphAccessDefinition'
     type: object
     x-go-package: github.com/TykTechnologies/tyk/user
*/
