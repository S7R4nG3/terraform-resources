package tfresources

import (
	"testing"

	"github.com/go-test/deep"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/sirupsen/logrus"
)

func TestParseChanges(t *testing.T) {
	tests := []struct {
		name    string
		changes []*tfjson.ResourceChange
		want    map[string]tfjson.Change
	}{
		{
			name: "Changes should be collected by resource address",
			changes: []*tfjson.ResourceChange{
				{
					Address: "aws_s3_bucket.default",
					Change: &tfjson.Change{
						Actions:      tfjson.Actions{tfjson.ActionCreate},
						AfterUnknown: map[string]any{"arn": true},
					},
				},
				{
					Address: "module.function.aws_lambda_function.default",
					Change: &tfjson.Change{
						Actions: tfjson.Actions{tfjson.ActionUpdate},
					},
				},
			},
			want: map[string]tfjson.Change{
				"aws_s3_bucket.default": {
					Actions:      tfjson.Actions{tfjson.ActionCreate},
					AfterUnknown: map[string]any{"arn": true},
				},
				"module.function.aws_lambda_function.default": {
					Actions: tfjson.Actions{tfjson.ActionUpdate},
				},
			},
		},
		{
			name: "Deposed instances should not overwrite their live resource",
			changes: []*tfjson.ResourceChange{
				{
					Address: "aws_s3_bucket.default",
					Change: &tfjson.Change{
						Actions: tfjson.Actions{tfjson.ActionCreate},
					},
				},
				{
					Address:    "aws_s3_bucket.default",
					DeposedKey: "abc12345",
					Change: &tfjson.Change{
						Actions: tfjson.Actions{tfjson.ActionDelete},
					},
				},
			},
			want: map[string]tfjson.Change{
				"aws_s3_bucket.default": {
					Actions: tfjson.Actions{tfjson.ActionCreate},
				},
				"aws_s3_bucket.default (deposed abc12345)": {
					Actions: tfjson.Actions{tfjson.ActionDelete},
				},
			},
		},
		{
			name: "Changes without any detail should be skipped",
			changes: []*tfjson.ResourceChange{
				nil,
				{
					Address: "aws_s3_bucket.empty",
					Change:  nil,
				},
			},
			want: map[string]tfjson.Change{},
		},
		{
			name:    "A plan without changes should produce an empty collection",
			changes: nil,
			want:    map[string]tfjson.Change{},
		},
	}

	for _, tt := range tests {
		t.Logf("Running test -- %s", tt.name)
		plan := Plan{
			Debug:  true,
			Logger: logrus.New(),
		}
		got := plan.parseChanges(tt.changes)
		diff := deep.Equal(got, tt.want)
		if diff != nil {
			t.Fatalf("%s", testResults(tt.name, diff))
		}
	}
}

