package harg_test

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/jtagcat/harg"
	"github.com/stretchr/testify/assert"
)

// see FORMAT.md for what test is
// responsible for what part of the spec

func ExampleDefinitions_Parse() {
	kOne, kTwo, kThree := "o", "t", "three"
	defs := harg.Definitions{
		kOne:   {Type: harg.String},
		kTwo:   {Type: harg.Bool},
		kThree: {Type: harg.Duration},
	}

	osArgs := strings.Split("programName hello -to foo -o bar --three 5s --t -t -t world", " ")

	args, _, err := defs.Parse(osArgs[1:], nil)
	if err != nil {
		log.Fatalf("parsing command-line arguments: %e", err)
	}

	fmt.Println(args) // [hello world]

	sl, ok := defs[kOne].SlString() // ok: is valid and set
	if ok {
		fmt.Println(sl) // [foo bar]
	}
	s, ok := defs[kOne].String()
	if ok {
		fmt.Println(s) // bar
	}

	two, _ := defs[kTwo].SlBool()
	fmt.Println(two) // [true false true true]

	count, _ := defs[kTwo].Count() // how many true in a row
	fmt.Println(count)             // 2

	dur, _ := defs[kThree].Duration()
	fmt.Println(dur) // 5s

	// Output:
	// [hello world]
	// [foo bar]
	// bar
	// [true false true true]
	// 2
	// 5s
}

func TestParseNilDefs(t *testing.T) {
	t.Parallel()

	defs := harg.Definitions{}

	args, chokeReturn, err := defs.Parse([]string{
		"hello", "-", "world",
		"cHOKe", "return",
	},
		[]string{"choke"},
	)

	assert.Nil(t, err)
	assert.Equal(t, []string{"hello", "-", "world"}, args)
	assert.Equal(t, []string{"cHOKe", "return"}, chokeReturn)

	args, chokeReturn, err = defs.Parse(nil, nil)
	assert.Nil(t, err)
	assert.Nil(t, chokeReturn)
	assert.Nil(t, args)
}

func TestParseDoubledash(t *testing.T) {
	t.Parallel()

	defs := harg.Definitions{}

	args, chokeReturn, err := defs.Parse([]string{
		"hello", "world",
		"--",
		"choke",
		"--argument",
		"-a",
	},
		[]string{"choke"},
	)

	assert.Nil(t, err)
	assert.Nil(t, chokeReturn)
	assert.Equal(t, []string{"hello", "world", "choke", "--argument", "-a"}, args)
}

func TestAliasParse(t *testing.T) {
	t.Parallel()

	oneKey := "one"

	defs := harg.Definitions{
		oneKey: {Type: harg.String},
	}
	assert.Nil(t, defs.Alias("twõか", oneKey))

	args, chokeReturn, err := defs.Parse([]string{
		"hello",
		"--one=one",
		"--twõか", "two",
		"world",
	}, nil,
	)

	assert.Nil(t, err)
	assert.Nil(t, chokeReturn)
	assert.Equal(t, []string{"hello", "world"}, args)

	sl, ok := defs[oneKey].SlString()
	assert.Equal(t, true, ok)
	assert.Equal(t, []string{"one", "two"}, sl)

	s, ok := defs[oneKey].String()
	assert.Equal(t, true, ok)
	assert.Equal(t, "two", s)
}

func TestParseLongOptEat(t *testing.T) {
	t.Parallel()

	oneKey, twoKey, fooKey := "oかe", "t", "f"

	defs := harg.Definitions{
		oneKey: {Type: harg.String},
		twoKey: {},
		fooKey: {},
	}

	args, chokeReturn, err := defs.Parse([]string{
		"hello",
		"--OかE=-t",
		"--oかE", "-f",
		"--oかe",
		"world",
	}, []string{"world"},
	)

	assert.Nil(t, err)
	assert.Nil(t, chokeReturn)
	assert.Equal(t, []string{"hello"}, args)

	sl, ok := defs[oneKey].SlString()
	assert.Equal(t, true, ok)
	assert.Equal(t, []string{"-t", "", "world"}, sl)

	assert.Equal(t, true, defs[twoKey].Default())
	assert.Equal(t, false, defs[fooKey].Default())
}

func TestParseShortOptEat(t *testing.T) {
	t.Parallel()

	oneKey, twoKey, fooKey := "か", "t", "f"

	defs := harg.Definitions{
		oneKey: {Type: harg.String},
		twoKey: {},
		fooKey: {},
	}

	args, chokeReturn, err := defs.Parse([]string{
		"hello",
		"-かt",
		"-か=-t",
		"-か", "=-t",
		"-か", "-f",
		"-か",
		"world",
	}, []string{"world"},
	)

	assert.Nil(t, err)
	assert.Nil(t, chokeReturn)
	assert.Equal(t, []string{"hello"}, args)

	sl, ok := defs[oneKey].SlString()
	assert.Equal(t, true, ok)
	assert.Equal(t, []string{"t", "-t", "=-t", "", "world"}, sl)

	assert.Equal(t, true, defs[twoKey].Default())
	assert.Equal(t, false, defs[fooKey].Default())
}

