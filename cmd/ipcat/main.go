package main

import (
	"flag"
	"log"
	"os"

	"github.com/client9/ipcat"
)

func main() {
	updateAWS := flag.Bool("aws", false, "update AWS records")
	datafile := flag.String("csvfile", "datacenters.csv", "read/write from this file")
	flag.Parse()

	filein, err := os.Open(*datafile)
	if err != nil {
		log.Fatalf("Unable to read %s: %s", *datafile, err)
	}
	set := ipcat.IntervalSet{}
	err = set.ImportCSV(filein)
	if err != nil {
		log.Fatalf("Unable to import: %s", err)
	}
	filein.Close()

	if *updateAWS {
		body, err := ipcat.DownloadAWS()
		if err != nil {
			log.Fatalf("Unable to download AWS rules: %s", err)
		}
		err = ipcat.UpdateAWS(&set, body)
		if err != nil {
			log.Fatalf("Unable to parse AWS rules: %s", err)
		}
	}

	fileout, err := os.OpenFile(*datafile, os.O_WRONLY, 0)
	if err != nil {
		log.Fatalf("Unable to open file to write: %s", err)
	}
	err = set.ExportCSV(fileout)
	if err != nil {
		log.Fatalf("Unable to export: %s", err)
	}
	fileout.Close()
}