package main

import (
	"fmt"
	"sort"
	"sync"

	tfresources "github.com/S7R4nG3/terraform-resources"
)

func main() {
	plan := tfresources.Plan{
		PlanFile:        "../../testdata/complex/plan.json",
		ModulesFilePath: "../../testdata/complex/modules.json",
	}
	plan.GetResources()
	wg := new(sync.WaitGroup)
	rules := []asyncWorker{
		resourcesMustBeTagged,
		resourcesCannotUsePublicRegistryModules,
		bucketsMustBeEncrypted,
	}

	ruleEngine(rules, plan.Resources, wg)
	wg.Wait()

	// Rules finish in a nondeterministic order.
	sort.Strings(results)
	for _, res := range results {
		fmt.Println(res)
	}
}
