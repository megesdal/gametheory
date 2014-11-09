package main

import (
	"encoding/json"
	"flag"
	"fmt"
)

var inputFile = flag.Bool("f", false, "input is a file name and not a matrix")

func main() {
	fmt.Println("Hello, 世界")

	flag.Parse()
	fmt.Println(flag.Args())

	who := flag.Arg(0)
	fmt.Printf("Hello, %s\n", who)

	if who == "lemke" {
		fmt.Println("I need an M, q, and d")
	} else if who == "nash" {
		fmt.Println("I need a payment matrix")
		//tableau := lemke.NewLCP(5)
		//fmt.Println(tableau)
		what := flag.Arg(1)
		if *inputFile {
			fmt.Println("Looking for file", what)
		} else {
			var payMatrix [][][]float64
			err := json.Unmarshal([]byte(what), &payMatrix)
			if err != nil {
				fmt.Println(err)
				return
			}

			// TODO: verify input
			fmt.Println(payMatrix)
		}
	}
}
