// Package opt provides command-line argument parsing and management
// for Go applications.
//
// Unlike Go's standard flag package which follows Plan 9 style (single
// dash flags only), this package implements Linux-style argument parsing
// with support for both short (-v) and long (--verbose) flags to match
// common Unix/Linux command line conventions.
//
// Features:
// - Parses command-line arguments into Go structs using struct tags
// - Supports both short (-v) and long (--verbose) flag formats (Linux-style)
// - Handles positional arguments with automatic ordering
// - Provides default values through struct tags
// - Generates help documentation automatically
// - Supports grouped short flags (-abc equivalent to -a -b -c)
// - Allows flag aliases through the alt tag
//
// Supported field types:
// - Basic types: int, int8, int16, int32, int64
// - Unsigned types: uint, uint8, uint16, uint32, uint64
// - Floating point: float32, float64
// - Other basic types: string, bool
// - URL types: url.URL and *url.URL
// - Arrays and slices of the above types
//
// Struct tags:
// - opt: Defines the primary flag name (required)
// - alt: Defines an alternative flag name (optional)
// - def: Sets the default value (optional)
// - sep: Specifies list separator for array/slice types (optional)
// - help: Provides help text for documentation (optional)
//
// Special opt tag values:
// - "?" : Field will store generated help text
// - "[]": Field will store positional arguments
// - "-" : Field will be ignored during parsing
// - "0", "1", ...: Field will store specific positional argument
//
// Example usage:
//
//	type Args struct {
//	    Host    string `opt:"host" alt:"H" def:"localhost" help:"server host address"`
//	    Port    int    `opt:"p" alt:"port" def:"8080" help:"server port number"`
//	    Debug   bool   `opt:"d" help:"enable debug mode"`
//	    Config  string `opt:"c" help:"path to config file"`
//	    Help    bool   `opt:"h" alt:"help" help:"show help information"`
//	    DocText string `opt:"?"`
//	}
//
//	func main() {
//	    var args Args
//	    if err := opt.Unmarshal(&args); err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
// Command line examples:
//
//	./app --host=localhost -p 8080 --debug
//	./app -H 127.0.0.1 --port=8080 -d
//	./app -h (displays help)
//
// Error handling:
// - Returns errors for unknown flags
// - Returns errors for invalid value types
// - Returns errors for array/slice overflow
// - Generates panic for invalid struct configuration
//
// Thread safety:
// The package is safe to use from multiple goroutines
// after initial configuration.
package opt
