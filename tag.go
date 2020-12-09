package opt

import (
	"reflect"
	"regexp"
	"strings"
)

var (
	shortRgx = regexp.MustCompile(`^(\?|\[\]|[A-Za-z]{1})$`)
	longRgx  = regexp.MustCompile(`^[a-z]{1}[a-z\-1-9]{1,}$`)
)

// The tagSample is struct of the fields of the tag for opt's package.
// Tag example: "opt:[short[,long[,value[,help]]]]" where:
//
//    short - short name of the option (one char from the A-Za-z);
//    long - long name of the option (at least two characters);
//    value - default value;
//    help - help information.
type tagSample struct {
	Short string
	Long  string
	Value string
	Help  string
}

type fieldSample struct {
	TagSample *tagSample
	Item      *reflect.Value
}

// The isValid returns true if key name is valid.
func (args tagSample) IsValid() bool {
	return (len(args.Short) != 0 && shortRgx.Match([]byte(args.Short))) ||
		(len(args.Long) != 0 && longRgx.Match([]byte(args.Long)))
}

// The isIgnored returns true if key name is "-" or incorrect.
func (args tagSample) IsIgnored() bool {
	return !args.IsValid() || args.Short == "-"
}

// The getTagValues returns field valueas as array: [flag, cmd, value, help].
func getTagValues(tag string) (r [4]string) {
	var chunks = splitN(tag, ",", 4)
	for i, c := range chunks {
		// Save the last piece without changed.
		if i == len(r)-1 {
			if v := strings.Join(chunks[i:], ","); v != "" {
				r[i] = v
			}
			break
		}

		r[i] = c
	}

	return
}

// The getTagArgs returns tagSample object for tag.
// If key isn't sets in the tag, it will be assigned from the second argument.
//
// Examples
//
//    getTagArgs("a,b,c,d", "default") // [a b c d]
//    getTagArgs(",b,c,d", "default")  // [ b c d]
//    getTagArgs("a", "default")       // [a]
//    getTagArgs(",,c,d", "default")   // [  defaulr c d]
func getTagSample(tag, cmd string) *tagSample {
	var args, v = &tagSample{}, getTagValues(tag)

	// Short.
	args.Short = strings.Trim(v[0], " ")

	// Long.
	args.Long = strings.Trim(v[1], " ")
	if len(args.Short) == 0 && len(args.Long) == 0 {
		args.Long = cmd
	}

	// Value.
	args.Value = strings.TrimRight(strings.TrimLeft(v[2], "({[\"'`"), ")}]\"'`")

	// Help.
	args.Help = strings.Trim(v[3], " ")

	return args
}
