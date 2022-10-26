package mservice_util

import (
	"os"
	"fmt"
	"path/filepath"
	"hbservice/src/util"
)

func GetConfigName() string {
	return fmt.Sprintf("%s.cfg", util.GetExecName())
}

func GetConfigPath() string {
	return fmt.Sprintf("%s.cfg", util.GetExecPath())
}

func GetServiceConfigPath() string {
	full_path, _ := os.Executable()
	path, _ := filepath.Split(full_path)
	return fmt.Sprintf("%smservice.cfg", path)
}