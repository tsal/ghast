package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"reflect"

	"gopkg.in/yaml.v2"
)

/// Parse opens, loads and parses a YAML file via path
/// returns a potentially nested Map and a possible error.
func Parse(configPath string) (*map[string]interface{}, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	return parseYamlString(data)
}

func parseYamlString(data []byte) (*map[string]interface{}, error) {
	m := make(map[string]interface{})

	err := yaml.Unmarshal(data, &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return &m, err
}

/// ParsedConfigToContainerKeys parses the config generated by Parse into key, value pairs for injection into the Container
func ParsedConfigToContainerKeys(parsed *map[string]interface{}) (map[string]string, error) {
	result := make(map[string]string)

	for k, raw := range *parsed {
		flatten(result, k, reflect.ValueOf(raw))
	}

	return result, nil // TODO: Can this error?
}

func flatten(result map[string]string, prefix string, v reflect.Value) {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			result[prefix] = "true"
		} else {
			result[prefix] = "false"
		}
	case reflect.Int:
		result[prefix] = fmt.Sprintf("%d", v.Int())
	case reflect.Map:
		flattenMap(result, prefix, v)
	case reflect.Slice:
		flattenSlice(result, prefix, v)
	case reflect.String:
		result[prefix] = v.String()
	default:
		panic(fmt.Sprintf("Unknown: %s", v))
	}
}

func flattenMap(result map[string]string, prefix string, v reflect.Value) {
	for _, k := range v.MapKeys() {
		if k.Kind() == reflect.Interface {
			k = k.Elem()
		}

		if k.Kind() != reflect.String {
			panic(fmt.Sprintf("%s: map key is not string: %s", prefix, k))
		}

		flatten(result, fmt.Sprintf("%s.%s", prefix, k.String()), v.MapIndex(k))
	}
}

func flattenSlice(result map[string]string, prefix string, v reflect.Value) {
	prefix = prefix + "."

	result[prefix+"#"] = fmt.Sprintf("%d", v.Len())
	for i := 0; i < v.Len(); i++ {
		flatten(result, fmt.Sprintf("%s%d", prefix, i), v.Index(i))
	}
}
