![License: GPL v3](https://img.shields.io/badge/License-GPL_v3-blue.svg)

# terraform-resources

A package to identify and parse through all resources in a known Terraform plan.

Instead of dicatating a particular configuration language to evaluate your Terraform resources, this package can be leveraged to write your own Golang application to parse the resources in ways that matter to YOU and your organization.

## Usage

```golang
// Currently requires that your plan file be exported in json format.
// This can be accomplished using the following terraform commands:
//      terraform plan -out=plan.tfplan
//      terraform show plan.tfplan -json > plan.json
//
plan := tfresources.Plan{
    PlanFile: "./deployment/tfplan.json"
}
plan.GetResources()
for _,r := range plan.Resources {
    if strings.Contains(r.Planned.ProviderName, "hashicorp/aws") && r.Planned.Mode != "data" && r.Planned.Type == "aws_s3_bucket" {
        if _, exists := r.Planned.AttributeValues["server_side_encryption_configuration"].([]interface{}); !exists {
            err := fmt.Errorf("S3 bucket -- %s -- does not have encryption enabled!!", r.Name)
            log.Fatal(err)
        }
    }
}
```

