package custom_errors

import "fmt"

type NotFoundErr struct {
	Message string
}

func (e NotFoundErr) Error() string {
	return fmt.Sprintf("Not found: %s", e.Message)
}
