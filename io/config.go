package io

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/sirupsen/logrus"
)

func recoverConfigAction() {
	if r := recover(); r != nil {
		logrus.Errorf("Recovered from some config action: %s\n", r)
		debug.PrintStack()
	}
}

// SaveConfig saves the provided config into the file with the given name
func SaveConfig(cfg interface{}, file string) (err error) {
	defer recoverConfigAction()

	file = validateFileName(file)

	os.MkdirAll(filepath.Dir(file), os.ModePerm)
	err = ioutil.WriteFile(file, hclEnconde(cfg), 0777)
	return
}

// LoadConfig loads a config from the given file, into the interface
func LoadConfig(file string, cfg interface{}) (err error) {
	defer recoverConfigAction()

	file = validateFileName(file)
	src, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	err = hclDecode(src, cfg)
	return
}

func validateFileName(fileName string) string {
	if !strings.HasSuffix(fileName, ".hcl") {
		fileName += ".hcl"
	}

	return fileName
}
