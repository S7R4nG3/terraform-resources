package main

import (
	"runtime"
	"sync"

	tfresources "github.com/S7R4nG3/terraform-resources"
)

func newline() string {
	if runtime.GOOS == "windows" {
		return "\r\n" // barf...
	}
	return "\n"
}

func loader(resources []tfresources.Resource, i chan<- tfresources.Resource, wg *sync.WaitGroup) {
	for _, res := range resources {
		wg.Add(1)
		i <- res
	}
}

func ruleEngine(rules []asyncWorker, i <-chan tfresources.Resource, wg *sync.WaitGroup, buffer int) {
	for idx := 0; idx < buffer; idx++ {
		for _, rule := range rules {
			go rule(i, wg)
		}
	}
}
