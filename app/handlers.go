package main

import (
	"fmt"
	"os"
)

func echoHandler(s ...string) error {
	for _, x := range s {
		fmt.Printf("%s", x)
	}
	fmt.Print("\n")
	return nil
}

func exitHandler(s ...string) error {
	os.Exit(0)
	return nil
}