func TestParseKnownAfterApply(t *testing.T) {
	tests := []struct {
		name         string
		afterUnknown any
		want         KnownAfterApply
	}{
		{
			name: "Top level unknown attributes should be collected in sorted order",
			afterUnknown: map[string]any{
				"id":            true,
				"arn":           true,
				"force_destroy": false,
			},
			want: KnownAfterApply{
				{Path: "arn", Steps: []any{"arn"}},
				{Path: "id", Steps: []any{"id"}},
			},
		},
		{
			name: "Nested block attributes should be addressed by their full path",
			afterUnknown: map[string]any{
				"statement": []any{
					map[string]any{
						"actions":   []any{false},
						"resources": []any{true, false, true},
						"condition": []any{},
					},
				},
			},
			want: KnownAfterApply{
				{Path: "statement[0].resources[0]", Steps: []any{"statement", 0, "resources", 0}},
				{Path: "statement[0].resources[2]", Steps: []any{"statement", 0, "resources", 2}},
			},
		},
		{
			name: "A wholly unknown block should be collected at its root",
			afterUnknown: map[string]any{
				"server_side_encryption_configuration": true,
				"versioning": []any{
					map[string]any{
						"mfa_delete": true,
						"enabled":    false,
					},
				},
			},
			want: KnownAfterApply{
				{Path: "server_side_encryption_configuration", Steps: []any{"server_side_encryption_configuration"}},
				{Path: "versioning[0].mfa_delete", Steps: []any{"versioning", 0, "mfa_delete"}},
			},
		},
		{
			name: "Deeply nested unknowns should be collected at every depth",
			afterUnknown: map[string]any{
				"origin": map[string]any{
					"custom_header": []any{
						map[string]any{
							"value": true,
						},
					},
					"domain_name": true,
				},
			},
			want: KnownAfterApply{
				{Path: "origin.custom_header[0].value", Steps: []any{"origin", "custom_header", 0, "value"}},
				{Path: "origin.domain_name", Steps: []any{"origin", "domain_name"}},
			},
		},
		{
			name:         "A change with no unknown values should collect nothing",
			afterUnknown: map[string]any{"bucket": false, "tags": map[string]any{}},
			want:         nil,
		},
		{
			name:         "A change without an after_unknown object should collect nothing",
			afterUnknown: nil,
			want:         nil,
		},
		{
			name:         "An unknown at the root addresses no attribute and should be ignored",
			afterUnknown: true,
			want:         nil,
		},
	}

	for _, tt := range tests {
		t.Logf("Running test -- %s", tt.name)
		got := parseKnownAfterApply(tt.afterUnknown)
		diff := deep.Equal(got, tt.want)
		if diff != nil {
			t.Fatalf("%s", testResults(tt.name, diff))
		}
	}
}

func TestKnownAfterApplyLookups(t *testing.T) {
	known := KnownAfterApply{
		{Path: "arn", Steps: []any{"arn"}},
		{Path: "grant[0].id", Steps: []any{"grant", 0, "id"}},
		{Path: "versioning[0].mfa_delete", Steps: []any{"versioning", 0, "mfa_delete"}},
		{Path: "website.endpoint", Steps: []any{"website", "endpoint"}},
	}

	hasTests := []struct {
		name string
		path string
		want bool
	}{
		{name: "An exact top level match is known after apply", path: "arn", want: true},
		{name: "An exact nested match is known after apply", path: "grant[0].id", want: true},
		{name: "A parent of an unknown value is not itself an exact match", path: "grant", want: false},
		{name: "An unrelated attribute is not known after apply", path: "bucket", want: false},
	}
	for _, tt := range hasTests {
		t.Logf("Running test -- %s", tt.name)
		if got := known.Has(tt.path); got != tt.want {
			t.Fatalf("%s", testResults(tt.name, []string{"Has(" + tt.path + ")"}))
		}
	}

	prefixTests := []struct {
		name string
		path string
		want bool
	}{
		{name: "An exact match matches by prefix as well", path: "arn", want: true},
		{name: "A block containing an unknown element matches by prefix", path: "grant", want: true},
		{name: "A block containing an unknown attribute matches by prefix", path: "website", want: true},
		{name: "An indexed element containing an unknown attribute matches by prefix", path: "versioning[0]", want: true},
		{name: "A partial attribute name is not a prefix match", path: "ar", want: false},
		{name: "An unrelated attribute does not match by prefix", path: "bucket", want: false},
	}
	for _, tt := range prefixTests {
		t.Logf("Running test -- %s", tt.name)
		if got := known.HasPrefix(tt.path); got != tt.want {
			t.Fatalf("%s", testResults(tt.name, []string{"HasPrefix(" + tt.path + ")"}))
		}
	}

	t.Log("Running test -- Paths should list every unknown attribute address")
	wantPaths := []string{"arn", "grant[0].id", "versioning[0].mfa_delete", "website.endpoint"}
	diff := deep.Equal(known.Paths(), wantPaths)
	if diff != nil {
		t.Fatalf("%s", testResults("Paths should list every unknown attribute address", diff))
	}
}

