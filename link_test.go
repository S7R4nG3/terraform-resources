package tfresources

import (
	"testing"

	"github.com/go-test/deep"
	tfjson "github.com/hashicorp/terraform-json"
)

func TestLinkResourcesWithModules(t *testing.T) {
	tests := []struct {
		name      string
		modules   []Module
		resources []tfjson.StateResource
		want      []Resource
	}{
		{
			name: "Resources should link to modules properly",
			modules: []Module{
				{
					Key:     "function",
					Source:  "some.registry.source",
					Dir:     "a/magic/dir",
					Version: "0.0.0",
				},
			},
			resources: []tfjson.StateResource{
				{
					Address: "module.function.aws_lambda_function.default",
					Type:    "aws_lambda_function",
					Name:    "default",
				},
			},
			want: []Resource{
				{
					Module: Module{
						Key:     "function",
						Source:  "some.registry.source",
						Dir:     "a/magic/dir",
						Version: "0.0.0",
					},
					Planned: tfjson.StateResource{
						Address: "module.function.aws_lambda_function.default",
						Type:    "aws_lambda_function",
						Name:    "default",
					},
				},
			},
		},
		{
			name: "Resources should not link to modules properly",
			modules: []Module{
				{
					Key:     "function",
					Source:  "another.registry.surce",
					Dir:     "a/magic/dir",
					Version: "1.1.1",
				},
			},
			resources: []tfjson.StateResource{
				{
					Address: "aws_lambda_function.worker",
				},
			},
			want: []Resource{
				{
					Module: Module{},
					Planned: tfjson.StateResource{
						Address: "aws_lambda_function.worker",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Logf("Running test - %s", tt.name)
		plan := Plan{}
		plan.linkResourcesWithModules(tt.modules, tt.resources)
		diff := deep.Equal(plan.Resources, tt.want)
		if diff != nil {
			t.Fatal(testResults(tt.name, diff))
		}
	}
}

func TestParseModuleFromResourceAddress(t *testing.T) {
	tests := []struct {
		name     string
		resource tfjson.StateResource
		modules  []Module
		want     Module
	}{
		{
			name: "Resource address should link to module key",
			resource: tfjson.StateResource{
				Address: "module.iam_role.aws_iam_role.this[0]",
				Type:    "aws_iam_role",
				Name:    "this",
			},
			modules: []Module{
				{
					Key: "a.non.matching.key",
				},
				{
					Key:    "iam_role",
					Source: "registry.terraform.io",
					Dir:    "a/magic/dir",
				},
			},
			want: Module{
				Key:    "iam_role",
				Source: "registry.terraform.io",
				Dir:    "a/magic/dir",
			},
		},
		{
			name: "Non-Module resources do not link to a module",
			resource: tfjson.StateResource{
				Address: "aws_iam_role.lambda",
				Type:    "aws_iam_role",
				Name:    "lambda",
			},
			modules: []Module{
				{
					Key:    "function",
					Source: "registry.terraform.io/magic/module",
					Dir:    "./terraform/modules/magic",
				},
				{
					Key:    "bucket",
					Source: "registry.terraform.io/magic/bucket",
					Dir:    "./terraform/modules/bucket",
				},
				{
					Key:    "iam_role",
					Source: "registry.terraform.io/magic/role",
					Dir:    "./terraform/modules/role",
				},
			},
			want: Module{},
		},
	}

	for _, tt := range tests {
		t.Logf("Running test -- %s", tt.name)
		got := parseModuleFromResourceAddress(tt.resource, tt.modules)
		diff := deep.Equal(got, tt.want)
		if diff != nil {
			t.Fatalf(testResults(tt.name, diff))
		}
	}
}

func TestParseParentFromChildModule(t *testing.T) {
	tests := []struct {
		name    string
		child   Module
		modules []Module
		want    Module
	}{
		{
			name: "Nested modules should link to their parent",
			child: Module{
				Key: "function.execution_role",
			},
			modules: []Module{
				{
					Key:    "function",
					Source: "a.magic.registry",
					Dir:    "./terraform/magic",
				},
				{
					Key:    "execution_role",
					Source: "another.magic.registry",
					Dir:    "./terraform/other/magic",
				},
				{
					Key:    "random",
					Source: "hello.isitme.yourelooking.for",
					Dir:    "./terraform/lionel/richie",
				},
			},
			want: Module{
				Key:    "function",
				Source: "a.magic.registry",
				Dir:    "./terraform/magic",
			},
		},
		{
			name: "Super nested modules should still link to their parent",
			child: Module{
				Key: "static_website.origin.access_role.policy",
			},
			modules: []Module{
				{
					Key:    "static_website",
					Source: "a.magic.registry",
					Dir:    "./terraform/magic",
				},
				{
					Key:    "origin",
					Source: "another.magic.registry",
					Dir:    "./terraform/other/magic",
				},
				{
					Key:    "access_role",
					Source: "hello.isitme.yourelooking.for",
					Dir:    "./terraform/lionel/richie",
				},
				{
					Key:    "policy",
					Source: "ill.gladly.pay.you.tuesday.for.a.hamburger.today",
					Dir:    "./terraform/wimpy",
				},
			},
			want: Module{
				Key:    "static_website",
				Source: "a.magic.registry",
				Dir:    "./terraform/magic",
			},
		},
	}

	for _, tt := range tests {
		t.Logf("Running test -- %s", tt.name)
		got := parseParentFromChildModule(tt.child, tt.modules)
		diff := deep.Equal(got, tt.want)
		if diff != nil {
			t.Fatal(testResults(tt.name, diff))
		}
	}
}
