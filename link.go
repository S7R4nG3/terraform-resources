package tfresources

import (
	"fmt"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
)

func (p *Plan) linkResourcesWithModules(planModules []Module, planResources []tfjson.StateResource) {
	for _, resource := range planResources {
		module := parseModuleFromResourceAddress(resource, planModules)
		parent := parseParentFromChildModule(module, planModules)
		if parent != (Module{}) {
			// This module has a parent - link it
			module = parent
		}
		this := Resource{
			Module:  module,
			Planned: resource,
		}
		p.Resources = append(p.Resources, this)
	}
}

func parseModuleFromResourceAddress(resource tfjson.StateResource, modules []Module) Module {
	if !strings.HasPrefix(resource.Address, "module.") {
		return Module{}
	} else {
		regexes := []string{
			`^module\.`,  // Strip leading "module." ref module.<friendly name>.<resource>...
			`\.module\.`, // Remove any intermediate ".module." refs <friendly name>.module.<resource>
			`\[\d+\]`,    // Remove any list indexed refs <resource>[0].<friendly name>
			`\[\".+\"\]`, // Remove any refs to map keys <resource>.<friendly name>[\"name\"]
			fmt.Sprintf(`\.%s\.%s`, resource.Type, resource.Name), // Finally trim off the resource type and friendly name refs <friendly name>.aws_s3_bucket.default
		}
		key := removeString(regexes, resource.Address)
		for idx := range modules {
			if modules[idx].Key == key {
				return modules[idx]
			}
		}
	}
	return Module{}
}

func parseParentFromChildModule(child Module, modules []Module) Module {
	if strings.Contains(child.Key, ".") {
		parent := removeString([]string{`\..+$`}, child.Key)
		for idx := range modules {
			if modules[idx].Key == parent {
				return modules[idx]
			}
		}
	}
	return Module{}
}
