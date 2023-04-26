![License: GPL v3](https://img.shields.io/badge/License-GPL_v3-blue.svg)
![latest build](https://github.com/S7R4nG3/terraform-resources/actions/workflows/test.yml/badge.svg)
![latest release](https://img.shields.io/github/release-date/S7R4nG3/terraform-resources)
[![Go Reference](https://pkg.go.dev/badge/github.com/S7R4nG3/terraform-resources.svg)](https://pkg.go.dev/github.com/S7R4nG3/terraform-resources)
# terraform-resources

A package to identify and parse through all resources in a known Terraform plan.

Instead of dicatating a particular configuration language to evaluate your Terraform resources, this package can be leveraged to write your own Golang application to parse the resources in ways that matter to YOU and your organization.

## Usage

```golang
package main

import (
	"fmt"
    "log"
    "strings"

	tfresources "github.com/S7R4nG3/terraform-resources"
)
func main() {
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
    fmt.Println("DONE!")
}
```

## Authors

This package was written and maintained by [David Streng](https://www.linkedin.com/in/dave-streng) with the original concept created by [Patric Carman](https://www.linkedin.com/in/plcarman/).


## License
GNU General Public License v3.0 or later

See LICENSE to see the full text.

