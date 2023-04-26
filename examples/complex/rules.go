package main

import (
	"fmt"
	"strings"
	"sync"

	tfresources "github.com/S7R4nG3/terraform-resources"
)

type asyncWorker func(<-chan tfresources.Resource, *sync.WaitGroup)

func resourcesMustBeTagged(i <-chan tfresources.Resource, wg *sync.WaitGroup) {
	defer wg.Done()
	ruleName := "AWS Resources must be tagged."
	r := <-i
	if _, exists := r.Planned.AttributeValues["tags"]; !exists {
		results = append(results, fmt.Sprintf("FAIL -- %s -- %s", r.Planned.Address, ruleName))
	} else {
		results = append(results, fmt.Sprintf("PASS -- %s -- %s", r.Planned.Address, ruleName))
	}
}

func resourcesCannotUsePublicRegistryModules(i <-chan tfresources.Resource, wg *sync.WaitGroup) {
	defer wg.Done()
	ruleName := "AWS Resources cannot use Public Terraform Module Registry"
	r := <-i
	if strings.Contains(r.Module.Source, "registry.terraform.io") {
		results = append(results, fmt.Sprintf("FAIL -- %s -- %s", r.Planned.Address, ruleName))
	} else {
		results = append(results, fmt.Sprintf("PASS -- %s -- %s", r.Planned.Address, ruleName))
	}
}
