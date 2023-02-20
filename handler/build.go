package handler

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/scryner/swagroller/static"
)

func DoBuild(inputFilepath, outputDir string) error {
	// try to open input file
	inputFinfo, err := os.Stat(inputFilepath)
	if err != nil {
		return fmt.Errorf("failed to stat input file '%s': %v", inputFilepath, err)
	}

	if inputFinfo.IsDir() {
		return fmt.Errorf("input path '%s' is directory", inputFilepath)
	}

	// try to open output directory
	outputFinfo, err := os.Stat(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			// try to make directory
			err = os.Mkdir(outputDir, 0755)
			if err != nil {
				return fmt.Errorf("failed to make directory '%s': %v\n", outputDir, err)
			}

		} else {
			return fmt.Errorf("failed to stat output path '%s': %v", outputDir, err)

		}
	} else {
		if !outputFinfo.IsDir() {
			return fmt.Errorf("output path '%s' is not directory", outputDir)
		}
	}

	// try to read yaml file
	title, specJSON, err := readYAMLtoJSON(inputFilepath)
	if err != nil {
		return fmt.Errorf("failed to convert to json: %v", err)
	}

	err = static.Build(title, specJSON, outputDir)
	if err != nil {
		return fmt.Errorf("failed to build: %v", err)
	}

	return nil
}

func readYAMLtoJSON(inputFilepath string) (string, []byte, error) {
	// try to read yaml file
	specYAML, err := ioutil.ReadFile(inputFilepath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to input file '%s': %v", inputFilepath, err)
	}

	title := "Swagroller"

	m := make(map[string]interface{})
	err = yaml.Unmarshal(specYAML, &m)
	if err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal yaml: %v", err)
	}

	info, ok := m["info"]
	if ok {
		info2 := info.(map[string]interface{})
		title, _ = info2["title"].(string)
	}

	// try to convert yaml to json
	specJSON, err := yaml.YAMLToJSON(specYAML)
	if err != nil {
		return "", nil, fmt.Errorf("failed to convert yaml to json: %v", err)
	}

	return title, specJSON, nil
}
