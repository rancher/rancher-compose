package rancher

import (
	"fmt"
)

func NestedMapsToMapInterface(data map[string]interface{}) map[string]interface{} {
	newMapInterface := map[string]interface{}{}

	for k, v := range data {
		switch v.(type) {
		case map[interface{}]interface{}:
			v = mapWalk(v.(map[interface{}]interface{}))
		case []interface{}:
			v = listWalk(v.([]interface{}))
		}
		newMapInterface[k] = v
	}

	return newMapInterface
}

func listWalk(val []interface{}) []interface{} {
	for i, v := range val {
		switch v.(type) {
		case map[interface{}]interface{}:
			val[i] = mapWalk(v.(map[interface{}]interface{}))
		case []interface{}:
			val[i] = listWalk(v.([]interface{}))
		}
	}
	return val
}

func mapWalk(val map[interface{}]interface{}) map[string]interface{} {
	newMap := map[string]interface{}{}

	for k, v := range val {
		strK := fmt.Sprintf("%v", k)
		switch v.(type) {
		case map[interface{}]interface{}:
			v = mapWalk(v.(map[interface{}]interface{}))
		case []interface{}:
			v = listWalk(v.([]interface{}))
		}
		newMap[strK] = v
	}

	return newMap
}

func Contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

func ToMapInterface(data map[string]string) map[string]interface{} {
	ret := map[string]interface{}{}

	for k, v := range data {
		ret[k] = v
	}

	return ret
}

func MapUnion(left, right map[string]string) map[string]string {
	ret := map[string]string{}

	for k, v := range left {
		ret[k] = v
	}

	for k, v := range right {
		ret[k] = v
	}

	return ret
}
