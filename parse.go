package harg

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	internal "github.com/jtagcat/harg/internal"
)

type (
	Definitions struct {
		D       DefinitionMap
		Aliases map[string]*string // map[alias slug]defSlug
	}
	DefinitionMap map[string]Definition // map[slug]; 1-character: short option, >1: long option
	Definition    struct {
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
		originalType Type // for AlsoBool
		found        bool
		iface        option
	}
)

var (
	// end user (runtime) error
	ErrOptionHasNoDefinition = errors.New("option has no definition")
	ErrBoolAfterValue        = errors.New("AlsoBool does not accept bools after value inputs") // --foo=value --foo --foo=value
	ErrIncompatibleValue     = errors.New("")                                                  // TODO: strconv.Atoi("this is not a number")

	// library user (runtime) error (Definition.Foobar())
	ErrIncompatibleMethod = errors.New("method not compatible with type")

	// runtime error
	ErrInternalBug = errors.New("internal bug in harg or undefined enum") // anti-panic safetynet

	// depends on definitions (Parse() always fails):
	ErrSlugConflict = errors.New("conflicting same-named alias")
)

func (defs *Definitions) Parse(
	// [^chokes]: Chokes allow for global-local-whatever argument definitions by using Parse() multiple times.
	// args: "--foo", "bar", "chokename", "--foo", "differentDef"
	//          ^ parsed ^    ^choke, chokeReturn: "chokename", "--foo", "differentDef"

	args []string, // usually os.Args
	chokes []string, //[^chokes]// [case insensitive] parse arguments until first choke
	// Chokes are not seen after "--", or in places of argument values ("--foo choke", "-f choke")
) (
	// parsed options get added to defs (method parent)
	parsed []string, // non-options, arguments
	chokeReturn []string, //[^chokes]//  args[chokePos:], [0] is the found choke, [1:] are remaining unparsed args
	err error, // see above var(); errContext not provided: use fmt.Errorf("parsing arguments: %w", err)
) {
	args = args[1:] // remove program name TODO: should this be external?

	if err := defs.checkDefs(); err != nil {
		return nil, nil, err
	}
	chokeM := internal.SliceToLowercaseMap(chokes)

	var skipNext bool
	for i, a := range args {

		if skipNext {
			// (current) i is "next", signal to skip
			// as i-1 already parsed i as it's value
			skipNext = false
			continue
		}

		switch argumentKind(&a) {
		case e_argument:
			if _, isChoke := chokeM[strings.ToLower(a)]; isChoke {
				return parsed, args[i:], nil
			}
			parsed = append(parsed, a)

		case e_argumentDivider:
			if len(args)-1 != i { // there are more args
				parsed = append(parsed, args[i+1:]...)
			}
			return parsed, nil, nil

		case e_shortOption:
			skipNext, err = defs.parseShortOption(&i, &args)
			if err != nil {
				return nil, nil, err
			}

		case e_longOption:
			skipNext, err = defs.parseLongOption(&i, &args) // len(a) >= 3
			if err != nil {
				return nil, nil, err
			}
		}

	}
	return parsed, nil, nil
}

func (defs *Definitions) checkDefs() error {
	defs.D = internal.LowercaseLongMapNames(defs.D)
	defs.Aliases = internal.LowercaseLongMapNames(defs.Aliases)

	for n := range defs.D {
		if _, ok := defs.Aliases[n]; ok {
			return fmt.Errorf("option definition %s: %w", n, ErrSlugConflict)
		}
	}
	return nil
}

type argumentKindT int

const ( // enum
	e_argument        argumentKindT = iota
	e_argumentDivider               // "--"
	e_shortOption                   // "-something"
	e_longOption                    // "--something", len() >= 3
)

func argumentKind(arg *string) argumentKindT {
	if len(*arg) < 2 || !strings.HasPrefix(*arg, "-") {
		return e_argument // including "", "-"
	}

	// "-x"
	if !strings.HasPrefix((*arg)[1:], "-") {
		return e_shortOption
	}

	// begins with "--"
	switch utf8.RuneCountInString(*arg) {
	case 2: // "--"
		return e_argumentDivider
	case 3: // "--x", single negative short
		return e_shortOption
	default: // > 3
		return e_longOption
	}
}
