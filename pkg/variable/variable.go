package variable

import (
	"fmt"
)

// FindVarsInString is a variable function for finding any $ delimited variables in a string
func FindVarsInString(str string) []string {
	vars := []string{}
	isVar := false
	varStartIndex := 0

	for curIndex, char := range str {
		if char == '$' {
			if isVar {
				vars = append(vars, str[varStartIndex:curIndex+1])
				isVar = false
			} else {
				varStartIndex = curIndex
				isVar = true
			}
		} else if isVar && (char == '/' || char == '.') {
			vars = append(vars, str[varStartIndex:curIndex])
			isVar = false
		}
	}

	if isVar {
		vars = append(vars, str[varStartIndex:len(str)])
	}

	return vars
}

// FindVarsInMap is a variable function for finding any $ delimited variables in a map
func FindVarsInMap(mp map[string]interface{}) []string {
	vars := []string{}

	for _, value := range mp {
		switch value.(type) {
		case string:
			vars = append(vars, FindVarsInString(value.(string))...)
		case []interface{}:
			vars = append(vars, FindVarsInArray(value.([]interface{}))...)
		case map[string]interface{}:
			vars = append(vars, FindVarsInMap(value.(map[string]interface{}))...)
		}
	}

	return vars
}

// FindVarsInMapOfStrings is a variable function for finding any $ delimited variables in a map of strings
func FindVarsInMapOfStrings(mp map[string]string) []string {
	vars := []string{}
	for _, value := range mp {
		vars = append(vars, FindVarsInString(value)...)
	}
	return vars
}

// FindVarsInArray is a variable function for finding any $ delimited variables in an array
func FindVarsInArray(arr []interface{}) []string {
	vars := []string{}

	for _, value := range arr {
		switch value.(type) {
		case string:
			vars = append(vars, FindVarsInString(value.(string))...)
		case []interface{}:
			vars = append(vars, FindVarsInArray(value.([]interface{}))...)
		case map[string]interface{}:
			vars = append(vars, FindVarsInMap(value.(map[string]interface{}))...)
		}
	}

	return vars
}

// SubstituteVarsInString is a variable function for substituting any $ delimited variables in a string
func SubstituteVarsInString(str string, vars map[string]interface{}) string {
	strWithVars := ""
	isVar := false
	varStartIndex := 0
	fixedStartIndex := 0

	toString := func(value interface{}) string {
		return fmt.Sprintf("%v", value)
	}

	for curIndex, char := range str {
		if char == '$' {
			if isVar {
				strWithVars += toString(vars[str[varStartIndex:curIndex+1]])
				fixedStartIndex = curIndex + 1
				isVar = false
			} else {
				strWithVars += str[fixedStartIndex:curIndex]
				varStartIndex = curIndex
				isVar = true
			}
		} else if isVar && (char == '/' || char == '.') {
			strWithVars += toString(vars[str[varStartIndex:curIndex]])
			fixedStartIndex = curIndex
			isVar = false
		}
	}

	if isVar {
		strWithVars += toString(vars[str[varStartIndex:len(str)]])
	} else {
		strWithVars += str[fixedStartIndex:len(str)]
	}

	return strWithVars
}

// SubstituteVarsInMap is a variable function for substituting any $ delimited variables in a map
func SubstituteVarsInMap(mp map[string]interface{}, vars map[string]interface{}) map[string]interface{} {
	mapWithVars := make(map[string]interface{})

	for key, value := range mp {
		switch value.(type) {
		case string:
			if varToSubstitute, ok := vars[value.(string)]; ok {
				mapWithVars[key] = varToSubstitute
			} else {
				mapWithVars[key] = value
			}
		case []interface{}:
			mapWithVars[key] = SubstituteVarsInArray(value.([]interface{}), vars)
		case map[string]interface{}:
			mapWithVars[key] = SubstituteVarsInMap(value.(map[string]interface{}), vars)
		default:
			mapWithVars[key] = value
		}
	}

	return mapWithVars
}

// SubstituteVarsInMapOfStrings is a variable function for substituting any $ delimited variables in a map
func SubstituteVarsInMapOfStrings(mp map[string]string, vars map[string]interface{}) map[string]string {
	mapWithVars := make(map[string]string)
	for key, value := range mp {
		if varToSubstitute, ok := vars[value]; ok {
			mapWithVars[key] = fmt.Sprintf("%v", varToSubstitute)
		} else {
			mapWithVars[key] = value
		}
	}
	return mapWithVars
}

// SubstituteVarsInArray is a variable function for substituting any $ delimited variables in an array
func SubstituteVarsInArray(arr []interface{}, vars map[string]interface{}) []interface{} {
	arrWithVars := []interface{}{}

	for _, value := range arr {
		switch value.(type) {
		case string:
			if varToSubstitute, ok := vars[value.(string)]; ok {
				arrWithVars = append(arrWithVars, varToSubstitute)
			} else {
				arrWithVars = append(arrWithVars, value)
			}
		case []interface{}:
			arrWithVars = append(arrWithVars, SubstituteVarsInArray(value.([]interface{}), vars)...)
		case map[string]interface{}:
			arrWithVars = append(arrWithVars, SubstituteVarsInMap(value.(map[string]interface{}), vars))
		default:
			arrWithVars = append(arrWithVars, value)
		}
	}

	return arrWithVars
}
