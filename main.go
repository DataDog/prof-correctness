package main

import (
	"fmt"
	"regexp"
)

func main() {
	fmt.Printf("This is not the real main, this is just a small helper main to open the json data \n")
	fmt.Printf("Run the go test -v -run TestScenarios to run the tests\n")
	filter := regexp.MustCompile("<?php;main;b;standard\\|str_repeat$")
	stringMatch := "<?php;main;b;standard|str_repeat"
	if filter.MatchString(stringMatch) {
		println("Yes, haha")
	} else {
		println("No")
	}
}
