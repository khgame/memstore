package cachekey

import (
	"fmt"
	"strings"
)

type (
	// KeyFormat
	//
	// Usage:
	//
	// General Way
	//
	// CacheSchemaFTPackage : combined with accountID, appID, name
	// CacheSchemaFTPackage cache.KeyFormat = "pockets:ft:%v:%v:%s"
	// CacheSchemaFTPackage.Make(accountID, appID, name)
	//
	// Partial Example
	// ```
	// func (p *FTPocket) cacheKey(accountID int64) func(...any) string {
	//	return CacheSchemaFTPackage.Partial(accountID)
	// }
	// delegate.Func2[int64, string, string](p.cacheKey).Partial(accountID)
	// ```
	KeyFormat string
)

// Make returns a string with args formatted by fmt.Sprintf
func (ck KeyFormat) Make(args ...any) string {
	return fmt.Sprintf(string(ck), args...)
}

// Split returns a list that contains all parts of the key scheme
func (ck KeyFormat) Split() []string {
	return strings.Split(string(ck), ":")
}

// Partial returns a function that accepts the rest of args and returns a string
func (ck KeyFormat) Partial(args ...any) func(...any) string {
	return func(rest ...any) string {
		return ck.Make(append(args, rest...)...)
	}
}

// Test - placeholder must after the colon
func (ck KeyFormat) Test() bool {
	// find all placeholders
	for i, c := range ck {
		if c == '%' {
			// check if the next char is after the colon
			if i > 0 && ck[i-1] != ':' && ck[i-1] != '%' {
				return false
			}
		}
	}
	return true
}
