package testpkg

import "fmt"

func GoodFunc() error {
	return fmt.Errorf("error") // OK
}