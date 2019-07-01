package main

import (
	"errors"
	"io/ioutil"
	"strings"

	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	"gopkg.in/yaml.v2"
)

func promInitializeMappings() map[string]map[string]int {
	mappings := make(map[string]map[string]int)
	yamlFile, _ := ioutil.ReadFile(promValueMap)
	yaml.Unmarshal(yamlFile, &mappings)
	return mappings
}

func promTranslateString(kvpair *na_pb.KeyValue, m map[string]map[string]int) (float64, error) {
	ret := float64(0)
	for key_1, kv := range m {
		if !(strings.Contains(kvpair.Key, key_1)) {
			continue
		}
		for key_2, value := range kv {
			if !(strings.Contains(kvpair.GetStrValue(), key_2)) {
				continue
			}
			ret = float64(value)
			return ret, nil
		}
	}
	return ret, errors.New("Did not find the string")
}
