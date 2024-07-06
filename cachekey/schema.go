package cachekey

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/bagaking/goulp/jsonex"
	"github.com/bagaking/goulp/wlog"
	"github.com/khicago/got/util/typer"
	"github.com/khicago/irr"
)

const MAXPlaceHolders = 1000

// KeySchema used to build a real cache key
//
// Usage:
//
//	type A struct {
//	    XXX string
//	    YYY int
//	}
//
// var ckb = MustNewSchema[A]("key:{xxx}:{yyy}", 10*time.Minute)
//
// ... ckb.build()
type KeySchema[ParamsTable any] struct {
	schema string
	exp    time.Duration
	extra  any
}

// MustNewSchema create new schema, panic when failed
func MustNewSchema[ParamsTable any](schema string, exp time.Duration) *KeySchema[ParamsTable] {
	ck, err := NewSchema[ParamsTable](schema, exp)
	if err != nil {
		wlog.Common("MustNewSchema").WithError(err).Panicf("failed to build cache key: %s", schema)
	}
	return ck
}

// NewSchema create new schema
func NewSchema[ParamsTable any](schema string, exp time.Duration) (*KeySchema[ParamsTable], error) {
	ck := &KeySchema[ParamsTable]{
		schema: schema,
		exp:    exp,
	}

	if _, err := ck.Build(typer.ZeroVal[ParamsTable]()); err != nil {
		return nil, err
	}
	return ck, nil
}

// Build 方法中应用驼峰转蛇形
func (ckb *KeySchema[ParamsTable]) Build(params ParamsTable) (string, error) {
	var paramsMap map[string]any
	var err error

	// 检查 ParamsTable 是否为结构体或指向结构体的指针
	t := reflect.TypeOf(params)
	if t.Kind() == reflect.Struct || (t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct) {
		// 如果是结构体或指向结构体的指针，使用现有的 structToMap 方法
		paramsMap, err = structToMap(params)
		if err != nil {
			return "", err
		}
	} else {
		// 如果不是结构体，直接使用一个占位符替换
		name, start, end, _ := getNextPlaceholder(ckb.schema, 0)
		if start == -1 {
			return "", irr.Error("schema %s has no placeholders", ckb.schema)
		}
		paramsMap = map[string]any{
			name: fmt.Sprintf("%v", params),
		}
		_, start, _, _ = getNextPlaceholder(ckb.schema, end+1)
		if start != -1 {
			return "", irr.Error("value table not match, map= %v, params= %v", paramsMap, params)
		}
	}

	return replacePlaceholders(ckb.schema, paramsMap)
}

func (ckb *KeySchema[ParamsTable]) MustBuild(params ParamsTable) string {
	paramsMap, err := ckb.Build(params)
	if err != nil {
		panic(err)
	}
	return paramsMap
}

func (ckb *KeySchema[ParamsTable]) SetExp(exp time.Duration) *KeySchema[ParamsTable] {
	ckb.exp = exp
	return ckb
}

func (ckb *KeySchema[ParamsTable]) GetExp() time.Duration {
	return ckb.exp
}

func (ckb *KeySchema[ParamsTable]) SetExtra(extra any) *KeySchema[ParamsTable] {
	ckb.extra = extra
	return ckb
}

func (ckb *KeySchema[ParamsTable]) GetPlaceholders() []string {
	return getPlaceholders(ckb.schema)
}

func (ckb *KeySchema[ParamsTable]) ToFormat() KeyFormat {
	schema := ckb.schema
	result := strings.Builder{}
	curBeginPos := 0

	for i := 0; i < MAXPlaceHolders; i++ {
		_, start, end, err := getNextPlaceholder(schema, curBeginPos)
		if err != nil {
			wlog.Common("ToFormat").WithError(err).Errorf("Error parsing schema: %s", schema)
			return KeyFormat(schema)
		}
		if start == -1 {
			result.WriteString(schema[curBeginPos:])
			break
		}

		result.WriteString(schema[curBeginPos:start])
		result.WriteString("%v")

		curBeginPos = end + 1
	}

	return KeyFormat(result.String())
}

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
		return "", -1, -1, errors.New("malformed schema: unclosed placeholder")
	}
	end += start

	placeholderName := schema[start+1 : end]
	return placeholderName, start, end, nil
}

// replacePlaceholders replaces placeholders in the schema with values from paramsMap.
func replacePlaceholders(schema string, paramsMap map[string]any) (string, error) {
	var result strings.Builder
	curBeginPos := 0

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
			return "", fmt.Errorf("field %s does not exist, paramsMap= %+v", placeholderName, paramsMap)
		}
		result.WriteString(fmt.Sprintf("%v", fieldValue))

		curBeginPos = end + 1
	}

	return result.String(), nil
}

// structToMap 使用反射将结构体转换为映射
func structToMap(item any) (map[string]any, error) {
	str, err := jsonex.MarshalToString(item)
	if err != nil {
		return nil, err
	}

	result := make(map[string]any)
	if err = jsonex.UnmarshalFromString(str, &result); err != nil {
		return nil, err
	}
	return result, nil
}
