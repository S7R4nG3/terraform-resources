package main

import (
	"fmt"
	"runtime"
	tfresources "tf-resources"
)

func main() {
	plan := tfresources.Plan{
		PlanFile: "./testdata/simple/plan.json",
	}
	plan.GetResources()
	for _, resource := range plan.Resources {
		fmt.Printf("%sModule Source: %s", newline(), resource.Module.Source)
		fmt.Printf("%sResource Address: %s\n", newline(), resource.Planned.Address)
	}
}

func newline() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}
