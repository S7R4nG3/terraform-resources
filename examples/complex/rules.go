package main

import (
	"fmt"
	"strings"
	"sync"

	tfresources "github.com/S7R4nG3/terraform-resources"
)

type asyncWorker func(<-chan tfresources.Resource, *sync.WaitGroup)

func resourcesMustBeTagged(i <-chan tfresources.Resource, wg *sync.WaitGroup) {
	defer wg.Done()
	ruleName := "AWS Resources must be tagged."
	for r := range i {
		if _, exists := r.Planned.AttributeValues["tags"]; !exists {
			record(fmt.Sprintf("FAIL -- %s -- %s", r.Planned.Address, ruleName))
		} else {
			record(fmt.Sprintf("PASS -- %s -- %s", r.Planned.Address, ruleName))
		}
	}
}

func resourcesCannotUsePublicRegistryModules(i <-chan tfresources.Resource, wg *sync.WaitGroup) {
	defer wg.Done()
	ruleName := "AWS Resources cannot use Public Terraform Module Registry"
	for r := range i {
		if strings.Contains(r.Module.Source, "registry.terraform.io") {
			record(fmt.Sprintf("FAIL -- %s -- %s", r.Planned.Address, ruleName))
		} else {
			record(fmt.Sprintf("PASS -- %s -- %s", r.Planned.Address, ruleName))
		}
	}
}

// Encryption is often managed by a separate resource, which leaves the bucket's
// own attribute flagged as known after apply rather than configured. Checking
// KnownAfterApply avoids failing a bucket that will in fact be encrypted.
func bucketsMustBeEncrypted(i <-chan tfresources.Resource, wg *sync.WaitGroup) {
	defer wg.Done()
	ruleName := "AWS S3 buckets must be encrypted."
	encryption := "server_side_encryption_configuration"
	for r := range i {
		if r.Planned.Type != "aws_s3_bucket" || r.Planned.Mode == "data" {
			continue
		}
		if _, configured := r.Planned.AttributeValues[encryption]; configured {
			record(fmt.Sprintf("PASS -- %s -- %s", r.Planned.Address, ruleName))
		} else if r.KnownAfterApply.HasPrefix(encryption) {
			record(fmt.Sprintf("DEFERRED -- %s -- %s -- encryption is known after apply", r.Planned.Address, ruleName))
		} else {
			record(fmt.Sprintf("FAIL -- %s -- %s", r.Planned.Address, ruleName))
		}
	}
}
