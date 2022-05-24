package opt

import (
	"net/url"
	"testing"
)

// TestGetFieldCastList tests getFieldCastList function.
func TestGetFieldCastList(t *testing.T) {
	var obj = struct {
		Pos0    string   `opt:"0" help:"app name"`
		User    string   `opt:"U" alt:"user" help:"user nickname"`
		Age     string   `opt:"age" def:"17" help:"age of user"`
		PosArgs []string `opt:"[]" help:"positional args"`
		Doc     string   `opt:"?" help:"positional args"`
	}{}

	if _, err := getFieldCastList(&obj); err != nil {
		t.Error(err)
	}
}

// TestGetFieldCastListExceptions tests getFieldCastList function
// for any exceptions.
func TestGetFieldCastListExceptions(t *testing.T) {
	var obj, sub = struct{}{}, 3

	if _, err := getFieldCastList(nil); err == nil {
		t.Error("there must be an error for nil value")
	}

	if _, err := getFieldCastList(obj); err == nil {
		t.Error("there must be an error for the non-pointer object")
	}

	if _, err := getFieldCastList(sub); err == nil {
		t.Error("there must be an error for the non-object")
	}

	if _, err := getFieldCastList(&obj); err != nil {
		t.Error(err)
	}
}

// TestGetFieldCastListHelpField tests getFieldCastList function
// for help field.
func TestGetFieldCastListHelpField(t *testing.T) {
	var (
		objWrong = struct {
			Help int `opt:"?"` // should be a string
		}{}
		objCorrect = struct {
			Help string `opt:"?"`
		}{}
	)

	if _, err := getFieldCastList(&objWrong); err == nil {
		t.Error("there must be an error for non-string Help field")
	}

	if _, err := getFieldCastList(&objCorrect); err != nil {
		t.Error(err)
	}
}

// TestGetFieldCastListPosArgs tests getFieldCastList function
// for positional arguments.
func TestGetFieldCastListPosArgs(t *testing.T) {
	var (
		objWrong = struct {
			PosArgs string `opt:"[]"` // should be a slice or an array
		}{}
		objCorrect = struct {
			PosArgs []string `opt:"[]"`
		}{}
	)

	if _, err := getFieldCastList(&objWrong); err == nil {
		t.Error("there must be an error for non-list PosArgs field")
	}

	if _, err := getFieldCastList(&objCorrect); err != nil {
		t.Error(err)
	}
}

// TestGetFieldCastListStruct tests getFieldCastList function
// for field as struct.
func TestGetFieldCastListStruct(t *testing.T) {
	var (
		objWrong = struct {
			Object struct{} // doesn't support nested structures
		}{}
		objCorrect = struct {
			Object url.URL // supports url.URL only
		}{}
	)

	if _, err := getFieldCastList(&objWrong); err == nil {
		t.Error("there must be an error for struct Object field")
	}

	if _, err := getFieldCastList(&objCorrect); err != nil {
		t.Error(err)
	}
}

// TestGetFieldCastListPtrStruct tests getFieldCastList function
// for field as a pointer on struct.
func TestGetFieldCastListPtrStruct(t *testing.T) {
	var (
		objWrong = struct {
			Object *struct{} // doesn't support nested structures
		}{}
		objCorrect = struct {
			Object *url.URL // supports url.URL only
		}{}
	)

	if _, err := getFieldCastList(&objWrong); err == nil {
		t.Error("there must be an error for *struct Object field")
	}

	if _, err := getFieldCastList(&objCorrect); err != nil {
		t.Error(err)
	}
}

// TestGetFieldCastListWrongTags tests getFieldCastList function
// for field with wrong tags.
func TestGetFieldCastListWrongTags(t *testing.T) {
	var (
		obj = struct {
			Ignored bool   `opt:"-"`
			User    string `opt:"user" alt:"***"` // incorrect *** tag
		}{}
	)

	if _, err := getFieldCastList(&obj); err == nil {
		t.Error("there must be an error for incorrect tag")
	}
}

// TestGetTagGroup tests getTagGroup function.
func TestGetTagGroup(t *testing.T) {
	// Ignored.
	tg, err := getTagGroup("User", "-", "U", "Goloop", "", "")
	if err != nil {
		t.Error(err)
	}

	if !tg.isIgnored {
		t.Error("must be an ignored field")
	}

	// Default opt.
	tg, err = getTagGroup("User", "", "U", "Goloop", "", "")
	if err != nil {
		t.Error(err)
	}

	if tg.longFlag != "user" {
		t.Errorf("expected user but %s", tg.longFlag)
	}

	// Folded field name.
	tg, err = getTagGroup("DatabaseUserName", "", "U", "Goloop", "", "")
	if err != nil {
		t.Error(err)
	}

	if tg.longFlag != "database-user-name" {
		t.Errorf("expected database-user-name but %s", tg.longFlag)
	}

	// Option one: opt is long, alt is short.
	tg, err = getTagGroup("User", "user", "U", "Goloop", ",", "")
	if err != nil {
		t.Error(err)
	}

	if tg.longFlag != "user" {
		t.Errorf("expected user but %s", tg.longFlag)
	}

	if tg.shortFlag != "U" {
		t.Errorf("expected U but %s", tg.shortFlag)
	}

	if tg.sepList != "," {
		t.Errorf("expected , but %s", tg.sepList)
	}

	// Option two: opt is short, alt is long.
	tg, err = getTagGroup("User", "U", "user", "Goloop", "", "")
	if err != nil {
		t.Error(err)
	}

	if tg.longFlag != "user" {
		t.Errorf("expected user but %s", tg.longFlag)
	}

	if tg.shortFlag != "U" {
		t.Errorf("expected U but %s", tg.shortFlag)
	}
}

// TestGetTagGroupWrongTag tests getTagGroup function
// for field with wrong tags.
func TestGetTagGroupWrongTag(t *testing.T) {
	// Incorrect name.
	if _, err := getTagGroup("User", "***", "U",
		"Goloop", "", ""); err == nil {
		t.Error("there must be an error for incorrect tag")
	}

	if _, err := getTagGroup("User", "", "***",
		"Goloop", "", ""); err == nil {
		t.Error("there must be an error for incorrect tag")
	}

	// Two long or two short tags.
	if _, err := getTagGroup("User", "user", "user",
		"Goloop", "", ""); err == nil {
		t.Error("there must be an error for incorrect tag")
	}

	if _, err := getTagGroup("User", "U", "U",
		"Goloop", "", ""); err == nil {
		t.Error("there must be an error for incorrect tag")
	}

	// Correct.
	if _, err := getTagGroup("User", "U", "user",
		"Goloop", "", ""); err != nil {
		t.Error(err)
	}

	if _, err := getTagGroup("User", "user", "U",
		"Goloop", "", ""); err != nil {
		t.Error(err)
	}

	if _, err := getTagGroup("User", "U", "",
		"Goloop", "", ""); err != nil {
		t.Error(err)
	}

	if _, err := getTagGroup("User", "user", "",
		"Goloop", "", ""); err != nil {
		t.Error(err)
	}
}
