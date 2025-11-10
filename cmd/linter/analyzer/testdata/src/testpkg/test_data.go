// Package testpkg contains test data for analyzer.
package testpkg

import (
	"log"
	mylog "log"
	"os"
	myos "os"
)

// BadFunction is not correct for custom linter
func BadFunction() {
	panic("should be reported") // want "panic is not allowed"

	log.Fatal("should be reported")               // want "log.Fatal is not allowed outside main function of main package"
	log.Fatalf("should be reported: %s", "error") // want "log.Fatal is not allowed outside main function of main package"
	log.Fatalln("should be reported")             // want "log.Fatal is not allowed outside main function of main package"

	os.Exit(1) // want "os.Exit is not allowed outside main function of main package"
}

// BadFunctionWithAliases tests imports with aliases
func BadFunctionWithAliases() {
	mylog.Fatal("should be reported with alias")   // want "log.Fatal is not allowed outside main function of main package"
	mylog.Fatalf("should be reported with alias")  // want "log.Fatal is not allowed outside main function of main package"
	mylog.Fatalln("should be reported with alias") // want "log.Fatal is not allowed outside main function of main package"

	myos.Exit(1) // want "os.Exit is not allowed outside main function of main package"
}

// CorrectFunction is correct for custom linter
func CorrectFunction() {
	log.Println("this is ok")
	println("this is also ok")
}
