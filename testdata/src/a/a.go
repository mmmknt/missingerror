package a

import (
	"errors"
	"fmt"
)

func f() error {
	err0 := a(1) // OK
	if err0 != nil {
		return err0
	}

	err1 := a(1) // want "error wasn't returned"
	if err1 != nil {
		fmt.Println(err1)
	}

	err3 := a(0) // want "error wasn't returned"
	err3 = a(2)  // OK
	if err3 != nil {
		return err3
	}

	err4 := a(4)                   // want "error wasn't returned"
	if err4 := a(4); err4 != nil { // OK
		return err4
	}
	fmt.Println(err4)

	if err5 := a(5); err5 != nil { // want "error wasn't returned"
		fmt.Println(err5)
	}

	dummy, _ := b() // want "error wasn't returned"
	if dummy {
		// nothing to do
	}

	// TODO check blank identifier
	_ = a(0)
	_, _ = b()

	// TODO wrapped error case

	return nil
}

func a(i int) error {
	if i != 0 {
		return errors.New("dummy error")
	}
	return nil
}

func b() (bool, error) {
	return true, nil
}
