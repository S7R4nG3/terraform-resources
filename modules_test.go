package tfresources

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
)

type wantParseModules struct {
	modules []Module
	err     error
}

func TestParseModules(t *testing.T) {
	tests := []struct {
		name        string
		modulespath string
		want        wantParseModules
	}{
		{
			name:        "No module file should not error",
			modulespath: "",
			want: wantParseModules{
				modules: []Module{},
				err:     nil,
			},
		},
		{
			name:        "Module file that doesn't exist should error",
			modulespath: "./nonexistant",
			want: wantParseModules{
				modules: []Module{},
				err:     fmt.Errorf("Unable to locate 'modules.json' file at the specified path: %s\t%s%sPlease ensure that you ahve Terraform initialized in the current directory, or you have specified a custom 'modules.json' path via tfresources.Plan{}", newline(), "./nonexistant", newline()),
			},
		},
		{
			name:        "Modules should load properly",
			modulespath: "./testdata/complex/modules.json",
			want: wantParseModules{
				modules: []Module{
					{
						Key:     "",
						Source:  "",
						Version: "",
						Dir:     ".",
					},
					{
						Key:     "iam_role",
						Source:  "registry.terraform.io/terraform-aws-modules/iam/aws//modules/iam-assumable-role",
						Version: "5.17.0",
						Dir:     ".terraform/modules/iam_role/modules/iam-assumable-role",
					},
					{
						Key:     "my_bucket",
						Source:  "./modules/local",
						Version: "",
						Dir:     "modules/local",
					},
				},
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Logf("Running test -- %s", tt.name)
		plan := Plan{
			ModulesFilePath: tt.modulespath,
		}
		gotModules, gotErr := plan.ParseModules()
		modDiff := deep.Equal(gotModules, tt.want.modules)
		errDiff := deep.Equal(gotErr, tt.want.err)
		if tt.want.err != nil && errDiff != nil {
			t.Fatal(testResults(tt.name, errDiff))
		}
		if tt.want.err == nil && modDiff != nil {
			t.Fatal(testResults(tt.name, modDiff))
		}
	}
}
