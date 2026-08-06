package validation

import (
	"reflect"
	"strings"
	"sync"
)

type rulesCacheKey struct {
	typeOf  reflect.Type
	tagName string
}

var rulesCache sync.Map

// HasRules reports whether value contains an effective validate tag. Nested
// struct fields are inspected because validator validates them recursively.
func HasRules(value any) bool {
	return HasRulesForTag(value, "validate")
}

// HasRulesForTag reports whether value contains an effective validation tag
// with the supplied name. It can be used with binding or custom tag names.
func HasRulesForTag(value any, tagName string) bool {
	typeOf := indirectType(reflect.TypeOf(value))
	tagName = strings.TrimSpace(tagName)
	if typeOf == nil || typeOf.Kind() != reflect.Struct || tagName == "" {
		return false
	}

	cacheKey := rulesCacheKey{
		typeOf:  typeOf,
		tagName: tagName,
	}
	if cached, ok := rulesCache.Load(cacheKey); ok {
		return cached.(bool)
	}
	hasRules := typeHasRules(typeOf, tagName, make(map[reflect.Type]bool))
	rulesCache.Store(cacheKey, hasRules)
	return hasRules
}

func typeHasRules(typeOf reflect.Type, tagName string, visiting map[reflect.Type]bool) bool {
	typeOf = indirectType(typeOf)
	if typeOf == nil || typeOf.Kind() != reflect.Struct || visiting[typeOf] {
		return false
	}
	visiting[typeOf] = true
	defer delete(visiting, typeOf)

	for index := 0; index < typeOf.NumField(); index++ {
		field := typeOf.Field(index)
		if field.PkgPath != "" {
			continue
		}
		tag, exists := field.Tag.Lookup(tagName)
		tag = strings.TrimSpace(tag)
		if exists && tag == "-" {
			continue
		}
		if exists && tag != "" {
			return true
		}
		if typeHasRules(field.Type, tagName, visiting) {
			return true
		}
	}
	return false
}

func indirectType(typeOf reflect.Type) reflect.Type {
	for typeOf != nil && typeOf.Kind() == reflect.Pointer {
		typeOf = typeOf.Elem()
	}
	return typeOf
}
