/*
# opt

The opt module implements methods for manage arguments of the command-line.

## Installation

To install this package we can use `go get`:

    $ go get -u github.com/goloop/opt

## Usage

To use this package import it as:

    import "github.com/goloop/opt"


## Quick start

Command-line options are stored in a Go-struct, like this, for example:

    // Args is structure for parsing os.Args data.
    type Args struct {
        Help     bool     `opt:"h,help,,show this help"`
        Verbose  bool     `opt:"v,verbose,true,verbose mode (default true)"`
        Debug    bool     `opt:"d,debug,,debug mode"`
        Users    []string `opt:"U,users,{John,Bob,Robert},user list"`
        Greeting string   `opt:"g,,'Howdy!',greeting message"`

        // Special variable that contains whole documentation string has
        // short option name `?`. This variable will be filled automatically
        // after parsing the command-line.
        Doc string `opt:"?"`

        // To save the path to the app use short option name as `0` (zero).
        Path string `opt:"0,,,path to app"` // short name as: '0'

        // Positional arguments can be stored in a slice or array.
        // Note: use either an array or a slice but not both at once
        // at the same time. Set short option name as `[]`.
        // ABC [3]int `opt:"[]"` // short name as: '[]'
        ABC []int `opt:"[]"` // short name as: '[]'

        // Distribute positional arguments to fields, use short
        // option name as: 1,2.. N.
        A int `opt:"1"` // 1st positional argument
        B int `opt:"2"` // 2nd positional argument
        C int `opt:"3"` // 3rd positional argument
    }

    // HelpOPT adds general information to the automatic documentation data.
    func (a Args) HelpOPT() string {
        return "Test application for testing opt package features."
    }

To import opt package use: `import "github.com/goloop/opt"`

Use the Unmarshal method to parse os.Args and save it intp Go-struct, like:

    var args = Args{}
    err := opt.Unmarshal(&args) // pass a pointer to the struct
    if err != nil {
        log.Fatal(err)
    }


Now the application can be launched by any of the following methods (all call
options are equivalent for `opt` parser):

    ./app -dU Jack,Harry --no-verbose -g "Hello, world" 5 10 15
    ./app -d -U Jack,Harry --no-verbose -g "Hello, world" -- 5 10 15
    ./app -dU Jack,Harry --verbose false -g "Hello, world" 5 10 15
    ./app --no-verbose true -dU Jack,Harry -g "Hello, world" 5 10 15
    ./app -dUJack,Harry --verbose=false -g"Hello, world" 5 10 15
    ./app 5 10 15 -dUJack,Harry --verbose false -g"Hello, world"
    ./app 5 10 15 -UJack,Harry --verbose false -g"Hello, world" -d true
    ./app 5 -UJack,Harry --verbose false -g"Hello, world" -d -- 10 15

For this example, after parsing, the data will be as follows:

    // args.Help == false
    // args.Verbose == false
    // args.Debug == true
    // args.Users == []string{"Jack", "Harry"}
    // args.Greeting) == "Hello, world"

    // args.Path == "./main"

    // args.ABC == []int{5, 10, 15}

    // args.A == 5
    // args.B == 10
    // args.C == 15

The opt package provides the ability to auto-generate help. In this example,
the `Doc` field is a container for storing auto-generate help data:


    // Show help information.
    if args.Help {
        fmt.Println(args.Doc)
        return
    }


After calling the application as `./main -h`, help information
will be displayed:

    Test application for testing opt package features.

    Usage: ./main [-h] [-v] [-d] [-U value] [-g value] -- a1, ..., a3
        -h,--help      show this help
        -v,--verbose   verbose mode (default true)
        -d,--debug     debug mode
        -U,--users     user list
        -g             greeting message
        a1, ..., a3    positional arguments

The struct that implements the `opt.Helper` interface that allows you to add
arbitrary information to automatically generated help data.

## Flag rules

### Named options

The opt is flag identifier and the flag has a structure
as `[short[,long[,value[,help]]]]` where is:

    short  short option name, like: `-p`, `-h`, `-d`, `-U`;
    long   long option name, like: `--port`, `--help`, `--debug`;
    value  default value;
    help   argument summary.

At least one option identifier must be specified, short or long or both at once.

Arguments may be skipped or set to underscore as: `,port,8080` equal
to `_,port,8080,_` or `p,,,port number` equal to `p,_,_,port number`.

Examples

    type Args struct {
        Help bool   `opt:"h,help,,show help"` // default 'false' for bool type
        Host string `opt:"_,host,localhost,host addr"`
        Port int    `opt:"p,port,8080,port number"`
    }


#### Short name

Only one character long in the range `A-Za-z`.

#### Long name

From two or more characters in a range `a-z`

#### Value

Literal of the appropriate type. To specify a complex default value like
string with `,` symbol, or sequence of values - escape the expression
accordingly:

    OpenPorts []int  `opt:",op,{80,8080,8383},list of open ports"`
    Greeting  string `opt:"g,,'Hello, world',greeting message"`
    Parting   string `opt:"p,,\"Oh, bye\",greeting message"`

#### Help

Help line for argument. The string length is not limited. If the length of
the help line and the left part (option name) are longer than 79 characters,
the help will be cut into small pieces and formatted correctly.

// Args structure for parsing os.Args using opt package.
    type Args struct {
        // ...
        Verbose  bool `opt:"v,verbose,true,verbose mode (default true) you can use --no-verbose or --verbose=false to deactivate it."`
        // ...
    }


Format result

    ...
    -v,--verbose   verbose mode (default true) you can use --no-verbose or
                   --verbose=false to deactivate it.
    ...

### Positional options and path to the executable

To set the positional option needed to specify an order index of it as
a short option name. The long option name is ignored. Zero index always
indicates  to the executable path.

    type Args struct {
        Help bool   `opt:"h,help,,show help"`
        Path string `opt:"0"` // executable path
        Host string `opt:"1,,,host addr"`
        Port int    `opt:"2,,80,port number"`
    }

It looks like: `./main localhost 8080` after `Host` contains `localhost` and
`Port` - `8080`. Named positional arguments are not required,
so `./main localhost` is correct construction (port is default value: 80).

An array or slice can be used to indicate a list
of a specific/indefinite length of the same type of positional arguments.
The list contains positional arguments from 1 index (i.e. ignoring
executable path). As short option name the `[]` marker must be set.

    type Args struct {
        Path   string  `opt:"0"` // executable path
        Values []int   `opt:"[]"`
    }

It looks like: `./main 1 2 3 4 5 6 7`  after `Values` contains
`[]int{1,2,3,4,5,6,7}`.  Positional arguments are optional.

When using an array as container for positional options, there
may be an overflow error:

    type Args struct {
        Values [3]string `opt:"[]"`
    }

And `./main one two three four` make an overflow error.

### Help container

Documentation is automatically generated during unmarshalling. In order to
designate the field for help data set `?` as the short name. The field must
be of string type. The remaining positions (long, value, help) will be
ignored by the parser:

    type Args struct {
        HelpData string `opt:"?"`
    }

## Rules of the parsing

### Long option

It self-documenting option like: `--host`, `--user-name`, `--debug`.
There are two dashes in front of the name and the name can use one or more
dashes to separate words. This option case not sensitive,
so `--user-name`, `--User-Name`, `--USER-NAME` these are the same names.

After prefix from two dashes  there should be no spaces:

    `--debug`  correct;
    `-- debug` incorrect for long option, it's like positional argument.

If option has value it must be written after the equal sign (`=`) or after
the space, like: `--host=localhost` or  `--host localhost`.
Multi-word meaning must be enclosed in double quotes,
like: `--user-name="John Smith"` or `--user-name "John Smith"`.
The switcher can be specified without value: `--debug` or `--debug=true`
or `--debug true` for args configurations:

    type Args struct {
        Debug bool `opt:",debug"` // default false for bool type
    }

For boolean values in a long option the prefix `no-` can be specified
to indicate `false`: `--debug` equal to `--debug=true`,
`--no-debug` equal to `--debug=false`. This is also valid
for specifying the default value as:

    type Args struct {
        Debug bool `opt:",debug,true"` // default value sets as true
    }

### Short option

The abbreviation for long option is indicated as one character having
one dash character in front of the name, like:  `-h`, `-U`, `-d`.
This option case sensitive, so `-u` and `-U` it's different names.
Latin characters only: `A-Za-z`.

After prefix from one dashes there should be no spaces:

    `-d` - correct;
    `- d` - incorrect.

If option has value it must be written after the space character,
like: `-h localhost`. Multi-word meaning must be enclosed in double
quotes, like: `-U "John Smith"`.

The values as numbers can be written without a space,
i.e `-p8080` equal to `-p 8080`.

The string value can also be written without a space if the first
character of the value not expected short option:

    type Args struct {
        Host  string `opt:",host,0.0.0.0"`
        Port  int    `opt:"p,port,8080"`
        Debug int    `opt:"d"`
    }

-- short options registered: `-p` and `-d`, but options like `-t` and`-f`
isn't registered therefore an expression like: `-dtrue` equal to `-d true`,
or `-dfalse` equal to `-d false`. But option like: `-dpass` will be parsed
as `-d -p ass` because there is a registered `-p` option.

Short options can be grouped: `-d -i -Z` equal to `-diZ`. In the group, the
value can be specified only for the last option: `-diZp8080` or `-diZp 8080`
equal to `-diZ -p 8080`.

### Positional arguments

Zero positional argument is the full path to the program being launched.
This value is set automatically.

Starting with the first argument without prefixes `--` and `-` the positional
arguments are specified are divided with space or can be set at the end of a
line after a double blank `--` values: `positional arguments
[options[[--] positional arguments]]`. For examples:

Double dashes can be ommited if positional argument has no
prefix `--` or `-` and positional argument not located immediately
after a boolean option with value by default.

The `one two "three and four" -p8080 --debug` or `-p 8080 --debug -- one
two "three and four"` or `one -p8080 --debug=true two "three and
four"` equal to Go-map: `map[string]string{"1": "one", "2": "two",
"3": "three and four", "-p": "8080", "--debug": "true"}`.

For example

    type Args struct {
        Debug int `opt:"d,,,debug"` // boolean option
        Pos []int `opt:"[]"`
    }
*/
package opt
