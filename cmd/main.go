package main

import (
	"log"

	"github.com/motain/compass-compute/cmd/compute"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	compute.Execute()
}
