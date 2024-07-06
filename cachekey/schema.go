package cachekey

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/khicago/got/util/typer"

	"github.com/bagaking/goulp/wlog"
	"github.com/khicago/irr"
)

const (
	cacheKeyTag     = "cachekey"
	MAXPlaceHolders = 1000
)

var ErrSchemaValidateFailed = irr.Error("invalid schema for given type")

// KeySchema used to build a real cache key
//
// Standard Usages:
//
//  1. config with struct
//     type A struct {
//     XXX string `cachekey:"xxx_1"`
//     YYY int
//     }
//     var ckb = MustNewSchema[A]("key:{xxx_1}:{yyy}:{xxx_1}", 10*time.Minute)
//
//  2. config with single value
//     var ckb = MustNewSchema[int64]("key:{xxx_1}", 10*time.Minute)
//
//  3. build cache key
//     ckb.Build(A{XXX: "xxx", YYY: 1})
//
//  4. build cache key in dynamic way
//     ckb.ToFormat().Make("xxx", 1, "xxx")
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
	if !ValidateParamsType[ParamsTable](schema) {
		return nil, irr.Wrap(ErrSchemaValidateFailed, "fields: %v", typer.Keys(StructToMap(typer.ZeroVal[ParamsTable]())))
	}
	ck := &KeySchema[ParamsTable]{
		schema: schema,
		exp:    exp,
	}
	return ck, nil
}

// FingerPrint returns a unique string for the schema
func (ckb *KeySchema[ParamsTable]) FingerPrint() string {
	return ckb.schema
}

// Build a real cache key
func (ckb *KeySchema[ParamsTable]) Build(params ParamsTable) (string, error) {
	var paramsMap map[string]any

	t := reflect.TypeOf(params)
	// if params is a struct, convert it to map
	if t.Kind() == reflect.Struct || (t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct) {
		paramsMap = make(map[string]any)
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
		return replacePlaceholders(ckb.schema, paramsMap)
	}

	// if params is a single value, convert it to single value map
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
			wlog.Common("to_format").WithError(err).Errorf("error parsing schema: %s", schema)
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
