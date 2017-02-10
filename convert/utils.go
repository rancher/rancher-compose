package convert

func mapStringStringToMapStringInterface(m map[string]string) map[string]interface{} {
	if m == nil {
		return nil
	}
	newMap := map[string]interface{}{}
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

func mapStringInterfaceToStringString(m map[string]interface{}) map[string]string {
	if m == nil {
		return nil
	}
	newMap := map[string]string{}
	for k, v := range m {
		if s, ok := v.(string); ok {
			newMap[k] = s
		}
	}
	return newMap
}

func sliceField(slice []string) []string {
	if len(slice) == 0 {
		return nil
	}
	return slice
}

func Filter(vs []string, f func(string) bool) []string {
	r := make([]string, 0, len(vs))
	for _, v := range vs {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

func toMap(vs []string) map[string]struct{} {
	m := map[string]struct{}{}
	for _, v := range vs {
		if v != "" {
			m[v] = struct{}{}
		}
	}
	return m
}
