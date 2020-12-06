package opt

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// TestGetOptions tests getOptions function.
func TestGetOptions(t *testing.T) {
	var (
		exp     = "0:/main;1:a;2:b;3:c;d:true;h:localhost;lang:uk,en;port:8080"
		sopt    = "adh" // short option's names
		correct = [][]string{
			{
				"/main", "-dhlocalhost", "--port=8080",
				"--lang=uk,en", "a", "b", "c",
			},
			{
				"/main", "a", "-d", "-hlocalhost",
				"--lang uk,en", "--port 8080", "b", "c",
			},
			{
				"/main", "a", "--lang uk,en", "--port=8080",
				"-dh localhost", "b", "c",
			},
			{
				"/main", "--port=8080", "--lang=uk,en",
				"-h localhost", "-d", "--", "a", "b", "c",
			},
			{
				"/main", "a", "b", "--port=8080", "-h localhost",
				"-d", "--lang=uk,en", "c",
			},
			{
				"/main", "-h localhost", "-d", "--port 8080",
				"--lang=uk,en", "a", "b", "c",
			},
		}
		incorrect = [][]string{
			// teh -d is bool arg, but take `a` as value
			{
				"/main", "--port=8080", "--lang=uk,en",
				"-h localhost", "-d", "a", "b", "c",
			},
			// short option should not to use `=` for comparison
			{
				"/main", "--port=8080", "--lang=uk,en",
				"-h=localhost", "-d", "--", "a", "b", "c",
			},
		}
	)

	// Convert mpt[string]string into sorted arguments list.
	hash := func(v map[string]string) string {
		var tmp = make([]string, 0, len(v))
		for key, value := range v {
			tmp = append(tmp, fmt.Sprintf("%s:%s", key, value))
		}
		sort.Strings(tmp)
		return strings.Join(tmp, ";")
	}

	// Make correct tests.
	for _, test := range correct {
		v, err := getOptions(test, sopt)
		if err != nil {
			t.Error(err.Error())
		}

		if hash(v) != exp {
			t.Errorf("for %v incorrect value: %s != %s", test, exp, hash(v))
		}
	}

	// Make wrong tests.
	for _, test := range incorrect {
		v, err := getOptions(test, sopt)
		if err != nil {
			t.Error(err.Error())
		}

		if hash(v) == exp {
			t.Errorf("must be wrong %v", test)
		}
	}
}
