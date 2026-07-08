package opt

import (
	"errors"
	"os"
)

// Unmarshal parses the argument-line options and stores the result
// to go-struct. If obj is nil or is not a pointer to a non-empty struct,
// it returns an error (it does not panic).
//
// Unmarshal method supports the following field's types: int, int8, int16,
// int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string,
// bool, url.URL and pointers, array or slice from those types (i.e. *int, ...,
// []int, ..., []bool, ..., [2]*url.URL, etc.).
//
// The function generates a panic only for a struct-tag misconfiguration - a
// developer error visible on the first run:
//
// - a non-string type field is specified for the `opt:"?"` doc-field;
// - field for positional arguments `opt:"[]"` is not a list (slice/array);
// - field has structure type (except url.URL);
// - field has pointer to structure type (except *url.URL);
// - two fields declare the same option name.
//
// Use the following tags in the fields of structure to
// set the marshalling parameters:
//
//	opt  indicates a short or long option;
//	alt  optional, alternative option for position opt,
//	     if a long option is specified in opt, a short option
//	     can be specified in alt or vice versa;
//	def  default value (if empty, sets the default value
//	     for the field type of structure);
//	help brief help about the option.
//
// Suppose that the some values was set into argument-line as:
//
//	./main --host=0.0.0.0 -p8080
//
// Structure example:
//
//	// Args structure for containing values from the argument-line.
//	type Args struct {
//		Host string `opt:"host" def:"localhost"`
//		Port int    `opt:"p" alt:"port" def:"80" help:"port number"`
//		Help bool   `opt:"h" alt:"help"`
//	}
//
// Unmarshal data from the argument-line into Args struct.
//
//	var args Args
//	if err := opt.Unmarshal(&args); err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Host: %s\nPort: %d\n", args.Host, args.Port)
//	// Output:
//	//  Host: 0.0.0.0
//	//  Port: 8080
func Unmarshal(obj any) error {
	return UnmarshalArgs(obj, os.Args)
}

// UnmarshalArgs is like Unmarshal but parses the given argument list instead
// of os.Args. The first element is treated as the program path (os.Args[0]),
// so pass it the same shape as os.Args, e.g.:
//
//	err := opt.UnmarshalArgs(&cfg, []string{"app", "--host", "x", "-p", "80"})
//
// This is convenient for tests, sub-commands and REPLs.
func UnmarshalArgs(obj any, args []string) error {
	// Join all collected errors so the caller sees every problem at once
	// (errors.Join returns nil for an empty/all-nil slice).
	return errors.Join(unmarshalOpt(obj, args)...)
}
