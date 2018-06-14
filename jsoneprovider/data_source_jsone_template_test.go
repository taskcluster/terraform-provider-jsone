package jsoneprovider

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ghodss/yaml" // Use this because of https://github.com/go-yaml/yaml/issues/139
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/renstrom/dedent"
	"github.com/stretchr/testify/require"
)

var testProviders = map[string]terraform.ResourceProvider{
	"jsone": Provider(),
}

func TestJsoneTemplateRendering(t *testing.T) {
	var cases = []struct {
		context     string
		yamlContext string
		template    string
		want        string
		format      string
	}{
		{`{something="hello"}`, ``, `{"foo": "$${something}"}`, `{"foo": "hello"}`, `json`},
		{`{something="hello"}`, ``, `{"foo": "$${something}"}`, `{"foo": "hello"}`, `yaml`},
		{`{something="hello"}`, ``, `{"foo": "$${something}"}`, `foo: hello`, `yaml`},
		{`{something="hello"}`, ``, `foo: $${something}`, `foo: hello`, `yaml`},
		{
			``,
			`a: 1`,
			`baz: {$$map: [123, 456], each(x): {$$eval: 'x + a'}}`,
			`baz: [124, 457]`,
			`yaml`,
		},
		{
			``,
			`{\"foobar\": [123, 456], \"a\": 1}`,
			`baz: {$$map: {$$eval: foobar}, each(x): {$$eval: 'x + a'}}`,
			`baz: [124, 457]`,
			`yaml`,
		},
		{
			``,
			``,
			`baz: {$$map: [hey, whoa], each(x): '$$$${base64encode($${x})}'}`,
			`baz: ["${base64encode(hey)}", "${base64encode(whoa)}"]`,
			`yaml`,
		},
	}

	for _, tt := range cases {
		resource.UnitTest(t, resource.TestCase{
			Providers: testProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: testTemplateConfig("jsone_template", tt.template, tt.context, tt.yamlContext, tt.format),
					Check: func(s *terraform.State) error {
						got := s.RootModule().Outputs["rendered"]
						value := got.Value.(string)
						want := make(map[string]interface{})
						result := make(map[string]interface{})
						if tt.format == "json" {
							json.Unmarshal([]byte(tt.want), &want)
							json.Unmarshal([]byte(value), &result)
						} else {
							yaml.Unmarshal([]byte(tt.want), &want)
							yaml.Unmarshal([]byte(value), &result)
						}

						require.Equal(t, want, result, fmt.Sprintf("template:\n%s\ncontext:\n%s\ngot:\n%s\nwant:\n%s\n", tt.template, tt.context, got, tt.want))
						return nil
					},
				},
			},
		})
	}
}

func TestJsoneTemplatesRendering(t *testing.T) {
	var cases = []struct {
		context     string
		yamlContext string
		template    string
		want        []string
		format      string
	}{
		{`{one="1", two="2"}`, ``, dedent.Dedent(`
		---
		{"document": "$${one}"}
		... # l-document-suffix
		# comment
		---
		{"document": "$${two}"}
		---
		{"document": "3"}
		`), []string{
			`{"document": "1"}`,
			`{"document": "2"}`,
			`{"document": "3"}`,
		}, `json`},
	}

	for _, tt := range cases {
		resource.UnitTest(t, resource.TestCase{
			Providers: testProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: testTemplateConfig("jsone_templates", tt.template, tt.context, tt.yamlContext, tt.format),
					Check: func(s *terraform.State) error {
						got := s.RootModule().Outputs["rendered"]
						value := got.Value.([]interface{})

						require.Equal(t, len(value), len(tt.want))
						for i := range value {
							want := make(map[string]interface{})
							result := make(map[string]interface{})
							if tt.format == "json" {
								json.Unmarshal([]byte(tt.want[i]), &want)
								json.Unmarshal([]byte(value[i].(string)), &result)
							} else {
								yaml.Unmarshal([]byte(tt.want[i]), &want)
								yaml.Unmarshal([]byte(value[i].(string)), &result)
							}

							require.Equal(t, want, result, fmt.Sprintf("template:\n%s\ncontext:\n%s\ngot:\n%s\nwant:\n%s\n", tt.template, tt.context, got, tt.want))
						}
						return nil
					},
				},
			},
		})
	}
}

func testTemplateConfig(resource, template, context, yamlContext, format string) string {
	if yamlContext != "" {
		return fmt.Sprintf(`
			data "%s" "t0" {
				template = <<EOF
%s
EOF
				yaml_context = "%s"
				format = "%s"
			}
			output "rendered" {
				value = "${data.%s.t0.rendered}"
			}`, resource, template, yamlContext, format, resource)
	} else if context != "" {
		return fmt.Sprintf(`
			data "%s" "t0" {
				template = <<EOF
%s
EOF
				context = %s
				format = "%s"
			}
			output "rendered" {
				value = "${data.%s.t0.rendered}"
			}`, resource, template, context, format, resource)
	} else {
		return fmt.Sprintf(`
			data "%s" "t0" {
				template = <<EOF
%s
EOF
				format = "%s"
			}
			output "rendered" {
				value = "${data.%s.t0.rendered}"
			}`, resource, template, format, resource)
	}
}
