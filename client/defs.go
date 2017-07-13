package client

import "fmt"

// ErrUndefined for undefined message type.
type ErrUndefined int32

func (e ErrUndefined) Error() string {
	return fmt.Sprintf("undefined message type %d", e)
}
