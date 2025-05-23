package structs

import (
	"github.com/wind959/ko-utils/validator"
	"strings"
)

type Tag struct {
	Name    string
	Options []string
}

func newTag(tag string) *Tag {
	res := strings.Split(tag, ",")
	return &Tag{
		Name:    res[0],
		Options: res[1:],
	}
}

// HasOption check if a struct field tag has option setting.
func (t *Tag) HasOption(opt string) bool {
	for _, o := range t.Options {
		if o == opt {
			return true
		}
	}
	return false
}

// IsEmpty check if a struct field has tag setting.
func (t *Tag) IsEmpty() bool {
	return validator.IsEmptyString(t.Name)
}
