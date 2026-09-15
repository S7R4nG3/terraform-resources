package tfresources

import (
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/sirupsen/logrus"
)

// A Plan initializes a primary configuration container
// that is used to specify the Terraform planfile as well
// as the Terraform modules file that are used to link the
// resources together. Each call to Plan will instantiate
// a new container to allow users to parse through multiple
// Terraform plan files.
//
// This container is then fed the results of the parsing
// to provide the end user with the list of their resource
// and module combinations for a particular Terraform plan.
//
// A Plan should be instantiated with at minimum the
// PlanFile so that it's GetResources() method can then
// be called to evaluate the resources in that plan.
//
//			plan := tfresources.Plan{
//				PlanFile: "./tfplan.json"
//			}
//			plan.GetResources()
//			for _, resource := range plan.Resources {
//				... Do things here ...
//	     }
//
// Multiple Plans can be declared at the same time to evaluate
// the resources from multiple Terraform executions simultaneously.
//
//	var allResources []tfresources.Resource
//	firstPlan := tfresources.Plan{
//		PlanFile: "./tfplan1.json"
//	}
//	firstPlan.GetResources()
//
//	secondPlan := tfresources.Plan{
//		PlanFile: "./tfplan2.json"
//	}
//	secondPlan.GetResources()
//
//	allResources = append(firstPlan.Resources, secondPlan.Resources)
//	for _, resource := range allResources {
//		... Do things here ...
//	}
//
// See the simple and complex examples for more detailed use cases.//
type Plan struct {

	// PlanFile is the primary entrypoint that is used to specify
	// the file path to your Terraform plan file that is to be parsed.
	// If PlanFile is not specified, the current program will exit
	// with an os.exit(1) error code.
	//
	// The specified PlanFile MUST be in JSON format. This can be
	// accomplished with Terraform itself using the following commands:
	//
	// 		terraform plan -out=plan.tfplan
	//		terraform show -json plan.tfpla > plan.json
	//
	PlanFile string

	// ModulesFilePath is an optional entrypoint parameter that can be
	// used to specify a custom modules.json file containing all the
	// modules that are being used for a particular Terraform plan.
	//
	// If not specified, the current program will default to Terraform's
	// default file path of `.terraform/modules/modules.json`.
	//
	// If unable to locate a modules file at the default path, it will
	// assume the deployment has no declared modules and will provide
	// the resources with a nil module value.
	ModulesFilePath string

	// A Resource container used to contain the results of parsing the
	// specified Terraform Plan.
	Resources []Resource

	// A Debug flag used to toggle stdout debug logging for the package.
	Debug bool

	// A Logger instantiation used for debug logging when toggled using
	// the above Debug flag.
	Logger *logrus.Logger

	// An internal lookup table of the planned changes declared in the
	// Terraform plan, keyed by resource address, that is used to link a
	// resource's planned change back to its planned values.
	changes map[string]tfjson.Change
}

// A Resource instantiates a new Terraform resource object
// that contains all the data Terraform was able to infer
// about the resource from the provided Terraform plan.
// These attributes are provided from Hashi's own
// [terraform-json] project.
//
// Additionally a Resource creates a default container that
// is used to hold the relevant information about the Terraform
// resource's parent Module (if available).
//
// This linkage allows users to scan a resource's configuration
// and link this configuration back to a specific module source
// (if available).
//
// [terraform-json]: https://github.com/hashicorp/terraform-json
type Resource struct {
	// A Module container used to link a particular Terraform resource
	// back to its parent module address (if available).
	Module Module

	// Planned stores the unmarshalled contents of a Terraform resource
	// provided from the [terraform-json] project.
	//
	// The values of Planned provide all the details that Terraform was
	// able to infer about the creation of a resource from the provided
	// Terraform plan file.
	//
	// It is important to note that not all values in Planned will have
	// a meaningful result, as Terraform is only able to infer the value of
	// some resource attributes during a Terraform `apply` execution. Those
	// attributes are described by KnownAfterApply below.
	//
	// [terraform-json]: https://github.com/hashicorp/terraform-json
	Planned tfjson.StateResource

	// Change stores the unmarshalled contents of the change Terraform plans
	// to make to this resource, provided from the [terraform-json] project.
	//
	// Where Planned describes what the resource will look like, Change
	// describes what Terraform intends to do to get there - the actions it
	// will take (create, update, delete, etc.), the prior state of the
	// resource, and which of its attributes are only resolvable during an
	// apply.
	//
	// If the plan contains no change entry for this resource, Change is left
	// as its zero value with a nil Actions slice.
	//
	// [terraform-json]: https://github.com/hashicorp/terraform-json
	Change tfjson.Change

	// KnownAfterApply enumerates every attribute of this resource whose value
	// Terraform is unable to resolve until the plan is applied - the values
	// Terraform renders as "(known after apply)" in its plan output.
	//
	// These attributes are deliberately absent from Planned.AttributeValues,
	// which makes an attribute that will be computed during the apply
	// indistinguishable from one that was never configured at all. Consulting
	// KnownAfterApply resolves that ambiguity.
	//
	//	if _, set := r.Planned.AttributeValues["kms_key_id"]; !set {
	//		if r.KnownAfterApply.Has("kms_key_id") {
	//			... Terraform will compute this during the apply ...
	//		} else {
	//			... this attribute genuinely isn't configured ...
	//		}
	//	}
	//
	// The collection is derived from the `after_unknown` object of the
	// resource's Change, and is empty for resources that have no change or no
	// unknown attributes.
	KnownAfterApply KnownAfterApply
}

