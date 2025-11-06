// Package testpkg contains test data for analyzer.
package testpkg

import (
	"log"
	"os"
)

// BadFunction is not correct for custom linter
func BadFunction() {
	panic("should be reported") // want "panic is not allowed"

	log.Fatal("should be reported")               // want "log.Fatal is not allowed outside main function of main package"
	log.Fatalf("should be reported: %s", "error") // want "log.Fatal is not allowed outside main function of main package"
	log.Fatalln("should be reported")             // want "log.Fatal is not allowed outside main function of main package"

	os.Exit(1) // want "os.Exit is not allowed outside main function of main package"
}

// CorrectFunction is correct for custom linter
func CorrectFunction() {
	log.Println("this is ok")
	println("this is also ok")
}
