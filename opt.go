package opt

import "os"

// Unmarshal to parses the argument-line options and stores the result in
// the value pointed to by obj. If the obj isn't pointer to struct or is nil -
// returns an error.
//
// Unmarshal method supports the following field's types: int, int8, int16,
// int32, int64, uin, uint8, uin16, uint32, in64, float32, float64, string,
// bool, url.URL and pointers, array or slice from thous types (i.e. *int, ...,
// []int, ..., []bool, ..., [2]*url.URL, etc.).
// For other filed's types (like chan, map ...) will be returned an error.
//
// Structure fields may have a opt tag as `opt:"short[,long[,value[,help]]]"`
// where:
//
//    short - short name of the option (one char from the A-Za-z);
//    long - long name of the option (at least two characters);
//    value - default value;
//    help - help information.
//
// Suppose that the some values was set into argument-line as:
//
//    ./main --host=0.0.0.0 -p8080
//
// Structure example:
//
//    // Config structure for containing values from the argument-line.
//    type Config struct {
//        Host string `opt:",host,localhost"`
//        Port int    `opt:"p,port,80,port number"`
//        Help bool   `opt:"h,help"`
//    }
//
// Unmarshal data from the argument-line into Config struct.
//
//    // Important: pointer to initialized structure!
//    var config = &Config{}
//
//    err := opt.Unmarshal(config)
//    if err != nil {
//        // something went wrong
//    }
//
//    config.Host // "0.0.0.0"
//    config.Port // 8080
func Unmarshal(obj interface{}) error {
	return unmarshalOPT(obj, os.Args)
}
