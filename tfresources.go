// A package that will parse through a provided Terraform plan file
// and a Terraform modules.json file to link all resources to each
// parent module that can be then used in further applications for
// security scanning.
//
// These resources can be used to drive your own compliance rules
// according to what you and your organization may need with the
// flexibility of the entire Golang programming language at your
// fingertips.
package tfresources
