package tfresources

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
	tfjson "github.com/hashicorp/terraform-json"
)

type wantPlanParsing struct {
	resources []tfjson.StateResource
	err       error
}

func TestPlanParsing(t *testing.T) {
	tests := []struct {
		name     string
		planfile string
		plan     tfjson.Plan
		want     wantPlanParsing
	}{
		{
			name:     "Not specifying planfile results in error",
			planfile: "",
			want: wantPlanParsing{
				resources: []tfjson.StateResource{},
				err:       fmt.Errorf("unable to locate plan file at path %s", ""),
			},
		},
		{
			name:     "Specifying a simple plan provides simple resources",
			planfile: "./testdata/simple/plan.json",
			want: wantPlanParsing{
				resources: []tfjson.StateResource{
					{
						Address:      "aws_s3_bucket.default",
						Mode:         "managed",
						Type:         "aws_s3_bucket",
						Name:         "default",
						Index:        nil,
						ProviderName: "registry.terraform.io/hashicorp/aws",
						AttributeValues: map[string]interface{}{
							"bucket":        "my-test-bucket",
							"force_destroy": false,
							"tags":          nil,
							"timeouts":      nil,
						},
						SensitiveValues: []byte(fmt.Sprint(`{
                        "cors_rule": [],
                        "grant": [],
                        "lifecycle_rule": [],
                        "logging": [],
                        "object_lock_configuration": [],
                        "replication_configuration": [],
                        "server_side_encryption_configuration": [],
                        "tags_all": {},
                        "versioning": [],
                        "website": []
                    }`)),
						DependsOn:  nil,
						Tainted:    false,
						DeposedKey: "",
					},
					{
						Address:      "aws_s3_object.obj",
						Mode:         "managed",
						Type:         "aws_s3_object",
						Name:         "obj",
						Index:        nil,
						ProviderName: "registry.terraform.io/hashicorp/aws",
						AttributeValues: map[string]interface{}{
							"acl":                           "private",
							"cache_control":                 nil,
							"content":                       nil,
							"content_base64":                nil,
							"content_disposition":           nil,
							"content_encoding":              nil,
							"content_language":              nil,
							"force_destroy":                 false,
							"key":                           "a/magic/object/key",
							"metadata":                      nil,
							"object_lock_legal_hold_status": nil,
							"object_lock_mode":              nil,
							"object_lock_retain_until_date": nil,
							"source":                        "./outputs.tf",
							"source_hash":                   nil,
							"tags":                          nil,
							"website_redirect":              nil,
						},
						SensitiveValues: []byte(fmt.Sprint(`{
                        "tags_all": {}
                    }`)),
						DependsOn:  nil,
						Tainted:    false,
						DeposedKey: "",
					},
				},
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Logf("Running test -- %s", tt.name)
		plan := Plan{
			PlanFile: tt.planfile,
		}
		gotResources, gotErr := plan.ParsePlan()
		resDiff := deep.Equal(gotResources, tt.want.resources)
		errDiff := deep.Equal(gotErr, tt.want.err)
		if tt.want.err != nil && errDiff != nil {
			t.Fatalf("BAD")
		}
		if tt.want.err == nil && resDiff != nil {
			t.Fatalf(testResults(tt.name, resDiff))
		}
	}
}
