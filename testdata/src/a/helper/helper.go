package helper

import "fmt"

func Wrap(err error) error {
	return fmt.Errorf("wrapped: %w", err)
}
