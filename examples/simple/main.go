package main

import (
	"fmt"
	"runtime"

	tfresources "github.com/S7R4nG3/terraform-resources"
)

func main() {
	plan := tfresources.Plan{
		PlanFile: "../../testdata/simple/plan.json",
	}
	plan.GetResources()
	for _, resource := range plan.Resources {
		fmt.Printf("%sModule Source: %s", newline(), resource.Module.Source)
		fmt.Printf("%sResource Address: %s", newline(), resource.Planned.Address)
		fmt.Printf("%sPlanned Actions: %v", newline(), resource.Change.Actions)
		fmt.Printf("%sKnown After Apply: %v\n", newline(), resource.KnownAfterApply.Paths())
	}
}

func newline() string {
	if runtime.GOOS == "windows" {
		return "\r\n" // barf...
	}
	return "\n"
}
