package main

import (
	"fmt"
	"sync"

	tfresources "github.com/S7R4nG3/terraform-resources"
)

var results = []string{}

func main() {
	plan := tfresources.Plan{
		PlanFile:        "../../testdata/complex/plan.json",
		ModulesFilePath: "../../testdata/complex/modules.json",
	}
	plan.GetResources()
	buffer := len(plan.Resources)
	wg := new(sync.WaitGroup)
	i := make(chan tfresources.Resource, buffer)
	rules := []asyncWorker{
		resourcesMustBeTagged,
		resourcesCannotUsePublicRegistryModules,
	}

	loader(plan.Resources, i, wg)
	ruleEngine(rules, i, wg, buffer)
	wg.Wait()

	for _, res := range results {
		fmt.Println(res)
	}
}
