package cachekey

import (
	"reflect"
	"strings"
	"unicode"
)

// ValidateParamsType checks if the given schema is valid for the given type.
func ValidateParamsType[ParamsTable any](schema string) bool {
	t := reflect.TypeOf((*ParamsTable)(nil)).Elem()
	placeholders := getPlaceholders(schema)

	if t.Kind() == reflect.Struct || (t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct) {
		// Create a zero value of the type and convert it to a map
		zeroValue := reflect.New(t).Interface()
		paramsMap := StructToMap(zeroValue)

		// Check if all placeholders exist in the map
		for _, p := range placeholders {
			if _, exists := paramsMap[p]; !exists {
				return false
			}
		}
		return true
	}

	return len(placeholders) == 1
}

// StructToMap converts a struct to a map.
// The key of the map is the field name of the struct, and the value is the field value.
// The algorithm for converting a field name to a cache key is as follows:
// 1. If the field has a cachekey tag, use the tag value as the cache key.
// 2. If the field does not have a cachekey tag, convert the field name to snake_case and use it as the cache key.
func StructToMap[ParamsTable any](params ParamsTable) map[string]any {
	paramsMap := make(map[string]any)
	v := reflect.ValueOf(params)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldName := getFieldName(field)
		paramsMap[fieldName] = v.Field(i).Interface()
	}
	return paramsMap
}

// getFieldName returns the name of the field to be used as a placeholder in the cache key.
func getFieldName(field reflect.StructField) string {
	if tag := field.Tag.Get(cacheKeyTag); tag != "" {
		return tag
	}
	// if the cachekey tag are not exist，using snake_case
	return toSnakeCase(field.Name)
}

// toSnakeCase converts a string to snake_case.
// where all letters are lowercase and words are separated by underscores. For example, "FooBar" becomes "foo_bar".
// for the continue upper case, it will be converted to lower case and no underscore will be added between them.
// for example, when the string is "HTTPCode", it should be converted to "http_code".
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			// 检查前一个字符是否为小写，或者后一个字符是否为小写（如果存在）
			if unicode.IsLower(rune(s[i-1])) || (i+1 < len(s) && unicode.IsLower(rune(s[i+1]))) {
				result.WriteRune('_')
			}
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}
