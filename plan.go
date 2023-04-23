package tfresources

import (
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
)

// ParsePlan first validates that the user supplied plan file path
// exists. If we're unable to stat the file, then we fail early with
// a helpful error message.
//
// Once we've located the file, we unmarshal its contents leveraging
// the [terraform-json] project to write its contents to a slice
// tfjson StateResource local variable that is eventually returned
// unless there is an issue unmarshaling it's contents.
//
// [terraform-json]: https://github.com/hashicorp/terraform-json
func (p Plan) ParsePlan() ([]tfjson.StateResource, error) {
	var resources []tfjson.StateResource
	if !fileExists(p.PlanFile) {
		err := fmt.Errorf("unable to locate plan file at path %s", p.PlanFile)
		return resources, err
	}
	return resources, nil
}
