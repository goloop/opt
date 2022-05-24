package opt

import (
	"os"
)

// Unmarshal parses the argument-line options and stores the result
// to go-struct. If the obj isn't a pointer to struct or is nil -
// returns an error.
//
// Unmarshal method supports the following field's types: int, int8, int16,
// int32, int64, uin, uint8, uin16, uint32, in64, float32, float64, string,
// bool, url.URL and pointers, array or slice from thous types (i.e. *int, ...,
// []int, ..., []bool, ..., [2]*url.URL, etc.).
//
// For other filed's types (like chan, map ...) will be returned an error.
//
// The function generates a panic if:
//
// - the object isn't a structure;
// - the object isn't transmitted by pointer;
// - a non-string type field is specified for the `opt:"?"` doc-field;
// - field for positional arguments `opt:"[]"` is not a list (slice/array);
// - field has structure type (except url.URL);
// - field has pointer to structure type (except *url.URL).
//
// Use the following tags in the fields of structure to
// set the marshing parameters:
//
//  opt  indicates a short or long option;
//  alt  optional, alternative option for position opt,
//       if a long option is specified in opt, a short option
//       can be specified in alt or vice versa;
//  def  default value (if empty, sets the default value
//       for the field type of structure);
//  help brief help about the option.
//
// Suppose that the some values was set into argument-line as:
//
//  ./main --host=0.0.0.0 -p8080
//
// Structure example:
//
//  // Args structure for containing values from the argument-line.
//  type Args struct {
//  	Host string `opt:"host" def:"localhost"`
//  	Port int    `opt:"p" alt:"port" def:"80" help:"port number"`
//  	Help bool   `opt:"h" alt:"help"`
//  }
//
// Unmarshal data from the argument-line into Args struct.
//
//  var args Args
//  if err := opt.Unmarshal(&args); err != nil {
//  	log.Fatal(err)
//  }
//
//  fmt.Printf("Host: %s\nPort: %d\n", args.Host, args.Port)
//  // Output:
//  //  Host: 0.0.0.0
//  //  Port: 8080
func Unmarshal(obj interface{}) error {
	if errs := unmarshalOpt(obj, os.Args); errs != nil {
		// Returns only the first error.
		return errs[0]
	}

	return nil
}
