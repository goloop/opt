package opt

import (
	"reflect"
	"testing"
)

// TestToolsSplit tests the split function.
func TestToolsSplit(t *testing.T) {
	var (
		test     = "a:a b c:b:c"
		expected = []string{"a", "a b c", "b", "c"}
	)

	if v := split(test); !reflect.DeepEqual(v, expected) {
		t.Errorf("expected %v but %v", expected, v)
	}
}
