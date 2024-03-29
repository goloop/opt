[![Go Report Card](https://goreportcard.com/badge/github.com/goloop/opt)](https://goreportcard.com/report/github.com/goloop/opt) [![License](https://img.shields.io/badge/license-MIT-brightgreen)](https://github.com/goloop/opt/blob/master/LICENSE) [![License](https://img.shields.io/badge/godoc-YES-green)](https://godoc.org/github.com/goloop/opt) [![Stay with Ukraine](https://img.shields.io/static/v1?label=Stay%20with&message=Ukraine%20♥&color=ffD700&labelColor=0057B8&style=flat)](https://u24.gov.ua/)

# opt

Package opt implements methods for manage arguments of the command-line.

## Installation

To install this package we can use `go get`:

```
$ go get -u github.com/goloop/opt
```

## Usage

To use this package import it as:

    import "github.com/goloop/opt"

## Quick start

The module supports parsing of different types of data, positional arguments, arguments passed on short and long flags. It can be any combination of command line arguments. For examples:

```shell
./app -H 0.0.0.0 --port=80 --no-verbose -U goloop.one -dc a.yaml,b.yaml 5 10 15
./app 5 --host 0.0.0.0 -p 80 --verbose false -U goloop.one -c a.yaml -c b.yaml -d -- 10 15
./app 5 10 -c a.yaml -p 80 --host=0.0.0.0 --no-verbose -dU goloop.one -c b.yaml 15
./app 5 10 15 -c a.yaml,b.yaml -p 80 --host=0.0.0.0 --no-verbose -dU goloop.one
./app 5 10 15 -c a.yaml,b.yaml -p 80 -H 0.0.0.0 --verbose=false -d true -U goloop.one
./app -c a.yaml,b.yaml -p 80 -H 0.0.0.0 --no-verbose -dU goloop.one 5 10 15
./app 5 --verbose=false -c a.yaml -c b.yaml -p 80 -H 0.0.0.0 -dU goloop.one 10 15
```

Example of use:

```go
package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/goloop/opt"
)

// Args is command-line argument object.
type Args struct {
	Host string `opt:"H" alt:"host" def:"localhost" help:"host of the server"`
	Port int    `opt:"port" alt:"p" def:"8080" help:"port of the server"`
	Help bool   `opt:"h" alt:"help" help:"show application usage information"`

	Debug   bool     `opt:"d" help:"debug mode"`
	Verbose bool     `def:"true" help:"enable verbose mode"` // --verbose
	Configs []string `opt:"c" sep:","`

	ServerURL url.URL `opt:"U" help:"URL to the server"`

	Doc        string `opt:"?"`
	Positional []int  `opt:"[]"`
}

func main() {
	var args Args

	if err := opt.Unmarshal(&args); err != nil {
		log.Fatal(err)
	}

	if args.Help {
		fmt.Println(args.Doc)
		os.Exit(0)
	}

	fmt.Println("Host:", args.Host)
	fmt.Println("Port:", args.Port)
	fmt.Println("Help:", args.Help)
	fmt.Println("Debug:", args.Debug)
	fmt.Println("Verbose:", args.Verbose)
	fmt.Println("Configs:", args.Configs)
	fmt.Println("Server URL:", args.ServerURL.String())
	fmt.Println("Positional:", args.Positional)

	// Output:
	//  Host: 0.0.0.0
	//  Port: 80
	//  Help: false
	//  Debug: true
	//  Verbose: false
	//  Configs: [a.yaml b.yaml]
	//  Server URL: goloop.one
	//  Positional: [5 10 15]
}

```

Show help on using command line arguments.

```shell
./app -h
```

Result:

```shell
Options:
    -H, --host    host of the server;
    -d            debug mode;
    -h, --help    show application usage information;
    -p, --port    port of the server;
        --verbose enable verbose mode;
    -U            URL to the server.

Positional arguments:
The app takes an unlimited number of positional arguments.
```

## Rules of the parsing

### Command-line arguments

Command-line arguments are a way to provide the additional parameters to the GoLang application in the process of launching. GoLang has an os package that contains a slice called as Args, this one is an array of string that contains all the command line arguments passed.

For example: `./app -Vp80 --host=127.0.0.1 --user goloop` which will be presented in os.Args as: `[]string{"./app", "-Vp80", "--host=127.0.0.1", "--user", "goloop"}`.

There are various discussions on how to name certain items in command-line: flags, values, options, parameters, switches, etc. In this package we use the following terminology: command-line arguments can be divided into two categories - flags and values.

Flags are instructions that allow to change the behavior of the program. Flags can be divided into long and short, with values or as switches.

Values are data that will be set as additional program parameters. Values can be passed as position arguments or through a flag (for flags with values).

Thus, in the example above: `./app` is a positional value (program name, this argument is automatically added by GoLang to the argument list); `-Vp80` is a group of two short flags `-V` and `-p` where the latter has a value of `80`; `--host=127.0.0.1` is a long flag `--host` that has a value of `127.0.0.1` etc.

### Long flags

A long flag requires two dashes before the argument name, for example: `--host`, `--debug`, `--verbose`. The name of the flag must consist of latin letters and numbers. Also to separate the words of the argument name can be used dash, for example: `--user-name`.

The long flag name is not case sensitive, ie. `--host`, `--HOST` and `--Host` it is the same names.

After prefix from two dashes there should be no spaces, for example `--host` is correct flag but `-- host` is incorrect flag identifier (two dashes `--` are used to separate a group of position values from a group of flags, see below).

If flag has value it must be written after the equal sign `=` or after the space, for example: `--host=localhost` and `--host localhost` the same. Value from a few words separated by a space should be enclosed in quotation marks, for example: `--user="Smith J."` or `--user "Smith J."`.

The flag can be used as a sweater, in this case, it is assumed that its absence in the command line this flag contains a false value (but it is not required, and depends on the default value specified in the program). The switch can be prefixed `no-`, which sets the value to false.

For example, for `--verbose` flag: `--verbose true` and `--verbose` the same; `--verbose false` and `--no-verbose` the same too.

The sequence of flags does not create problems, so the following arguments will give the same result:

```
./app -dUJack -U Bob --user=Roy --no-verbose -g"L R" 5 10 15
./app -dU Jack -U Bob --USER=Roy --No-Verbose -g "L R" 5 10 15
./app -d -UJack -U Bob -URoy --no-verbose -g"L R" 5 10 15
./app -g "L R" -U Jack -UBob -URoy --no-verbose -d -- 5 10 15
./app 5 -dU Jack --user Bob -URoy --verbose false -g"L R" 10 15
./app 5 10 --no-verboSe -dU Jack -U Bob -U Roy -g "L R" 15
./app 5 10 15 -UJack -UBob -URoy --verbose=false -g "L R" -d
./app 5 10 15 -U Jack -UBob -URoy -d --VERBOSE false -g"L R"
./app 5 -UJack --no-verbose -g"L R" -UBob -d true -URoy 10 15
./app 5 -UJack -UBob --verbose false -URoy -g"L R" -d -- 10 15
```

### Short flags

A short flag requires one dash before the argument name, for example: `-h`, `-d`, `-v`. The name of the flag must consist of latin letters only.

The flag name is case sensitive, ie. `-h` and `-H` are different flags.

After prefix from one dash there should be no spaces, for example `-h` is correct flag but `- h` is incorrect flag identifier.

If flag has value it must be written after the space or without separating the value from the flag, for example: `-p 8080` and `-p8080` the same. Value from a few words separated by a space should be enclosed in quotation marks, for example: `-u "Smith J."` or `-u"Smith J."`.

The flag can be used as a sweater, in this case, it is assumed that its absence in the command line this flag contains a false value.

For example, for `-v` flag: `-v true` and `-v` the same; `-v false` to set the value to false.

Short flags can be grouped. For example for flags `-v`,` -d`, `-p` where the latter has the value` 8080` can be written as `-vdp8080` or` -dvp 8080` or `-vd -p 8080` or` -vp8080 - etc.

Pay attention to duplication of short flags in one group, for example: `-vpv` or `-vvp`, where second `v` is duplicated. In this case, its second iteration will be considered as value for the previous flag, ie `-vpv` equivalent `-v -p v` where `-p` gets the value `v`, and `-vvp` equivalent for `-v vp` where `-v` gets the value vp.

### Positional arguments

Positional arguments are arguments that do not relate to the value of a flag.

Positional arguments start on the left and follow the first short or long flag. For example: `./app 5 10 15 -p8080`, where `./app`, `5`, `10`, `15` is a positional arguments.

Zero positional argument is the full path to the program being launched. This value is set automatically.

Also, position arguments are written after the value to the last short or long flag. For example: `./app --host localhost 5 10 15` or `./app -p 8080 5 10 15` etc., where `localhost` it is value for `--host` flag and `8080` for `-p` flag, but `./app`, `5`, `10`, `15` is a positional arguments.

If position arguments are written to the right and the last flag on the command line is a switcher (doesn't contain value, for example the boolean flag of debug), positional arguments must be written after a double dash. For example: `./app --host localhost --debug -- 5 10 15`. If you omit the double dash, argument `5` will be passed as a value to the `--debug` flag.

Thus, it is not the `-f` flag that is passed, but the positional argument (value) `-f`.

Positional arguments can be written both left and right simultaneously. For example: `./app 5 10 --host localhost -d -- 15` and `./app 5 10 -d --host localhost 15` and `./app 5 --host localhost -d -- 10 15` etc,. the same.


### Duplicate flags

In the command line, flags that are processed as lists (slice/array) can be duplicated. For example, we need to pass a list of users, for which there is a short flag `-u` and a long flag` --user`, and in the program the list has the type `[] string`: `./app -uJohn --user=Bob -u Roy` will give a result as `[]string{"John", "Bob", "Roy"}`.

Duplicate flags that are not declared in the program as a list (slice/array) don't cause an error. In this situation the value for the item will be taken from the last entry in the list.

### Flags that are not declared

If the flag is not declared in the program, but it is in the command line - this should cause an error, or display help information with available commands.

## Tag structure

You can use the following tags to configure command line parsing rules:

- opt - short or long flag name;
- alt - alternate flag name of opt value;
- def - default field value;
- spe - if the field is a list, indicates the delimiter of the list;
- help - short description of the option.

### Tag `opt`

Specifies the name of the short or long flag whose data must be entered in the appropriate field. If no tag is specified, it will be automatically set to the value of the field name converted to kebab-case. For example: `UserName` converts to `user-name`.

Has reserved values:

- `-` - field to ignore;
- `?` - field to save the generated help information;
- `[]` - field to save the positional arguments;
- `0`, `1`, ..., `N` where N is digit - the specific value of the position argument of the specified index (for index 0 the value is reserved - the full path of the application call).

For example: `./app --host localhost --user-name Goloop 1 2 3`

```go
var args = struct {
	Host       string `opt:"host"`
	UserName   string // automatically `opt:"user-name"`
	Ignored    int    `opt:"-"`
	Positional []int  `opt:"[]"`
}{}

if err := opt.Unmarshal(&args); err != nil {
	log.Fatal(err)
}

fmt.Println("Host:", args.Host)
fmt.Println("UserName:", args.UserName)
fmt.Println("Ignored:", args.Ignored)
fmt.Println("Positional:", args.Positional)

// Output:
//  Host: localhost
//  UserName: Goloop
//  Ignored: 0
//  Positional: [1 2 3]
```

### Tag `alt`

Sets the alternate flag name of opt value. For example, if opt has value of the long flag name, then alt can take the value of the short ensign and vice versa. The values of opt and alt cannot be either a long or a short flag name at the same time.

```go
var args = struct {
	Host string `opt:"host" alt:"h"` // -h, --host
	Port int    `opt:"p" alt:"port"` // -p, --port
	// SimTwoShort string `opt:"s" alt:"t"`     // panic
	// SimTwoLong  string `opt:"sim" alt:"two"` // panic
}{}

if err := opt.Unmarshal(&args); err != nil {
	log.Fatal(err)
}

fmt.Println("Host:", args.Host)
fmt.Println("Port:", args.Port)

// Output:
//  Host: localhost
//  Port: 80
```

Can handle the following command line arguments:

```shell
./app -h localhost -p 80
./app --host localhost -p 80
./app -h localhost --port 80
./app --host=localhost --port=80
```

### Tag `def`

Sets the default value of the field. To set a bool value are used `true` or `false`. To set a list value are used elements of the corresponding type separated by some separator are used. The delimiter type must be specified in the sep tag. For example: `./app`

```go
var args = struct {
	Host     string   `opt:"h" def:"localhost"`
	Port     int      `opt:"p" def:"8080"`
	Verbose  bool     `opt:"v" def:"true"`
	UserList []string `opt:"U" def:"John,Bob,Roy" sep:","`
	AgeList  []int    `opt:"A" def:"23/25/27" sep:"/"`
}{}

if err := opt.Unmarshal(&args); err != nil {
	log.Fatal(err)
}

fmt.Println("Host:", args.Host)
fmt.Println("Port:", args.Port)
fmt.Println("Verbose:", args.Verbose)
fmt.Println("UserList:", args.UserList)
fmt.Println("AgeList:", args.AgeList)

// Output:
//  Host: localhost
//  Port: 8080
//  Verbose: true
//  UserList: [John Bob Roy]
//  AgeList: [10 20 30]
```

To pass the list, you can use the flag for each new argument or write them through the separator specified in the sep tag.

```sh
./app -h localhost -UJohn -UBob,Roy
./app -p8080 -A23 -A25 -A27
./app -p 8080 -A23/25/27 -UJohn,Bob,Roy
```

### Tag `sep`

Specifies the symbol to divide the list into items. Relevant in list type fields only. By default is empty - forbids passing the list as one value (ie, you need to use a flag for each item, for example: `-A23 -A20 -A30` but it is impossible somehow so: `-A23,25,27`).

Specifies the symbol to divide the list into items. Relevant in list type fields only. By default is empty - forbids passing the list as one value (ie, you need to use a flag for each item, for example: `-A23 -A20 -A30` but it is impossible somehow so: `-A23,25,27`). If a value is specified, this value is used to distribute the list in both the def thesis and the argument command line. For example: `./app -a23,25:27 -b23,25:27 -c23,25:27`.


```go
var args = struct {
	ListA []string `opt:"a" sep:""`
	ListB []string `opt:"b" sep:","`
	ListC []string `opt:"c" sep:":"`
}{}

if err := opt.Unmarshal(&args); err != nil {
	log.Fatal(err)
}

fmt.Println("ListA:", args.ListA, "len:", len(args.ListA))
fmt.Println("ListB:", args.ListB, "len:", len(args.ListB))
fmt.Println("ListC:", args.ListC, "len:", len(args.ListC))

// Output:
//  ListA: [23,25:27] len: 1
//  ListB: [23 25:27] len: 2
//  ListC: [23,25 27] len: 2
```

### Tag `help`

The tag is used to briefly describe the arguments of the command line. If the tag is empty - the argument isn't displayed in the auto-generated help information. For example: `./app -h`

```go
var args = struct {
	Host string `opt:"host" alt:"H" help:"host of the server"`
	Port int    `opt:"port" alt:"p" help:"port of the server"`
	Help bool   `opt:"h" help:"show help information"`

	Doc      string    `opt:"?"`
	FileName string    `opt:"1" help:"configuration file"`
}{}

if err := opt.Unmarshal(&args); err != nil {
	log.Fatal(err)
}

if args.Help {
	fmt.Println(args.Doc)
}

// Output:
//  Options:
//      -H, --host host of the server;
//      -h         show help information;
//      -p, --port port of the server.
//
//  Positional arguments:
//  The app takes an one of positional argument, including:
//       1 configuration file.
```

## Panic or error

The Unmarshal function can cause panic or return an error. Panic occurs only when there is a development problem. The error occurs when the user has transmitted incorrect data.

### Panic

Panic occurs when the structure contains incorrect parsing fields or other technical problems. For example:

- the object isn't a structure;
- the object isn't transmitted by pointer;
- a non-string type field is specified for the `opt:"?"` documentation field;
- field for positional arguments `opt:"[]"` is not a list (slice/array);
- field has structure type (except url.URL);
- field has pointer to structure type (except *url.URL).

```go
var args = struct {
	Doc int      `opt:"?"`  // panic: Doc field should be a string
	Pos string   `opt:"[]"` // panic: Pos field should be a list
	One struct{}            // panic: One field has invalid type
	Two struct{} `opt:"-"`  // it's normal, the field is ignored
}{}

// panic: obj should be a pointer to an initialized struct
if err := opt.Unmarshal(args); err != nil {
	log.Fatal(err)
}
```

### Error

Error occurs when it is impossible to parse the command line passed by the user. For example:

- a flag is used that is not specified in the argument object;
- the specified argument value does not match the field type;
- for array (overflow) specified too large list;
- value is too large or too small for numeric fields.

```go
var args = struct {
	Host string `opt:"host" alt:"H" help:"host of the server"`
	Port int    `opt:"port" alt:"p" help:"port of the server"`
	Help bool   `opt:"h" help:"show help information"`

	Doc      string `opt:"?"`
	FileName string `opt:"1" help:"configuration file"`
}{}

if err := opt.Unmarshal(&args); err != nil {
	log.Fatalf("%v\nRun ./app -h for help ", err)
}
```

- `./app --verbose` - error: invalid argument --verbose;
- `./app --port=hello` - error: 'hello' is incorrect value;
- `./app --h yes` - error: 'yes' has incorrect type, bool expected.

If an error occurs, the text of the help info will still be generated. Therefore, a parsing error can be accompanied by a display of this one:

```go
switch err := opt.Unmarshal(&args); {
case err != nil:
	log.Fatalf("%v\n\nError: %v", args.Doc, err)
case args.Help:
	fmt.Println(args.Doc)
	os.Exit(0)
}
```

## Usage

#### func  Unmarshal

    func Unmarshal(obj interface{}) error

Unmarshal parses the argument-line options and stores the result to go-struct.
If the obj isn't a pointer to struct or is nil - returns an error.

Unmarshal method supports the following field's types: int, int8, int16, int32,
int64, uin, uint8, uin16, uint32, in64, float32, float64, string, bool, url.URL
and pointers, array or slice from thous types (i.e. *int, ..., []int, ...,
[]bool, ..., [2]*url.URL, etc.).

For other filed's types (like chan, map ...) will be returned an error.

The function generates a panic if:

- the object isn't a structure; - the object isn't transmitted by pointer; - a
non-string type field is specified for the `opt:"?"` doc-field; - field for
positional arguments `opt:"[]"` is not a list (slice/array); - field has
structure type (except url.URL); - field has pointer to structure type (except
*url.URL).

Use the following tags in the fields of structure to set the marshing
parameters:

    opt  indicates a short or long option;
    alt  optional, alternative option for position opt,
         if a long option is specified in opt, a short option
         can be specified in alt or vice versa;
    def  default value (if empty, sets the default value
         for the field type of structure);
    help brief help about the option.

Suppose that the some values was set into argument-line as:

    ./main --host=0.0.0.0 -p8080

Structure example:

    // Args structure for containing values from the argument-line.
    type Args struct {
    	Host string `opt:"host" def:"localhost"`
    	Port int    `opt:"p" alt:"port" def:"80" help:"port number"`
    	Help bool   `opt:"h" alt:"help"`
    }

Unmarshal data from the argument-line into Args struct.

    var args Args
    if err := opt.Unmarshal(&args); err != nil {
    	log.Fatal(err)
    }

    fmt.Printf("Host: %s\nPort: %d\n", args.Host, args.Port)
    // Output:
    //  Host: 0.0.0.0
    //  Port: 8080

#### func  Version

    func Version() string

Version returns the version of the module.
