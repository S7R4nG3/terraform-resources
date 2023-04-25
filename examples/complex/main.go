package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	tfresources "tf-resources"
)

func main() {
	plan := tfresources.Plan{
		PlanFile:        "./testdata/complex/plan.json",
		ModulesFilePath: "./testdata/complex/modules.json",
	}
	plan.GetResources()
	for _, resource := range plan.Resources {
		if strings.Contains(resource.Planned.ProviderName, "hashicorp/aws") && resource.Planned.Mode != "data" {
			if resource.Planned.Type == "aws_s3_bucket" {
				if _, exists := resource.Planned.AttributeValues["tags"]; !exists {
					log.Fatalf("S3 buckets must include tags!!%s\t%s", newline(), resource.Planned.AttributeValues)
				}
			}
		}
	}
	fmt.Println("DONE!")
}

func newline() string {
	if runtime.GOOS == "windows" {
		return "\r\n" // barf...
	}
	return "\n"
}
