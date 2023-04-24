package tfresources

import (
	"log"
)

// GetResources is the primary method used to orchestrate the
// parsing of both the user supplied Terraform plan file, as
// well as the optional terraform 'modules.json' file.
//
// It starts by parsing the 'modules.json' file and stores it's
// results into a slice of Module structs for later linking. This
// is quick since the contents of the file are relatively concise,
// and even for large deployments there are never more than a few
// hundred module definitions.
//
// Next it walks through the Terraform plan file and stores it's
// results into a slice of Resource structs for later linking. This
// process is more intensive since the amount of content in a Terraform
// plan does not necessarily scale linearly with the number of modules.
//
// Once we have both the resources and the modules, we finally link
// the resources toegether.
func (p *Plan) GetResources() {
	// modules, err := p.ParseModules()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// for _, module := range modules {
	// 	res := Resource{
	// 		Module: module,
	// 	}
	// 	p.Resources = append(p.Resources, res)
	// }
	resources, err := p.ParsePlan()
	if err != nil {
		log.Fatal(err)
	}
	for _, resource := range resources {
		res := Resource{
			Planned: resource,
		}
		p.Resources = append(p.Resources, res)
	}
}
