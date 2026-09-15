package main

import (
	"runtime"
	"sync"

	tfresources "github.com/S7R4nG3/terraform-resources"
)

var (
	results   = []string{}
	resultsMu sync.Mutex
)

func newline() string {
	if runtime.GOOS == "windows" {
		return "\r\n" // barf...
	}
	return "\n"
}

func record(result string) {
	resultsMu.Lock()
	defer resultsMu.Unlock()
	results = append(results, result)
}

// Queues every resource then closes the channel. The buffer fits the whole
// deployment, so this never blocks.
func loader(resources []tfresources.Resource, i chan<- tfresources.Resource) {
	for _, res := range resources {
		i <- res
	}
	close(i)
}

// Each rule gets its own channel - sharing one would let whichever rule read
// first claim a resource, leaving the others to never evaluate it.
func ruleEngine(rules []asyncWorker, resources []tfresources.Resource, wg *sync.WaitGroup) {
	for _, rule := range rules {
		i := make(chan tfresources.Resource, len(resources))
		loader(resources, i)
		wg.Add(1)
		go rule(i, wg)
	}
}
