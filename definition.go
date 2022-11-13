package harg

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func (defs Definitions) SetAlias(name string, to string) error {
	defP, ok := defs[to]
	if !ok {
		return fmt.Errorf("definition name %s: %w", to, ErrOptionHasNoDefinition)
	}

	defs[name] = defP
	return nil
}

type (
	Definitions map[string]*Definition // map[slug]; 1-character: short option, >1: long option
	Definition  struct {
		Type Type

		// For short options (1-char length), true means it's always bool
		// For long options:
		//   false: allows spaces (`--slug value` in addition to `--slug=value`)
		//   true: if "=" is not used, Type is changed to bool (or countable). Values are treated as bools, if strconv.ParseBool says so.
		// If bool is encountered after value, ErrBoolAfterValue will be returned on parsing. Any bools before value flags will be ignored.
		AlsoBool bool

		// use Definition.Methods() to get data, #TODO:
		parsed parsedT
	}
	parsedT struct {
		originalType Type // when AlsoBool
		found        bool
		opt          option
	}
)

func (defs Definitions) normalize() {
	for name, def := range defs {
		if def == nil {
			delete(defs, name)
			continue
		}

		// short args are case sensitive, skip
		if utf8.RuneCountInString(name) == 1 {
			continue
		}

		// case insensitivize long args
		lower := strings.ToLower(name)
		if name != lower {
			defs[lower] = def
			delete(defs, name)
		}
	}
}