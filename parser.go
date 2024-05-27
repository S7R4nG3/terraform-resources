package tfresources

import (
	"github.com/sirupsen/logrus"
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
// the resources toegether and return them as a slice of interlinked
// tjson StateResources with their associated module Address, Key,
// Source, and Dir attributes.
func (p *Plan) GetResources() {
	p.Logger = logrus.New()
	p.debugLogger("Starting resource aggregation...")

	modules, err := p.parseModules()
	if err != nil {
		p.Logger.Error(err)
	}
	resources, err := p.parsePlan()
	if err != nil {
		p.Logger.Error(err)
	}
	p.linkResourcesWithModules(modules, resources)
	p.debugLogger("Resource aggregation complete.")
}
