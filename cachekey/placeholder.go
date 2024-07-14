package cachekey

import (
	"fmt"
	"strings"

	"github.com/khicago/irr"
)

// getPlaceholders returns a list of placeholders in the schema.
func getPlaceholders(schema string) []string {
	var placeholders []string
	curBeginPos := 0
	for i := 0; i < MAXPlaceHolders; i++ {
		name, _, end, _ := getNextPlaceholder(schema, curBeginPos)
		if end == -1 {
			break
		}
		placeholders = append(placeholders, name)
		curBeginPos = end + 1
	}

	return placeholders
}

// getNextPlaceholder finds the next placeholder in the schema and returns its name and the positions of the braces.
func getNextPlaceholder(schema string, startPos int) (string, int, int, error) {
	start := strings.Index(schema[startPos:], "{")
	if start == -1 {
		return "", -1, -1, nil
	}
	start += startPos

	end := strings.Index(schema[start:], "}")
	if end == -1 {
		return "", -1, -1, irr.Error("malformed schema: unclosed placeholder")
	}
	end += start

	placeholderName := schema[start+1 : end]
	return placeholderName, start, end, nil
}

// replacePlaceholders replaces placeholders in the schema with values from paramsMap.
func replacePlaceholders(schema string, paramsMap map[string]any) (string, error) {
	var result strings.Builder
	curBeginPos := 0

	// the count of placeholders can be larger than the count of fields in the struct, cuz placeholders can repeat
	// the loop should not run more than MAXPlaceHolders times, to prevent infinite loop
	for i := 0; i < MAXPlaceHolders; i++ {
		placeholderName, start, end, err := getNextPlaceholder(schema, curBeginPos)
		if err != nil {
			return "", err
		}
		if start == -1 {
			result.WriteString(schema[curBeginPos:])
			break
		}

		result.WriteString(schema[curBeginPos:start])

		fieldValue, exists := paramsMap[placeholderName]
		if !exists {
			return "", irr.Error("field %s does not exist, paramsMap= %+v", placeholderName, paramsMap)
		}
		result.WriteString(fmt.Sprintf("%v", fieldValue))

		curBeginPos = end + 1
	}

	return result.String(), nil
}
