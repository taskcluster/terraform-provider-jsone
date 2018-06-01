package jsoneprovider

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/ghodss/yaml" // Use this because of https://github.com/go-yaml/yaml/issues/139
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/taskcluster/json-e"
)

func dataSourceJsoneTemplate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceJsoneTemplateRead,

		Schema: map[string]*schema.Schema{
			"template": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Contents of the template. Must be valid json or yaml.",
			},
			"format": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "json",
				Description:  "Output format. Choose either json or yaml.",
				ValidateFunc: validation.StringInSlice([]string{"json", "yaml"}, false),
			},
			"context": &schema.Schema{
				Type:          schema.TypeMap,
				ConflictsWith: []string{"yaml_context"},
				Optional:      true,
				Default:       make(map[string]interface{}),
				Description:   "json-e context variables. This is convenient hcl syntax version if you don't need types",
			},
			"yaml_context": &schema.Schema{
				Type:          schema.TypeString,
				ConflictsWith: []string{"context"},
				Optional:      true,
				Default:       "",
				Description:   "json-e context variables. Pass in context as yaml if you need numbers or booleans.",
			},
			"rendered": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "json-e rendered string in format chosen in format.",
			},
		},
	}
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

func dataSourceJsoneTemplateRead(d *schema.ResourceData, meta interface{}) error {
	template := d.Get("template").(string)
	format := d.Get("format").(string)
	context := d.Get("context").(map[string]interface{})
	yamlContext := d.Get("yaml_context").(string)

	if yamlContext != "" {
		yaml.Unmarshal([]byte(yamlContext), &context)
	}

	t := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(template), &t)
	if err != nil {
		return err
	}

	result, err := jsone.Render(t, context)
	if err != nil {
		return err
	}

	var m []byte
	if format == "json" {
		m, err = json.Marshal(result)
	} else {
		m, err = yaml.Marshal(result)
	}
	if err != nil {
		return err
	}

	rendered := string(m)
	d.Set("rendered", rendered)
	d.SetId(hash(rendered))
	return nil
}
