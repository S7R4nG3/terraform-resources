package tfresources

import (
	"fmt"
	"sort"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
)

// parseChanges collects the contents of a plan's `resource_changes` collection
// into a lookup table keyed by resource address so that each planned change
// can later be linked back to the resource it describes.
//
// Terraform records the changes it intends to make separately from the values
// it was able to plan, and it is the change - specifically its `after_unknown`
// object - that describes which of a resource's attributes will only become
// known once the plan has been applied.
//
// Changes without a body are skipped rather than stored empty, so that linking
// can tell a resource with no planned change apart from one whose change
// carried no detail.
func (p *Plan) parseChanges(changes []*tfjson.ResourceChange) map[string]tfjson.Change {
	p.debugLogger("Begin parsing planned resource changes...")
	parsed := make(map[string]tfjson.Change, len(changes))
	for _, change := range changes {
		if change == nil || change.Change == nil {
			continue
		}
		key := changeKey(change.Address, change.DeposedKey)
		p.debugLogger(fmt.Sprintf("Collecting planned change for resource -- %v -- with actions %v", key, change.Change.Actions))
		parsed[key] = *change.Change
	}
	p.debugLogger(fmt.Sprintf("Finished parsing planned resource changes, collected %d changes.", len(parsed)))
	return parsed
}

// changeKey builds the lookup key for a planned change.
//
// A deposed instance shares its address with the resource it was deposed from,
// which happens when a previous create_before_destroy replacement failed part
// way through. Folding the deposed key into the lookup keeps the two entries
// from overwriting one another.
func changeKey(address string, deposedKey string) string {
	if deposedKey == "" {
		return address
	}
	return fmt.Sprintf("%s (deposed %s)", address, deposedKey)
}

// parseKnownAfterApply flattens the `after_unknown` object of a planned change
// into the list of attributes that Terraform will only be able to resolve
// during an apply.
//
// Terraform reports these attributes as a deep object of booleans mirroring
// the shape of the resource itself, where a `true` marks the value at that
// position as unknown. A `true` may appear at any depth - against a top level
// attribute, an element of a list, or an attribute of a nested block - so the
// object is walked in full rather than only at its surface.
func parseKnownAfterApply(afterUnknown any) KnownAfterApply {
	var known KnownAfterApply
	walkUnknownAttributes(afterUnknown, []any{}, &known)
	return known
}

// walkUnknownAttributes recursively descends through an `after_unknown` object,
// tracking the traversal steps taken to reach the current node so that any
// unknown value it finds can be reported with its full address.
//
// Object attributes are visited in sorted order so that the resulting
// collection is stable between runs - Go randomizes map iteration, and an
// unstable ordering would make two parses of the same plan difficult to
// compare.
func walkUnknownAttributes(node any, steps []any, known *KnownAfterApply) {
	switch value := node.(type) {
	case bool:
		// A bare `true` at the root would refer to the resource as a whole
		// rather than to any attribute of it, so there is nothing to report.
		if value && len(steps) > 0 {
			*known = append(*known, UnknownAttribute{
				Path:  attributePath(steps),
				Steps: append([]any{}, steps...),
			})
		}
	case map[string]any:
		for _, key := range sortedKeys(value) {
			walkUnknownAttributes(value[key], append(steps, key), known)
		}
	case []any:
		for idx, element := range value {
			walkUnknownAttributes(element, append(steps, idx), known)
		}
	}
}

// attributePath renders a set of traversal steps as a single readable address,
// joining object attributes with a "." and wrapping collection indexes in
// brackets - "statement[0].resources[1]".
func attributePath(steps []any) string {
	var path strings.Builder
	for _, step := range steps {
		switch value := step.(type) {
		case int:
			fmt.Fprintf(&path, "[%d]", value)
		default:
			if path.Len() > 0 {
				path.WriteString(".")
			}
			fmt.Fprintf(&path, "%v", value)
		}
	}
	return path.String()
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Has reports whether the attribute at the provided path is known after apply.
//
// The path is matched exactly, so an unknown value nested beneath the provided
// path is not a match - use HasPrefix for that.
//
//	r.KnownAfterApply.Has("tags_all")
//	r.KnownAfterApply.Has("statement[0].resources[1]")
func (k KnownAfterApply) Has(path string) bool {
	for _, attribute := range k {
		if attribute.Path == path {
			return true
		}
	}
	return false
}

// HasPrefix reports whether the attribute at the provided path, or anything
// nested beneath it, is known after apply.
//
// This is the more forgiving check of the two, and is the one to reach for
// when an attribute may be reported as unknown either in whole or in part.
// Terraform will flag a nested block as unknown at its root when the entire
// block is computed, but will flag only the individual unknown values within
// it when the block itself is configured.
//
//	r.KnownAfterApply.HasPrefix("server_side_encryption_configuration")
func (k KnownAfterApply) HasPrefix(path string) bool {
	for _, attribute := range k {
		if attribute.Path == path ||
			strings.HasPrefix(attribute.Path, path+".") ||
			strings.HasPrefix(attribute.Path, path+"[") {
			return true
		}
	}
	return false
}

// Paths returns the addresses of every attribute that is known after apply,
// which is convenient for reporting and logging.
func (k KnownAfterApply) Paths() []string {
	paths := make([]string, 0, len(k))
	for _, attribute := range k {
		paths = append(paths, attribute.Path)
	}
	return paths
}
