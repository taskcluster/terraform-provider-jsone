package jsoneprovider

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	ghodssYaml "github.com/ghodss/yaml" // Use this because of https://github.com/go-yaml/yaml/issues/139
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	jsone "github.com/taskcluster/json-e"
	jsoneInterpreter "github.com/taskcluster/json-e/interpreter"
	"github.com/taskcluster/terraform-provider-jsone/yaml"
)

var templateSchema = schema.Schema{
	Type:        schema.TypeString,
	Required:    true,
	Description: "Contents of the template. Must be valid json or yaml.",
}
var formatSchema = schema.Schema{
	Type:         schema.TypeString,
	Optional:     true,
	Default:      "json",
	Description:  "Output format. Choose either json or yaml.",
	ValidateFunc: validation.StringInSlice([]string{"json", "yaml"}, false),
}
var contextSchema = schema.Schema{
	Type:          schema.TypeMap,
	ConflictsWith: []string{"yaml_context"},
	Optional:      true,
	Default:       make(map[string]interface{}),
	Description:   "json-e context variables. This is convenient hcl syntax version if you don't need types",
}
var yamlContextSchema = schema.Schema{
	Type:          schema.TypeString,
	ConflictsWith: []string{"context"},
	Optional:      true,
	Default:       "",
	Description:   "json-e context variables. Pass in context as yaml if you need numbers or booleans.",
}

func dataSourceJsoneTemplate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceJsoneTemplateRead,

		Schema: map[string]*schema.Schema{
			"template":     &templateSchema,
			"format":       &formatSchema,
			"context":      &contextSchema,
			"yaml_context": &yamlContextSchema,
			"rendered": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "json-e rendered string in format chosen in format.",
			},
		},
	}
}

func dataSourceJsoneTemplates() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceJsoneTemplatesRead,

		Schema: map[string]*schema.Schema{
			"template":     &templateSchema,
			"format":       &formatSchema,
			"context":      &contextSchema,
			"yaml_context": &yamlContextSchema,
			"rendered": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "json-e rendered strings, one per yaml document, in format chosen in format.",
			},
		},
	}
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

func readCommon(d *schema.ResourceData) ([]string, error) {
	template := d.Get("template").(string)
	format := d.Get("format").(string)
	ctx := d.Get("context").(map[string]interface{})
	yamlContext := d.Get("yaml_context").(string)

	if yamlContext != "" {
		ghodssYaml.Unmarshal([]byte(yamlContext), &ctx)
	}

	context := map[string]interface{}{
		"base64encode": jsoneInterpreter.WrapFunction(func (s string) (string) {
			return base64.StdEncoding.EncodeToString([]byte(s))
		}),
	}

	for k, v := range ctx {
		context[k] = v
	}

	dec := yaml.NewDecoder(bytes.NewReader([]byte(template)))
	allDocs := make([]string, 0)

	for {
		t := make(map[string]interface{})
		err := dec.Decode(&t)
		if err != nil {
			if err == io.EOF {
				break
			}
			return []string{}, err
		}

		result, err := jsone.Render(t, context)
		if err != nil {
			return []string{}, err
		}

		var m []byte
		if format == "json" {
			m, err = json.Marshal(result)
		} else {
			m, err = ghodssYaml.Marshal(result)
		}
		if err != nil {
			return []string{}, err
		}

		allDocs = append(allDocs, string(m))
	}

	return allDocs, nil
}

func dataSourceJsoneTemplateRead(d *schema.ResourceData, meta interface{}) error {
	rendered, err := readCommon(d)
	if err != nil {
		return err
	}

	if len(rendered) != 1 {
		return fmt.Errorf("YAML template contained more than one document (--- separator)")
	}

	d.Set("rendered", rendered[0])
	d.SetId(hash(rendered[0]))
	return nil
}

func dataSourceJsoneTemplatesRead(d *schema.ResourceData, meta interface{}) error {
	rendered, err := readCommon(d)
	if err != nil {
		return err
	}

	d.Set("rendered", rendered)
	d.SetId(hash(strings.Join(rendered, "\n")))
	return nil
}
