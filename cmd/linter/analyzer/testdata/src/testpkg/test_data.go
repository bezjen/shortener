package testpkg

import (
	"log"
	"os"
)

func BadFunction() {
	panic("should be reported") // want "panic is not allowed"

	log.Fatal("should be reported")               // want "log.Fatal is not allowed outside main function of main package"
	log.Fatalf("should be reported: %s", "error") // want "log.Fatal is not allowed outside main function of main package"
	log.Fatalln("should be reported")             // want "log.Fatal is not allowed outside main function of main package"

	os.Exit(1) // want "os.Exit is not allowed outside main function of main package"
}

func AnotherFunction() {
	log.Println("this is ok")
	println("this is also ok")
}
