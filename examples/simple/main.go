package main

import (
	"fmt"
	tfresources "tf-resources"
)

func main() {
	plan := tfresources.Plan{
		PlanFile:        "../../testdata/simple/plan.json",
		ModulesFilePath: "../../testdata/simple/.terraform/modules/modules.json",
	}
	plan.GetResources()
	for resource := range plan.Resources {
		fmt.Println(resource)
	}
}
