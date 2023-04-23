package tfresources

import tfjson "github.com/hashicorp/terraform-json"

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
// 		plan := tfresources.Plan{
//			PlanFile: "./tfplan.json"
// 		}
//		plan.GetResources()
//		for resource := range plan.Resources {
//			... Do things here ...
//      }
//
// Multiple Plans can be declared at the same time to evaluate
// the resources from multiple Terraform executions simultaneously.
//
//		var allResources []tfresources.Resource
//		firstPlan := tfresources.Plan{
//			PlanFile: "./tfplan1.json"
//		}
//		secondPlan := tfresources.Plan{
//			PlanFile: "./tfplan2.json"
//		}
// 		firstPlan.GetResources()
//		secondPlan.GetResources()
//		allResources = firstPlan.Resources + secondPlan.Resources
//		for resource := range allResources {
//			... Do things here ...
//		}
//
type Plan struct {

	// PlanFile is the primary entrypoint that is used to specify
	// the file path to your Terraform plan file that is to be parsed.
	// If PlanFile is not specified, the current program will exit
	// with an os.exit(1) error code.
	PlanFile string

	// ModulesFilePath is an optional entrypoint parameter that can be
	// used to specify a custom modules.json file containing all the
	// modules that are being used for a particular Terraform plan.
	// If not specified, the current program will default to Terraform's
	// default file path of `.terraform/modules/modules.json`.
	// If the default path is unable to be found, the current program
	// will exit with an os.exit(1) error code.
	ModulesFilePath string

	// A Resource container used to contain the results of parsing the
	// specified Terraform Plan.
	Resources []Resource
}

// A Resource instantiates a new Terraform resource object
// that contains all the data Terraform was able to infer
// about the resource from the provided Terraform plan.
// These attributes are provided from Hashi's own
// [terraform-json] project.
// Additionally a Resource creates a default container that
// is used to hold the relevant information about the Terraform
// resource's parent Module (if available).
// This linkage allows users to scan a resource's configuration
// and link this configuration back to a specific module source
// (if available).
//
// [terraform-json]: https://github.com/hashicorp/terraform-json
type Resource struct {
	// A Module container used to link a particular Terraform resource
	// back to its parent module address (if available).
	Module Module

	// Value stores the unmarshalled contents of a Terraform resource
	// provided from the terraform-json project.
	// The attributes of Value provide all the details that Terraform was
	// able to infer about the creation of a resource from the provided
	// Terraform plan file.
	// It is important to note that not all attributes in a Value will have
	// a meaningful result, as Terraform is only able to infer the value of
	// some resource attributes during a Terraform `apply` execution.
	//
	// [terraform-json]: https://github.com/hashicorp/terraform-json
	Value tfjson.StateResource
}

// A Module is a container to hold the parsed contents of a
// Terraform `modules.json` file.
// During the execution of a Terraform `init`, Terraform will
// read the contents of the Terraform files in the current
// directory and assemble this `modules.json` file so that it
// is able to track what module source URLs and versions are to
// be downloaded and used with the deploy.
//
// Once assembled, the Terraform `init` command uses this file to
// download each module and store them locally under the `.terraform/modules`
// folder in the execution directory.
type Module struct {
	// Source identifies the module's source directory or download URL.
	// This value can be a local filesystem path, or a remote download
	// URL depending on where a user is sourcing their module contents
	// from.
	// By default, all Terraform deployments include a module definition
	// with a source of "." (indicating the current directory).
	Source string `json:"Source"`

	// Version identifies the module's designated version. This value is only
	// provided if the end user is referencing a Terraform module from a configured
	// [Terraform Module Registry].
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
	// that may be using the same module code.
	Key string `json:"Key"`
}
