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

## Known After Apply

Terraform can only resolve some attribute values while it is applying a plan, and renders those attributes as `(known after apply)`. Those values are **absent** from `Planned.AttributeValues`, which makes an attribute that Terraform will compute during the apply indistinguishable from one that was never configured at all — a rule like the one above will happily fail a bucket that is in fact going to be encrypted.

Every resource therefore also carries the change Terraform plans to make to it, along with a flattened list of the attributes that will only be known once that change has been applied.

```golang
for _, r := range plan.Resources {
    fmt.Println(r.Planned.Address)
    fmt.Println(r.Change.Actions)              // [create], [update], [delete replace], ...
    fmt.Println(r.KnownAfterApply.Paths())     // [arn id statement[0].resources[0] tags_all]
}
```

`KnownAfterApply` offers two lookups. `Has` matches an attribute address exactly, while `HasPrefix` also matches anything nested beneath it — which is the one to reach for when Terraform may report a block as unknown either in whole or in part:

```golang
encryption := "server_side_encryption_configuration"
if _, configured := r.Planned.AttributeValues[encryption]; configured {
    fmt.Printf("PASS -- %s -- encryption is configured\n", r.Planned.Address)
} else if r.KnownAfterApply.HasPrefix(encryption) {
    fmt.Printf("DEFER -- %s -- encryption will be resolved during the apply\n", r.Planned.Address)
} else {
    log.Fatalf("FAIL -- %s -- does not have encryption enabled!!", r.Planned.Address)
}
```

Attributes are addressed relative to the resource they belong to. Nested attributes are joined with a `.` and elements of a list or set are addressed by their index, producing paths such as `arn`, `tags_all`, or `statement[0].resources[1]`. Each entry also keeps that address broken into its individual `Steps` — a `string` for every object attribute or map key, an `int` for every collection index — for cases where a map key contains a `.` or a `[` and the rendered path would be ambiguous.

```golang
for _, attr := range r.KnownAfterApply {
    fmt.Println(attr.Path)   // statement[0].resources[1]
    fmt.Println(attr.Steps)  // [statement 0 resources 1]
}
```

The full change is available on `r.Change` as a [terraform-json] `Change`, so the prior state (`Before`), the planned state (`After`), the sensitive value masks, and the attributes forcing a replacement (`ReplacePaths`) are all reachable as well. Resources that the plan declares no change for are left with a zero valued `Change` and an empty `KnownAfterApply`.

[terraform-json]: https://github.com/hashicorp/terraform-json

## Authors

This package was written and maintained by [David Streng](https://www.linkedin.com/in/dave-streng) with the original concept created by [Patric Carman](https://www.linkedin.com/in/plcarman/).


## License
GNU General Public License v3.0 or later

See LICENSE to see the full text.

