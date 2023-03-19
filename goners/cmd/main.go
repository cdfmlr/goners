package main

import (
	"log"
	"os"
)

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
