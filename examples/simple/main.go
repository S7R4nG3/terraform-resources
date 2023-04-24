package main

import (
	"fmt"
	tfresources "tf-resources"
)

func main() {
	plan := tfresources.Plan{
		PlanFile:        "../../testdata/simple/plan.json",
		ModulesFilePath: "../../testdata/simple/modules.json",
	}
	plan.GetResources()
	for _, resource := range plan.Resources {
		fmt.Println(resource.Module)
		fmt.Println(resource.Planned.Address)
	}
}
