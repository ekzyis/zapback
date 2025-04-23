package types

import (
	"fmt"
	"strings"
)

type FormError map[string]string

func (f FormError) Error() string {
	s := "FormError: "
	for k, v := range f {
		s += fmt.Sprintf("%s=\"%s\", ", k, v)
	}
	return strings.TrimRight(s, ", ")
}
