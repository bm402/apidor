package copy

// Map is a http function that makes a copy of the map[string]interface{} type
func Map(mp map[string]interface{}) map[string]interface{} {
	mpCopy := make(map[string]interface{})
	for key, value := range mp {
		switch value.(type) {
		case []interface{}:
			mpCopy[key] = Array(value.([]interface{}))
		case map[string]interface{}:
			mpCopy[key] = Map(value.(map[string]interface{}))
		default:
			mpCopy[key] = value
		}
	}
	return mpCopy
}

// MapOfStrings is a http function that makes a copy of the map[string]string type
func MapOfStrings(mp map[string]string) map[string]string {
	mpCopy := make(map[string]string)
	for key, value := range mp {
		mpCopy[key] = value
	}
	return mpCopy
}

// Array is a http function that makes a copy of the []interface{} type
func Array(arr []interface{}) []interface{} {
	arrCopy := []interface{}{}
	for _, value := range arr {
		switch value.(type) {
		case []interface{}:
			arrCopy = append(arrCopy, Array(value.([]interface{})))
		case map[string]interface{}:
			arrCopy = append(arrCopy, Map(value.(map[string]interface{})))
		default:
			arrCopy = append(arrCopy, value)
		}
	}
	return arrCopy
}