// KnownAfterApply is the collection of a resource's attributes that Terraform
// has flagged as "(known after apply)".
//
// Attributes are ordered deterministically - nested object attributes are
// sorted by name and collection elements are held in their plan order - so
// the collection can be safely compared between runs of the same plan.
type KnownAfterApply []UnknownAttribute

// An UnknownAttribute identifies a single resource attribute whose value
// Terraform cannot determine until the plan is applied.
//
// Terraform reports these attributes as a deeply nested object that mirrors
// the shape of the resource itself, so an unknown value may live several
// levels beneath the resource's top level attributes. Each UnknownAttribute
// flattens one of those nested values into a single addressable entry.
type UnknownAttribute struct {
	// Path is the address of the attribute relative to the resource it
	// belongs to. Nested object attributes are joined with a "." and elements
	// of a list or set are addressed by their index, producing paths such as
	// "arn", "tags_all", or "statement[0].resources[1]".
	//
	// Terraform does not distinguish an object's attribute names from a map's
	// keys in its plan output, so a map key containing a "." or a "[" will
	// produce an ambiguous Path. Use Steps below if your resources include
	// such keys.
	Path string

	// Steps is the same address broken into its individual traversal steps,
	// where each step is either a string naming an object attribute or map
	// key, or an int indexing into a list or set.
	Steps []any
}

// A Module is a container to hold the parsed contents of a
// Terraform `modules.json` file.
//
// During the execution of a Terraform `init`, Terraform will
// read the contents of the Terraform files in the current
// directory and assemble this `modules.json` file so that it
// is able to track what module source URLs and versions are to
// be downloaded and used with which declarations (organized by
// key).
//
// Once assembled, the Terraform `init` command uses this file to
// download each module and store them locally under the `.terraform/modules`
// folder in the execution directory.
//
// If no modules are used for the deployment (i.e. a deployment using
// only raw Terraform resources) this file is not generated.
type Module struct {
	// Source identifies the module's source directory or download URL.
	// This value can be a local filesystem path, or a remote download
	// URL depending on where a user is sourcing their module contents
	// from.
	//
	// By default, all Terraform deployments include a module definition
	// with a source of "." (indicating the current directory).
	Source string `json:"Source"`

	// Version identifies the module's designated version. This value is only
	// provided if the end user is referencing a Terraform module from a configured
	// [Terraform Module Registry].
	//
	// If sourcing module content from a Git URL, the version will be included as a
	// `&ref=` parameter at the end of the source URL.
	//
	// [Terraform Module Registry]: https://developer.hashicorp.com/terraform/internals/module-registry-protocol
	Version string `json:"Version"`

	// Dir identifies the file path of the module's content relative to the
	// execution directory.
	Dir string `json:"Dir"`

	// Key identifies the full address of the Terraform resource that is to
	// be created. This value is used to uniquely identify separate resources
	// that may be created under a single module.
	Key string `json:"Key"`
}
