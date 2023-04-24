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
func (p Plan) ParseModules() ([]Module, error) {
	var path string
	var modules jsonModules
	cwd, _ := os.Getwd()
	if p.ModulesFilePath == "" {
		tfInitPath := filepath.Join(cwd, ".terraform")
		defaultModulesPath := filepath.Join(cwd, ".terraform", "modules", "modules.json")
		if !fileExists(tfInitPath) {
			errMessage := fmt.Sprintf("Unable to locate 'modules.json' file at the default path: %s\t%s%sPlease ensure that you ahve Terraofrm initialized in the current directory, or you have specified a custom 'modules.json' path via tfresources.Plan{}", newline(), tfInitPath, newline())
			err := errors.New(errMessage)
			return modules.Modules, err
		} else {
			path = defaultModulesPath
		}
	} else {
		path = p.ModulesFilePath
	}
	modulesFile, err := os.ReadFile(path)
	if err != nil {
		er := errors.Join(fmt.Errorf("error reading contents of %s", path), err)
		return modules.Modules, er
	}
	err = json.Unmarshal(modulesFile, &modules)
	if err != nil {
		er := errors.Join(fmt.Errorf("error unmarshalling json contents of %s", path), err)
		return modules.Modules, er
	}
	return modules.Modules, nil
}
