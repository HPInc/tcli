package parser

import (
	"os"
	"strings"

	"github.com/hpinc/tcli/pkg/common"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v2 "github.com/pb33f/libopenapi/datamodel/high/v2"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"gopkg.in/yaml.v3"
)

type extNode struct {
	Content []*extNode `yaml:"content"`
	Value   string     `yaml:"value"`
}

func extractClassFromNodeValue(val any) string {
	b, _ := yaml.Marshal(val)
	var n extNode
	_ = yaml.Unmarshal(b, &n)
	if len(n.Content) > 0 {
		for i := 0; i < len(n.Content)-1; i += 2 {
			if n.Content[i].Value == "class" {
				return n.Content[i+1].Value
			}
		}
	}
	return ""
}

func ReadSwagger(f string) (*Root, error) {
	bytes, err := os.ReadFile(f) // #nosec G304
	if err != nil {
		return nil, err
	}

	doc, err := libopenapi.NewDocument(bytes)
	if err != nil {
		return nil, err
	}

	m2, errs2 := doc.BuildV2Model()
	if errs2 == nil && m2 != nil {
		return buildFromV2(&m2.Model), nil
	}

	m3, errs3 := doc.BuildV3Model()
	if errs3 == nil && m3 != nil {
		return buildFromV3(&m3.Model), nil
	}

	// if both failed, try returning the v2 or v3 error
	if errs2 != nil {
		return nil, errs2
	}
	if errs3 != nil {
		return nil, errs3
	}

	return nil, common.ErrNotFound
}

func buildFromV2(doc *v2.Swagger) *Root {
	r := &Root{
		Paths:               make(map[string]*Path),
		Definitions:         make(map[string]Definition),
		SecurityDefinitions: make(map[string]SecurityDefinition),
	}

	if doc.Host != "" {
		r.Host = doc.Host
	}
	if doc.BasePath != "" {
		r.BasePath = doc.BasePath
	}
	r.Schemes = doc.Schemes

	for _, tag := range doc.Tags {
		r.Tags = append(r.Tags, Tag{
			Name:        tag.Name,
			Description: tag.Description,
		})
	}

	if doc.SecurityDefinitions != nil && doc.SecurityDefinitions.Definitions != nil {
		for pair := doc.SecurityDefinitions.Definitions.First(); pair != nil; pair = pair.Next() {
			if pair.Value() != nil {
				r.SecurityDefinitions[pair.Key()] = SecurityDefinition{Type: pair.Value().Type}
			}
		}
	}

	if doc.Paths != nil && doc.Paths.PathItems != nil {
		for pair := doc.Paths.PathItems.First(); pair != nil; pair = pair.Next() {
			p := pair.Value()
			pathObj := &Path{
				Get:     mapV2Method(p.Get, pair.Key(), "GET"),
				Put:     mapV2Method(p.Put, pair.Key(), "PUT"),
				Post:    mapV2Method(p.Post, pair.Key(), "POST"),
				Delete:  mapV2Method(p.Delete, pair.Key(), "DELETE"),
				Options: mapV2Method(p.Options, pair.Key(), "OPTIONS"),
				Patch:   mapV2Method(p.Patch, pair.Key(), "PATCH"),
			}
			r.Paths[pair.Key()] = pathObj
		}
	}

	if doc.Definitions != nil && doc.Definitions.Definitions != nil {
		for pair := doc.Definitions.Definitions.First(); pair != nil; pair = pair.Next() {
			r.Definitions[pair.Key()] = mapV2Definition(pair.Value())
		}
	}

	return r
}

func mapV2Method(op *v2.Operation, pathStr, methodStr string) *Method {
	if op == nil {
		return nil
	}

	m := &Method{
		Summary:     op.Summary,
		Description: op.Description,
		OperationId: op.OperationId,
		Tags:        op.Tags,
		MethodName:  methodStr,
		Consumes:    op.Consumes,
		Path:        pathStr,
	}

	for _, param := range op.Parameters {
		p := Parameter{
			Name:        param.Name,
			Description: param.Description,
			In:          param.In,
			Type:        param.Type,
			Format:      param.Format,
		}
		if param.Required != nil {
			p.Required = *param.Required
		}
		if param.Default != nil {
			p.Default = param.Default.Value
		}
		if param.Schema != nil {
			p.Schema = &Schema{Ref: param.Schema.GetReference()}
		}
		m.Parameters = append(m.Parameters, p)
	}

	m.Securities = make([]map[string][]string, 0)
	for _, req := range op.Security {
		secMap := make(map[string][]string)
		for reqPair := req.Requirements.First(); reqPair != nil; reqPair = reqPair.Next() {
			secMap[reqPair.Key()] = reqPair.Value()
		}
		m.Securities = append(m.Securities, secMap)
	}

	if val, ok := op.Extensions.Get("x-extension"); ok && val != nil {
		if cls := extractClassFromNodeValue(val); cls != "" {
			m.Extension = &Extension{Class: cls}
		}
	}

	return m
}

