# JSON-e Terraform Provider

A provider with a single resource `jsone_template`. This works similarly to the `template_file` resource
but allows us to use json-e constructs like $map to transforms lists and such things.

You can use [the json-e documentation](https://taskcluster.github.io/json-e/) to find out more about
how to use the templates.

## Usage

First `go get -u github.com/taskcluster/terraform-provider-jsone`. At that point, if you have built terraform from
source, simply running `go install` for this package will work. Otherwise, follow along with
[the official docs](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins) on how to use
third party plugins.

Now it is as simple as:

```terraform
locals {
  context = {
    foo = "bar"
    baz = ["bing", "qux"]
  }
}

data "jsone_template" "service" {
  template = "${file("${path.module}/service.yaml")}"
  yaml_context = "${jsonencode(local.context)}"
}

output "rendered" {
  value = "${data.jsone_template.service.rendered}"
}
```

The resource takes a string called `template` and either a field called `context` or `yaml_context`.
These are mutually exclusive. `context` should be good enough for most cases and is just a terraform map,
but if you run into issues with how terraform handles mixed-type maps or need to pass in typed data, opt
for `yaml_context` which just takes a string containing yaml. Oftentimes using the terraform interpolation
of `jsonencode` will be your best bet.

You can access the rendered template with `.rendered` much the same as with `template_file`. This will normally
be in `json` format unless you explicitely pass in `format = "yaml"` to the resource.

### Multiple Templates

In some cases, it is useful to have multiple templates in the input file.
In this case, the `rendered` value is a list of JSON- (or YAML-) formatted rendered templates.
Use the `jsone_templates` resource (note the plural) for this purpose.

```yaml
# service.yaml
---
one: '{$eval: "17 - 16"}'
---
two: '{$eval: "12 / 6"}'
```

```terraform
data "jsone_templates" "service" {
  template = "${file("${path.module}/service.yaml")}"
  yaml_context = "${jsonencode(local.context)}"
}

output "rendered" {
  value1 = "${data.jsone_template.service.rendered[0]}"
  # --> '{"one": 1}'
  value2 = "${data.jsone_template.service.rendered[1]}"
  # --> '{"two": 2}'
}
```

# Development

Go requirements are the same as Terraform itself (currently 1.9).

In a clean GOPATH, `go get github.com/taskcluster/terraform-provider-jsone`, then switch to that directory and run `go get -t ./...` to get dependencies.
Run `go test ./..` to test.
