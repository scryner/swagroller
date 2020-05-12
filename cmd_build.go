package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/scryner/swagroller/static"
	"github.com/ghodss/yaml"
)

func doBuild(usage func(), inputFilepath, outputDir string) {
	// check arguments
	if inputFilepath == "" || outputDir == "" {
		usage()
		os.Exit(1)
	}

	// try to open input file
	inputFinfo, err := os.Stat(inputFilepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to stat input file '%s': %v\n", inputFilepath, err)
		os.Exit(1)
	}

	if inputFinfo.IsDir() {
		fmt.Fprintf(os.Stderr, "input path '%s' is directory\n", inputFilepath)
		os.Exit(1)
	}

	// try to open output directory
	outputFinfo, err := os.Stat(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			// try to make directory
			err = os.Mkdir(outputDir, 0755)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to make directory '%s': %v\n", outputDir, err)
				os.Exit(1)
			}

		} else {
			fmt.Fprintf(os.Stderr, "failed to stat output path '%s': %v\n", outputDir, err)
			os.Exit(1)

		}
	} else {
		if !outputFinfo.IsDir() {
			fmt.Fprintf(os.Stderr, "output path '%s' is not directory\n", outputDir)
			os.Exit(1)
		}
	}

	// try to read yaml file
	title, specJSON, err := readYAMLtoJSON(inputFilepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to convert to json: %v\n", err)
		os.Exit(1)
	}

	err = static.Build(title, specJSON, outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build: %v\n", err)
		os.Exit(1)
	}
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
