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
func (p Plan) ParseModules() ([]Module, error) {
	var path string
	var jsonModules jsonModules
	cwd, _ := os.Getwd()
	// Check the supplied ModulesFilePath
	if p.ModulesFilePath != "" && fileExists(p.ModulesFilePath) {
		path = p.ModulesFilePath
		// If not supplied but it doesn't exist - return a more meaningful error
	} else if p.ModulesFilePath != "" && !fileExists(p.ModulesFilePath) {
		// errMessage := fmt.Sprintf("Unable to locate 'modules.json' file at the specified path: %s\t%s%sPlease ensure that you ahve Terraform initialized in the current directory, or you have specified a custom 'modules.json' path via tfresources.Plan{}", newline(), p.ModulesFilePath, newline())
		// err := errors.New(errMessage)
		return []Module{}, fmt.Errorf("Unable to locate 'modules.json' file at the specified path: %s\t%s%sPlease ensure that you ahve Terraform initialized in the current directory, or you have specified a custom 'modules.json' path via tfresources.Plan{}", newline(), p.ModulesFilePath, newline())
		// Else check for the default path and if it doesn't exist return early - no modules declared
	} else if p.ModulesFilePath == "" {
		tfDefaultModulesPath := filepath.Join(cwd, ".terraform", "modules", "modules.json")
		if !fileExists(tfDefaultModulesPath) {
			return []Module{}, nil
		}
		path = tfDefaultModulesPath
	}
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
	return jsonModules.Modules, nil
}