func TestLinkResourcesWithChanges(t *testing.T) {
	tests := []struct {
		name      string
		changes   map[string]tfjson.Change
		resources []tfjson.StateResource
		want      []Resource
	}{
		{
			name: "Resources should link to their planned change and unknown attributes",
			changes: map[string]tfjson.Change{
				"aws_s3_bucket.default": {
					Actions:      tfjson.Actions{tfjson.ActionCreate},
					AfterUnknown: map[string]any{"arn": true, "bucket": false},
				},
			},
			resources: []tfjson.StateResource{
				{
					Address: "aws_s3_bucket.default",
					Type:    "aws_s3_bucket",
					Name:    "default",
				},
			},
			want: []Resource{
				{
					Module: Module{},
					Planned: tfjson.StateResource{
						Address: "aws_s3_bucket.default",
						Type:    "aws_s3_bucket",
						Name:    "default",
					},
					Change: tfjson.Change{
						Actions:      tfjson.Actions{tfjson.ActionCreate},
						AfterUnknown: map[string]any{"arn": true, "bucket": false},
					},
					KnownAfterApply: KnownAfterApply{
						{Path: "arn", Steps: []any{"arn"}},
					},
				},
			},
		},
		{
			name: "Deposed resources should link to their own change",
			changes: map[string]tfjson.Change{
				"aws_s3_bucket.default": {
					Actions: tfjson.Actions{tfjson.ActionCreate},
				},
				"aws_s3_bucket.default (deposed abc12345)": {
					Actions: tfjson.Actions{tfjson.ActionDelete},
				},
			},
			resources: []tfjson.StateResource{
				{
					Address:    "aws_s3_bucket.default",
					DeposedKey: "abc12345",
				},
			},
			want: []Resource{
				{
					Module: Module{},
					Planned: tfjson.StateResource{
						Address:    "aws_s3_bucket.default",
						DeposedKey: "abc12345",
					},
					Change: tfjson.Change{
						Actions: tfjson.Actions{tfjson.ActionDelete},
					},
				},
			},
		},
		{
			name:    "Resources without a declared change should still be returned",
			changes: map[string]tfjson.Change{},
			resources: []tfjson.StateResource{
				{
					Address: "aws_s3_bucket.unchanged",
				},
			},
			want: []Resource{
				{
					Module: Module{},
					Planned: tfjson.StateResource{
						Address: "aws_s3_bucket.unchanged",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Logf("Running test -- %s", tt.name)
		plan := Plan{
			Debug:   true,
			Logger:  logrus.New(),
			changes: tt.changes,
		}
		plan.linkResourcesWithModules([]Module{}, tt.resources)
		diff := deep.Equal(plan.Resources, tt.want)
		if diff != nil {
			t.Fatalf("%s", testResults(tt.name, diff))
		}
	}
}

func TestKnownAfterApplyFromPlanFile(t *testing.T) {
	tests := []struct {
		name     string
		planfile string
		address  string
		want     []string
	}{
		{
			name:     "A simple plan collects the unknown attributes of its resources",
			planfile: "./testdata/simple/plan.json",
			address:  "aws_s3_object.obj",
			want: []string{
				"bucket",
				"bucket_key_enabled",
				"content_type",
				"etag",
				"id",
				"kms_key_id",
				"server_side_encryption",
				"storage_class",
				"tags_all",
				"version_id",
			},
		},
		{
			name:     "A data source collects unknown attributes nested within its blocks",
			planfile: "./testdata/complex/plan.json",
			address:  "data.aws_iam_policy_document.policy",
			want: []string{
				"id",
				"json",
				"statement[0].resources[0]",
			},
		},
		{
			name:     "A module resource collects the unknown attributes of its change",
			planfile: "./testdata/complex/plan.json",
			address:  "module.iam_role.aws_iam_role_policy_attachment.custom[1]",
			want: []string{
				"id",
				"policy_arn",
			},
		},
	}

	for _, tt := range tests {
		t.Logf("Running test -- %s", tt.name)
		plan := Plan{
			PlanFile: tt.planfile,
		}
		plan.GetResources()
		var got []string
		for _, resource := range plan.Resources {
			if resource.Planned.Address == tt.address {
				got = resource.KnownAfterApply.Paths()
			}
		}
		diff := deep.Equal(got, tt.want)
		if diff != nil {
			t.Fatalf("%s", testResults(tt.name, diff))
		}
	}
}