func mapV2Definition(schemaProxy *base.SchemaProxy) Definition {
	def := Definition{
		Properties: make(map[string]*Property),
		Type:       "object",
	}
	if schemaProxy == nil {
		return def
	}
	schema := schemaProxy.Schema()
	if schema != nil {
		if len(schema.Type) > 0 {
			def.Type = schema.Type[0]
		}
		def.Required = schema.Required
		if schema.Properties != nil {
			for propPair := schema.Properties.First(); propPair != nil; propPair = propPair.Next() {
				propSchema := propPair.Value().Schema()
				prop := &Property{}
				if propSchema != nil {
					if len(propSchema.Type) > 0 {
						prop.Type = propSchema.Type[0]
					}
					prop.Format = propSchema.Format
					prop.Description = propSchema.Description
					if propPair.Value().GetReference() != "" {
						prop.Ref = propPair.Value().GetReference()
					}
					if propSchema.Items != nil && propSchema.Items.A != nil {
						itemSchema := propSchema.Items.A.Schema()
						if itemSchema != nil {
							itemType := ""
							if len(itemSchema.Type) > 0 {
								itemType = itemSchema.Type[0]
							}
							prop.Items = &ArrayItem{
								Type:   itemType,
								Format: itemSchema.Format,
								Ref:    propSchema.Items.A.GetReference(),
							}
						}
					}
				} else if propPair.Value().GetReference() != "" {
					prop.Ref = propPair.Value().GetReference()
				}
				def.Properties[propPair.Key()] = prop
			}
		}
	}
	return def
}

func buildFromV3(doc *v3.Document) *Root {
	r := &Root{
		Paths:               make(map[string]*Path),
		Definitions:         make(map[string]Definition),
		SecurityDefinitions: make(map[string]SecurityDefinition),
	}

	if doc.Servers != nil && len(doc.Servers) > 0 {
		srv := doc.Servers[0]
		r.Host = strings.TrimPrefix(strings.TrimPrefix(srv.URL, "https://"), "http://")
		// rudimentary parsing of host/basePath from URL if absolute
		idx := strings.Index(r.Host, "/")
		if idx > 0 {
			r.BasePath = r.Host[idx:]
			r.Host = r.Host[:idx]
		}
		if strings.HasPrefix(srv.URL, "https") {
			r.Schemes = []string{"https"}
		} else if strings.HasPrefix(srv.URL, "http") {
			r.Schemes = []string{"http"}
		}
	}

	for _, tag := range doc.Tags {
		r.Tags = append(r.Tags, Tag{
			Name:        tag.Name,
			Description: tag.Description,
		})
	}

	if doc.Components != nil && doc.Components.SecuritySchemes != nil {
		for pair := doc.Components.SecuritySchemes.First(); pair != nil; pair = pair.Next() {
			if pair.Value() != nil {
				r.SecurityDefinitions[pair.Key()] = SecurityDefinition{Type: pair.Value().Type}
			}
		}
	}

	if doc.Paths != nil && doc.Paths.PathItems != nil {
		for pair := doc.Paths.PathItems.First(); pair != nil; pair = pair.Next() {
			p := pair.Value()
			pathObj := &Path{
				Get:     mapV3Method(p.Get, pair.Key(), "GET"),
				Put:     mapV3Method(p.Put, pair.Key(), "PUT"),
				Post:    mapV3Method(p.Post, pair.Key(), "POST"),
				Delete:  mapV3Method(p.Delete, pair.Key(), "DELETE"),
				Options: mapV3Method(p.Options, pair.Key(), "OPTIONS"),
				Patch:   mapV3Method(p.Patch, pair.Key(), "PATCH"),
			}
			r.Paths[pair.Key()] = pathObj
		}
	}

	if doc.Components != nil && doc.Components.Schemas != nil {
		for pair := doc.Components.Schemas.First(); pair != nil; pair = pair.Next() {
			r.Definitions[pair.Key()] = mapV3Definition(pair.Value())
		}
	}

	return r
}

