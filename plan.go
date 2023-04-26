package tfresources

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	tfjson "github.com/hashicorp/terraform-json"
)

// ParsePlan first validates that the user supplied plan file path
// exists. If we're unable to stat the file, then we fail early with
// a helpful error message.
//
// Once we've located the file, we ingest the contents and collect
// the root module contents (any raw terraform resources) declared
// in the plan. We then start collecting the child module contents,
// which include any resources that will be created as a result of
// declaring a `module` block.
//
// All of these resource contents are returned and used for linking
// later. If there is some problem with the linking process, this
// method can be called directly for a simple slice of raw tfjson
// StateResources for the deployment.
//
// [terraform-json]: https://github.com/hashicorp/terraform-json
func (p Plan) ParsePlan() ([]tfjson.StateResource, error) {
	var planContent tfjson.Plan
	var resources []tfjson.StateResource
	if !fileExists(p.PlanFile) {
		err := fmt.Errorf("unable to locate plan file at path %s", p.PlanFile)
		return resources, err
	}
	planFile, err := os.ReadFile(p.PlanFile)
	if err != nil {
		er := errors.Join(fmt.Errorf("unable to read specified plan file at %s", p.PlanFile), err)
		return resources, er
	}
	err = json.Unmarshal(planFile, &planContent)
	if err != nil {
		er := errors.Join(fmt.Errorf("error unmarshaling json contents of %s", p.PlanFile), err)
		return resources, er
	}

	rootModule := planContent.PlannedValues.RootModule
	children := planContent.PlannedValues.RootModule.ChildModules
	parseRootModuleResources(rootModule, &resources)
	parseChildModuleResources(children, &resources)
	return resources, nil
}

func parseRootModuleResources(root *tfjson.StateModule, resources *[]tfjson.StateResource) {
	for _, resource := range root.Resources {
		*resources = append(*resources, *resource)
	}
}

func parseChildModuleResources(children []*tfjson.StateModule, resources *[]tfjson.StateResource) {
	for _, child := range children {
		for _, res := range child.Resources {
			*resources = append(*resources, *res)
		}
		if child.ChildModules != nil {
			parseChildModuleResources(child.ChildModules, resources)
		}
	}
}
