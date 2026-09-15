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
			module = parent // This module has a parent - link it
		}
		this := Resource{
			Module:  module,
			Planned: resource,
		}
		p.linkResourceWithChange(&this)
		p.Resources = append(p.Resources, this)
	}
}

// linkResourceWithChange attaches the change Terraform plans to make to a
// resource, matched to it by the resource's address.
//
// The attributes that will only be known after the apply are flattened out of
// that change as they are linked, so that callers are handed a directly usable
// list rather than the deeply nested object Terraform reports.
//
// Not every planned resource has a corresponding change - a resource that is
// wholly unchanged may be absent from the plan's `resource_changes` collection
// - so a resource with no match is simply left with its zero valued change.
func (p *Plan) linkResourceWithChange(resource *Resource) {
	key := changeKey(resource.Planned.Address, resource.Planned.DeposedKey)
	change, exists := p.changes[key]
	if !exists {
		p.debugLogger(fmt.Sprintf("No planned change declared for resource -- %v --", key))
		return
	}
	resource.Change = change
	resource.KnownAfterApply = parseKnownAfterApply(change.AfterUnknown)
	p.debugLogger(fmt.Sprintf("Linked resource -- %v -- with %d attributes known after apply", key, len(resource.KnownAfterApply)))
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
