// A package that will parse through a provided Terraform plan file
// and a Terraform modules.json file to link all resources to each
// parent module that can be then used in further applications for
// security scanning.
//
// By default only the plan file path is required, but users
// additionally have the ability to specify their own modules.json
// file if Terraform has been initialized in a different directory
// or CI stage.
package tfresources
