package internal

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func SliceLowercaseIndex(s []string) map[string]struct{} {
	index := make(map[string]struct{})

	for _, str := range s {
		index[strings.ToLower(str)] = struct{}{}
	}

	return index
}

func LowercaseLongKey(key string) string {
	if utf8.RuneCountInString(key) == 1 {
		return key
	}

	return strings.ToLower(key)
}

func KeyErrorName(key string) string {
	var keyType string
	if utf8.RuneCountInString(key) > 1 {
		keyType = "long"
	} else {
		keyType = "short"
	}

	return fmt.Sprintf("%s option %s", keyType, key)
}