func mapV3Method(op *v3.Operation, pathStr, methodStr string) *Method {
	if op == nil {
		return nil
	}

	m := &Method{
		Summary:     op.Summary,
		Description: op.Description,
		OperationId: op.OperationId,
		Tags:        op.Tags,
		MethodName:  methodStr,
		Path:        pathStr,
	}

	if op.RequestBody != nil && op.RequestBody.Content != nil {
		for pair := op.RequestBody.Content.First(); pair != nil; pair = pair.Next() {
			m.Consumes = append(m.Consumes, pair.Key())
			// V3 RequestBody becomes a "body" parameter for backwards compatibility with cmd pkg
			param := Parameter{
				Name:        "body",
				Description: op.RequestBody.Description,
				In:          "body",
			}
			if op.RequestBody.Required != nil {
				param.Required = *op.RequestBody.Required
			}
			mediaType := pair.Value()
			if mediaType.Schema != nil {
				param.Schema = &Schema{Ref: mediaType.Schema.GetReference()}
			}
			m.Parameters = append(m.Parameters, param)
			break // just take the first content type mapping
		}
	}

	for _, paramProxy := range op.Parameters {
		param := paramProxy
		p := Parameter{
			Name:        param.Name,
			Description: param.Description,
			In:          param.In,
		}
		if param.Required != nil {
			p.Required = *param.Required
		}
		if param.Schema != nil {
			schema := param.Schema.Schema()
			if schema != nil && len(schema.Type) > 0 {
				p.Type = schema.Type[0]
				p.Format = schema.Format
				if schema.Default != nil {
					p.Default = schema.Default.Value
				}
			}
			// if schema is a ref
			if param.Schema.GetReference() != "" {
				p.Schema = &Schema{Ref: param.Schema.GetReference()}
			}
		}
		m.Parameters = append(m.Parameters, p)
	}

	m.Securities = make([]map[string][]string, 0)
	for _, req := range op.Security {
		secMap := make(map[string][]string)
		for reqPair := req.Requirements.First(); reqPair != nil; reqPair = reqPair.Next() {
			secMap[reqPair.Key()] = reqPair.Value()
		}
		m.Securities = append(m.Securities, secMap)
	}

	if val, ok := op.Extensions.Get("x-extension"); ok && val != nil {
		if cls := extractClassFromNodeValue(val); cls != "" {
			m.Extension = &Extension{Class: cls}
		}
	}

	return m
}

func mapV3Definition(schemaProxy *base.SchemaProxy) Definition {
	def := Definition{
		Properties: make(map[string]*Property),
		Type:       "object",
	}
	if schemaProxy == nil {
		return def
	}
	schema := schemaProxy.Schema()
	if schema != nil {
		if len(schema.Type) > 0 {
			def.Type = schema.Type[0]
		}
		def.Required = schema.Required
		if schema.Properties != nil {
			for propPair := schema.Properties.First(); propPair != nil; propPair = propPair.Next() {
				propSchemaProxy := propPair.Value()
				propSchema := propSchemaProxy.Schema()
				prop := &Property{}
				if propSchema != nil {
					if len(propSchema.Type) > 0 {
						prop.Type = propSchema.Type[0]
					}
					prop.Format = propSchema.Format
					prop.Description = propSchema.Description
					if propSchemaProxy.GetReference() != "" {
						prop.Ref = propSchemaProxy.GetReference()
					}
					if propSchema.Items != nil && propSchema.Items.A != nil {
						itemSchemaProxy := propSchema.Items.A
						itemSchema := itemSchemaProxy.Schema()
						if itemSchema != nil {
							itemType := ""
							if len(itemSchema.Type) > 0 {
								itemType = itemSchema.Type[0]
							}
							prop.Items = &ArrayItem{
								Type:   itemType,
								Format: itemSchema.Format,
								Ref:    itemSchemaProxy.GetReference(),
							}
						}
					}
				} else if propSchemaProxy.GetReference() != "" {
					prop.Ref = propSchemaProxy.GetReference()
				}
				def.Properties[propPair.Key()] = prop
			}
		}
	}
	return def
}
