package io

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func parseRecoverErr(err *error) {
	if r := recover(); r != nil {
		switch x := r.(type) {
		case string:
			*err = fmt.Errorf(x)
		case error:
			*err = x
		default:
			*err = fmt.Errorf("unknown error from panic: %s", x)
		}
	}
}

// SaveConfig saves the provided config into the file with the given name
func SaveConfig(cfg interface{}, file string) (err error) {
	defer parseRecoverErr(&err)

	file = validateFileName(file)

	if err = os.MkdirAll(filepath.Dir(file), os.ModePerm); err != nil {
		return
	}

	err = ioutil.WriteFile(file, hclEncode(cfg), 0777)
	return
}

// LoadConfig loads a config from the given file, into the interface
func LoadConfig(file string, cfg interface{}) (err error) {
	defer parseRecoverErr(&err)

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
