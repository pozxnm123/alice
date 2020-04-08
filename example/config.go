// Copyright © 2020 AMIS Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Threshold struct {
	DKG     uint32 `yaml:"dkg"`
	Signer  uint32 `yaml:"signer"`
	Reshare uint32 `yaml:"reshare"`
}

type Pubkey struct {
	X string `yaml:"x"`
	Y string `yaml:"y"`
}

type DKGResult struct {
	Share  string            `yaml:"share"`
	Pubkey Pubkey            `yaml:"pubkey"`
	BKs    map[string]string `yaml:"bks"`
}

type Config struct {
	Rank      uint32    `yaml:"rank"`
	FullAddr  string    `yaml:"fulladdr"`
	Threshold Threshold `yaml:"threshold"`
	DKGResult DKGResult `yaml:"dkgResult"`
}

func readYamlFile(filaPath string) (*Config, error) {
	c := &Config{}
	yamlFile, err := ioutil.ReadFile(filaPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func writeYamlFile(yamlData interface{}, filaPath string) error {
	data, err := yaml.Marshal(yamlData)
	if err != nil {
		return err
	}
	ioutil.WriteFile(filaPath, data, 0644)
	return nil
}

func getFilePath(id string) string {
	return fmt.Sprintf("config/%s.yaml", id)
}
