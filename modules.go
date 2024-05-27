package tfresources

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Helper struct to make it easier to unmarshal the 'modules.json' file.
type jsonModules struct {
	Modules []Module `json:"Modules"`
}

// ParseModules begins by locating the file path of the 'modules.json' file.
// By default, Terraform will initialize this file in the execution directory
// under `/.terraform/modules/modules.json`. However this parameter can be
// optionally specified by the user to designate multiple deployments.
//
// We start by checking to see if the user specified a custom module file path
// and validating that the file exists. If not found, we check the default
// file location and validate that it exists.
//
// If either of these checks fail, we fail early with helpful error prompts.
//
// Once we've located the file, we unmarshal its contents into a local slice
// Module variable that is eventually returned unless there is an error during
// unmarshaling.
//
// If there is an error with resource linking, this method can be called directly
// to return a simple list of all declared modules and their requisite information.
func (p Plan) parseModules() ([]Module, error) {
	p.debugLogger("Begin parsing modules...")
	var path string
	var jsonModules jsonModules
	cwd, _ := os.Getwd()
	// Check the supplied ModulesFilePath
	if p.ModulesFilePath != "" && fileExists(p.ModulesFilePath) {
		p.debugLogger(fmt.Sprintf("Utilizing modules.json file at provided path of %v", p.ModulesFilePath))
		path = p.ModulesFilePath
		// If supplied but it doesn't exist - return a more meaningful error
	} else if p.ModulesFilePath != "" && !fileExists(p.ModulesFilePath) {
		p.debugLogger(fmt.Sprintf("Provided ModulesFilePath of %s does not exist", p.ModulesFilePath))
		return []Module{}, fmt.Errorf("unable to locate 'modules.json' file at the specified path: %s\t%s%sPlease ensure that you have Terraform initialized in the current directory, or you have specified a custom 'modules.json' path via tfresources.Plan{}", newline(), p.ModulesFilePath, newline())
		// Else check for the default path and if it doesn't exist return early - no modules declared
	} else if p.ModulesFilePath == "" {
		tfDefaultModulesPath := filepath.Join(cwd, ".terraform", "modules", "modules.json")
		p.debugLogger(fmt.Sprintf("No ModulesFilePath provided, checking default path of %v", tfDefaultModulesPath))
		if !fileExists(tfDefaultModulesPath) {
			p.debugLogger(fmt.Sprintf("No modules.json file located at default path of %v", tfDefaultModulesPath))
			return []Module{}, nil
		}
		p.debugLogger(fmt.Sprintf("Using modules.json file at default path of %v", tfDefaultModulesPath))
		path = tfDefaultModulesPath
	}
	p.debugLogger("Reading modules.json file contents...")
	modulesFile, err := os.ReadFile(path)
	if err != nil {
		er := errors.Join(fmt.Errorf("error reading contents of %s", path), err)
		return jsonModules.Modules, er
	}
	err = json.Unmarshal(modulesFile, &jsonModules)
	if err != nil {
		er := errors.Join(fmt.Errorf("error unmarshalling json contents of %s", path), err)
		return jsonModules.Modules, er
	}
	p.debugLogger("Finished parsing modules.")
	return jsonModules.Modules, nil
}