func TestParseShortBoolOpt(t *testing.T) {
	t.Parallel()

	zeroKey, oneKey, twoKey := "か", "õ", "x"
	unsetKey := "u"

	defs := harg.Definitions{
		zeroKey:  {},
		oneKey:   {},
		twoKey:   {},
		unsetKey: {},
	}
	assert.Nil(t, defs.Alias("õx", zeroKey))

	for in, want := range map[string][]bool{
		"-か":      {true, false, false},
		"-か\n--か": {false, false, false},
		"-かõ-x":   {true, true, false},
		"-か-õx":   {true, false, true},
		"-か-õ-x":  {true, false, false},
		"--õx":    {true, false, false},
		"---õx":   {false, false, false},
	} {
		defs := defs

		args, chokeReturn, err := defs.Parse(
			strings.Split(in, "\n"), nil,
		)

		assert.Nil(t, err)
		assert.Nil(t, chokeReturn)
		assert.Nil(t, args)

		set := defs[unsetKey].Default()
		assert.Equal(t, true, set)

		b, ok := defs[zeroKey].Bool()
		assert.Equal(t, true, ok)
		assert.Equal(t, want[0], b)

		b, _ = defs[oneKey].Bool()
		assert.Equal(t, want[1], b)

		b, _ = defs[twoKey].Bool()
		assert.Equal(t, want[2], b)
	}
}

func TestParseCount(t *testing.T) {
	t.Parallel()

	// also responsible for testing if typeMap.new() actually copies or no

	zeroKey, oneKey := "a", "b"
	defs := harg.Definitions{
		zeroKey: {},
		oneKey:  {},
	}

	args, chokeReturn, err := defs.Parse([]string{
		"-a-aaaa-a",
		"--b", "-b-b-bbb",
	}, nil,
	)

	assert.Nil(t, err)
	assert.Nil(t, chokeReturn)
	assert.Nil(t, args)

	sl, ok := defs[zeroKey].SlBool()
	assert.Equal(t, true, ok)
	assert.Equal(t, []bool{true, false, true, true, true, false}, sl)
	c, ok := defs[zeroKey].Count()
	assert.Equal(t, true, ok)
	assert.Equal(t, 0, c)

	sl, ok = defs[oneKey].SlBool()
	assert.Equal(t, true, ok)
	assert.Equal(t, []bool{false, true, false, false, true, true}, sl)
	c, ok = defs[oneKey].Count()
	assert.Equal(t, true, ok)
	assert.Equal(t, 2, c)
}

func TestParseLongOptAlsoBool(t *testing.T) {
	t.Parallel()

	oneKey, twoKey := "foo", "bar"

	defs := harg.Definitions{
		oneKey: {Type: harg.String, AlsoBool: true},
		twoKey: {Type: harg.String, AlsoBool: true},
	}

	args, chokeReturn, err := defs.Parse([]string{
		"---foo", "bar", // false
		"--foo", "bar", // true
		"--bar=true", // "true", not true
	}, nil,
	)

	assert.Nil(t, err)
	assert.Nil(t, chokeReturn)
	assert.Equal(t, []string{"bar", "bar"}, args)

	sl, ok := defs[oneKey].SlBool()
	assert.Equal(t, true, ok)
	assert.Equal(t, []bool{false, true}, sl)

	s, ok := defs[twoKey].String()
	assert.Equal(t, true, ok)
	assert.Equal(t, "true", s)
}

func TestParseError(t *testing.T) {
	t.Parallel()

	defs := harg.Definitions{
		"str":      {Type: harg.String},
		"bool":     {},
		"alsobool": {Type: harg.String, AlsoBool: true},
	}

	for _, test := range []errTest{
		// Negating long option
		{in: []string{"---str"}, errIs: harg.ErrIncompatibleValue},       // not bool
		{in: []string{"---bool=true"}, errIs: harg.ErrIncompatibleValue}, // bool with value

		// AlsoBool after Value
		{in: []string{"--alsobool=val", "--alsobool"}, errIs: harg.ErrIncompatibleValue},

		// No definition
		{in: []string{"--nodef"}, errIs: harg.ErrOptionHasNoDefinition},
		{in: []string{"-n"}, errIs: harg.ErrOptionHasNoDefinition},

		// Some errors are tested in definition tests.
	} {
		defs := defs

		args, chokeReturn, err := defs.Parse(
			test.in, nil,
		)

		assert.ErrorIs(t, err, test.errIs)
		assert.Nil(t, chokeReturn)
		assert.Nil(t, args)
	}
}

type errTest struct {
	in    []string
	errIs error
}

func TestGetNormalizedKey(t *testing.T) {
	one := "hElLO" // will be lowercased
	defs := harg.Definitions{
		one: {},
	}

	args, chokeReturn, err := defs.Parse([]string{
		"--hello", "--HELlO", // any case should work
	}, nil)
	assert.Nil(t, err)
	assert.Nil(t, chokeReturn)
	assert.Nil(t, args)

	c, ok := defs[one].Count()
	assert.Equal(t, true, ok)
	assert.Equal(t, 2, c)
}

func TestGetNormalizedEnvKey(t *testing.T) {
	one := "hElLO world" // will be uppercased and joined with underscore
	defs := harg.Definitions{
		one: {},
	}

	assert.Nil(t, os.Setenv("HELLO_wORlD", "true"))

	assert.Nil(t, defs.ParseEnv())

	b, ok := defs[one].Bool()
	assert.Equal(t, true, ok)
	assert.Equal(t, true, b)
}
