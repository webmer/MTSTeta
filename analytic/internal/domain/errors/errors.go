package errors

import (
	"fmt"
)

type JsonErrWrapper struct {
	E string
}

func (j JsonErrWrapper) Error() string {
	return fmt.Sprintf("{\"error\": \"%s\"}", j.E)
}
