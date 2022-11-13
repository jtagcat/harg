package harg

import (
	"fmt"
	"strconv"
	"time"

	internal "github.com/jtagcat/harg/internal"
)

func (def *Definition) parseOptionContent(
	key string,
	value string, // "" means literally empty, caller has already defaulted booleans to true
) error { // errContext provided

	if def.parsed.found { // TODO:
	}

	if def.AlsoBool {
		boolFace := typeMetaM[Bool].emptyT

		err := boolFace.add(value)
		if err == nil {
			if def.parsed.found && def.Type != Bool { // we have already parsed opt with native type
				return fmt.Errorf("parsing %s as %s (AlsoBool): %w", internal.KeyErrorName(key), typeMetaM[Bool].name, ErrBoolAfterValue)
			}

			// TODO: broken asw, overwriting stuff
			def.parsed.originalType = def.Type
			def.Type, def.parsed.found, def.parsed.opt = Bool, true, boolFace
			return nil
		}

		// we have parsed it as bool

		// non-bool AlsoBool continues to switch
		if def.parsed.found { // restore original
			def.Type = def.parsed.originalType
			if def.Type == Bool { // discard previous bools
				def.parsed.found = false
			}
		}
	}

	// valueful
	if !def.parsed.found {
		def.parsed.opt = typeMetaM[def.Type].emptyT
	}

	err := def.parsed.opt.add(value)
	if err != nil {
		return fmt.Errorf("parsing %s as %s: %e: %w", internal.KeyErrorName(key), typeMetaM[def.Type], ErrIncompatibleValue, err)
	}

	def.parsed.found = true
	return nil
}

type option interface {
	contents() any           // resolved with option.Sl
	add(rawOpt string) error // string: type name (to use in error)
}

type Type uint32 // enum:
const (
	Bool Type = iota
	String
	Int
	Int64
	Uint
	Uint64
	Float64
	Duration
)

var typeMetaM = map[Type]struct {
	name   string
	emptyT option
}{
	Bool:     {"bool", &optBool{}},
	String:   {"string", &optString{}},
	Int:      {"int", &optInt{}},
	Int64:    {"int64", &optInt64{}},
	Uint:     {"uint", &optUint{}},
	Uint64:   {"uint64", &optUint64{}},
	Float64:  {"float64", &optFloat64{}},
	Duration: {"duration", &optDuration{}},
}

// bool / count

type (
	optBool struct {
		value optBoolVal
	}
	optBoolVal struct {
		count int
		value []bool
	}
)

func (o *optBool) contents() any {
	return o.value
}

func (o *optBool) add(s string) error {
	v, err := strconv.ParseBool(s) // TODO: drop "t", "f", add "yes", "no", maybe also "y", "n"
	if err != nil {
		return err
	}

	if v == true {
		o.value.count++
	} else {
		o.value.count = 0
	}

	o.value.value = append(o.value.value, v)
	return nil
}

// string

type optString struct {
	value []string
}

func (o *optString) contents() any {
	return o.value
}

func (o *optString) add(s string) error {
	o.value = append(o.value, s)
	return nil
}

// int

type optInt struct {
	value []int
}

func (o *optInt) contents() any {
	return o.value
}

func (o *optInt) add(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		return err
	}

	o.value = append(o.value, int(v))
	return err
}

// int64

type optInt64 struct {
	value []int64
}

func (o *optInt64) contents() any {
	return o.value
}

func (o *optInt64) add(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return err
	}

	o.value = append(o.value, int64(v))
	return err
}

// uint

type optUint struct {
	value []uint
}

func (o *optUint) contents() any {
	return o.value
}

func (o *optUint) add(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		return err
	}

	o.value = append(o.value, uint(v))
	return err
}

// uint64

type optUint64 struct {
	value []uint64
}

func (o *optUint64) contents() any {
	return o.value
}

func (o *optUint64) add(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return err
	}

	o.value = append(o.value, uint64(v))
	return err
}

// float64

type optFloat64 struct {
	value []float64
}

func (o *optFloat64) contents() any {
	return o.value
}

func (o *optFloat64) add(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}

	o.value = append(o.value, float64(v))
	return err
}

// duration

type optDuration struct {
	value []time.Duration
}

func (o *optDuration) contents() any {
	return o.value
}

func (o *optDuration) add(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	o.value = append(o.value, time.Duration(v))
	return err
}

// timestamp
// TODO:

// ip
// ipv4
// ipv6
// TODO: