![License: GPL v2](https://img.shields.io/badge/License-GPL_v2-blue.svg)

# terraform-resources

A package to identify and parse through all resources in a known Terraform plan.

Instead of dicatating a particular configuration language to evaluate your Terraform resources, this package can be leveraged to write your own Golang application to parse the resources in ways that matter to YOU and your organization.

## Usage

```golang
plan := tfresources.Plan{
    PlanFile: "./deployment/tfplan.json"
}
plan.GetResources()
for r := range plan.Resources {
    if r.ProviderName == "aws" && r.Mode != "data" && r.Type == "aws_s3_bucket" {
        if _, exists := r.AttributeValues["server_side_encryption_configuration"].([]interface{}); !exists {
            err := fmt.Errorf("S3 bucket -- %s -- does not have encryption enabled!!", r.Name)
            log.Fatal(err)
        }
    }
}
```

