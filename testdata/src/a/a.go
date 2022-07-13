package a

import (
	"errors"
	"fmt"
	fmtalias "fmt"

	"a/helper"
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
	err3 = a(1)  // want "error wasn't returned"
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

	_ = fmt.Errorf("blank identifier") // want "error wasn't returned"
	_, _ = b() // want "error wasn't returned"

	err6 := a(6)
	if err6 != nil {
		return fmt.Errorf("direct return: %w", err6)
	}

	err7 := a(7)
	if err7 != nil {
		return fmt.Errorf("direct return with no wrap: %+v", err7)
	}

	err8 := a(8) // want "error wasn't returned"
	if err8 != nil {
		err8 = fmt.Errorf("new error")
		return err8
	}

	err9 := a(9)
	if err9 != nil {
		return fmtalias.Errorf("custom wrapper: %w", err9)
	}

	err10 := a(10)
	if err10 != nil {
		return helper.Wrap(err10)
	}

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

func c() (err error) {
	err = errors.New("missing error") // want "error wasn't returned"
	err = errors.New("named return")
	return
}
