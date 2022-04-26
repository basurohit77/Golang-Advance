package main

import (
	"log"

	"github.ibm.com/cloud-sre/oss-globals/tlog"
)

func main() {
	log.Println(tlog.Log() + "This project contains only common data definitions which are used by all TIP API producers and consumers.")
}
