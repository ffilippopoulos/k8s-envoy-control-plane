package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type K8sSource struct {
	Name       string `json:"name"`
	KubeConfig string `json:"kubeconfig"`
}

func LoadSourcesConfig(path string) ([]K8sSource, error) {
	confFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}

	fileContent, err := ioutil.ReadAll(confFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	sources := &[]K8sSource{}
	if err = json.Unmarshal(fileContent, sources); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}
	return *sources, nil
}
