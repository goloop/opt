package opt

import (
	"reflect"
	"strings"
	"testing"
)

// TestGetOptionPrefix tests getOptionPrefix function.
func TestGetOptionPrefix(t *testing.T) {
	tests := []struct {
		short  string
		long   string
		result string
		length int
	}{
		{"v", "verbose", "    -v, --verbose", 17},
		{"", "port", "        --port", 14},
		{"h", "", "    -h", 6},
		{"", "", "", 0},
	}

	for i, test := range tests {
		r, l := getOptionPrefix(test.short, test.long)
		if r != test.result {
			t.Errorf("%d test, expected `%s` but `%s`", i, test.result, r)
		}

		if l != test.length {
			t.Errorf("%d test, expected %d but %d", i, test.length, l)
		}
	}
}

// TestWrapHelpMsg tests wrapHelpMsg function.
func TestWrapHelpMsg(t *testing.T) {
	tests := []struct {
		tab    int
		wc     int
		prefix string
		help   string
		result []string
	}{
		{
			4,
			20,
			"- ",
			"one two three",
			[]string{"- one two three"},
		},
		{
			4,
			15,
			"- ",
			"one two three",
			[]string{"- one two", "three"},
		},
		{
			4,
			15,
			"- ",
			"",
			[]string{},
		},
	}

	for i, test := range tests {
		r := wrapHelpMsg(test.prefix, test.help, test.tab, test.wc)
		if !reflect.DeepEqual(r, test.result) {
			if len(r) != 0 && len(test.result) != 0 {
				t.Errorf("%d test, expected %v but %v", i, test.result, r)
			}
		}
	}
}

// TestGetOptionBlock tests getOptionBlock function.
func TestGetOptionBlock(t *testing.T) {
	var (
		slice = reflect.ValueOf([]string{})
		array = reflect.ValueOf([8]string{})
		tests = []struct {
			fcl        fieldCastList
			subOptText string
			posArgsLen int
		}{
			{
				fcl: fieldCastList{
					&fieldCast{
						fieldName: "Port",
						tagGroup: &tagGroup{
							shortFlag: "p",
							longFlag:  "port",
							helpMsg:   "port of serve",
						},
						item: nil,
					},
					&fieldCast{
						fieldName: "Host",
						tagGroup: &tagGroup{
							shortFlag: "h",
							longFlag:  "host",
							helpMsg:   "host on server",
						},
						item: nil,
					},
					&fieldCast{
						fieldName: "Verbose",
						tagGroup: &tagGroup{
							shortFlag: "v",
							longFlag:  "verbose",
							helpMsg:   "",
						},
						item: nil,
					},
				},
				subOptText: "port of serve.",
				posArgsLen: -1,
			},
			{
				fcl: fieldCastList{
					&fieldCast{
						tagGroup: &tagGroup{
							shortFlag: "[]",
						},
						item: &array,
					},
				},
				posArgsLen: 8, // the array's full length is its capacity
			},
			{
				fcl: fieldCastList{
					&fieldCast{
						tagGroup: &tagGroup{
							shortFlag: "[]",
						},
						item: &slice,
					},
				},
				posArgsLen: 0,
			},
			{
				fcl: fieldCastList{
					&fieldCast{
						fieldName: "Host",
						tagGroup: &tagGroup{
							shortFlag: "h",
							longFlag:  "host",
							helpMsg: "network host is a computer or other " +
								"device connected to a computer network. " +
								"A host may work as a server offering " +
								"information resources, services, and " +
								"applications to users or other hosts " +
								"on the network",
						},
						item: nil,
					},
				},
				subOptText: "-h, --host" + separator + "network host is",
				posArgsLen: -1,
			},
		}
	)

	for i, test := range tests {
		optText, posArgsLen := getOptionBlock(test.fcl)
		if strings.Index(optText, test.subOptText) < 0 {
			t.Errorf(
				"%d test, expected substring %s but it is not "+
					"found in the text:\n%s",
				i, test.subOptText, optText,
			)
		}

		if test.posArgsLen != posArgsLen {
			t.Errorf(
				"%d test, expected %d but %d",
				i, test.posArgsLen, posArgsLen,
			)
		}
	}
}

// TestGetPositionalBlock tests getPositionalBlock function.
func TestGetPositionalBlock(t *testing.T) {
	tests := []struct {
		fcl        fieldCastList
		count      int
		subPosText string
	}{
		{
			fcl: fieldCastList{
				&fieldCast{
					tagGroup: &tagGroup{
						shortFlag: "1",
						helpMsg:   "port of serve",
					},
				},
				&fieldCast{
					tagGroup: &tagGroup{
						shortFlag: "2",
						helpMsg:   "host of serve",
					},
				},
			},
			count:      1,
			subPosText: "1 port of serve",
		},
		{
			fcl: fieldCastList{
				&fieldCast{
					tagGroup: &tagGroup{
						shortFlag: "1",
						helpMsg:   "port of serve",
					},
				},
				&fieldCast{
					tagGroup: &tagGroup{
						shortFlag: "2",
						helpMsg: "network host is a computer or other " +
							"device connected to a computer network. " +
							"A host may work as a server offering " +
							"information resources, services, and " +
							"applications to users or other hosts " +
							"on the network",
					},
				},
			},
			count:      0,
			subPosText: "2" + separator + "network host is a computer",
		},
	}

	for i, test := range tests {
		posText := getPositionalBlock(test.fcl, test.count)
		if strings.Index(posText, test.subPosText) < 0 {
			t.Errorf(
				"%d test, expected substring %s but it is not "+
					"found in the text:\n%s",
				i, test.subPosText, posText,
			)
		}
	}
}
