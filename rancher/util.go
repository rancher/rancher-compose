package rancher

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
