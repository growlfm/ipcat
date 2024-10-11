package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/growlfm/ipcat"
)

func CopyFile(src string, dest string) {
	bytesRead, err := os.ReadFile(src)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(dest, bytesRead, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	lookup := flag.String("l", "", "lookup an IP address")
	updateAWS := flag.Bool("aws", false, "update AWS records")
	updateAzure := flag.Bool("azure", false, "update Azure records")
	updateGoogle := flag.Bool("google", false, "update Google Cloud records")
	updateCloudflare := flag.Bool("cloudflare", false, "update Cloudflare records")
	updateFastly := flag.Bool("fastly", false, "update Fastly records")
	updateAkamai := flag.Bool("akamai", false, "update Akamai records")
	updateDigitalOcean := flag.Bool("digitalocean", false, "update DigitalOcean records")
	datafile := flag.String("csvfile", "datacenters.csv", "read/write from this file")
	statsfile := flag.String("statsfile", "datacenters-stats.csv", "write statistics to this file")
	addCIDR := flag.String("addcidr", "", "add this CIDR range to the data file [CIDR,name,url]")
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
	log.Printf("Loaded %d entries", set.Len())

	if *lookup != "" {
		rec, err := set.Contains(*lookup)
		if err != nil {
			log.Fatalf("Unable to find %s: %s", *lookup, err)
		}
		if rec == nil {
			log.Fatalf("Not found: %s", *lookup)
		}
		fmt.Printf("[%s:%s] %s %s\n", rec.LeftDots, rec.RightDots, rec.Name, rec.URL)
		return
	}

	if *updateAWS {
		body, err := ipcat.DownloadAWS()
		if err != nil {
			log.Fatalf("Unable to download AWS IP ranges: %s", err)
		}
		err = ipcat.UpdateAWS(&set, body)
		if err != nil {
			log.Fatalf("Unable to parse AWS IP ranges: %s", err)
		}
	}

	if *updateAzure {
		body, err := ipcat.DownloadAzure()
		if err != nil {
			log.Fatalf("Unable to download Azure IP ranges: %s", err)
		}
		err = ipcat.UpdateAzure(&set, body)
		if err != nil {
			log.Fatalf("Unable to parse Azure IP ranges: %s", err)
		}
	}

	if *updateGoogle {
		body, err := ipcat.DownloadGoogle()
		if err != nil {
			log.Fatalf("Unable to download Google Cloud IP ranges: %s", err)
		}
		err = ipcat.UpdateGoogle(&set, body)
		if err != nil {
			log.Fatalf("Unable to parse Google Cloud IP ranges: %s", err)
		}
	}

	if *updateCloudflare {
		body, err := ipcat.DownloadCloudflare()
		if err != nil {
			log.Fatalf("Unable to download Cloudflare IP ranges: %s", err)
		}
		err = ipcat.UpdateCloudflare(&set, body)
		if err != nil {
			log.Fatalf("Unable to parse Cloudflare IP ranges: %s", err)
		}
	}

	if *updateFastly {
		body, err := ipcat.DownloadFastly()
		if err != nil {
			log.Fatalf("Unable to download Fastly IP ranges: %s", err)
		}
		err = ipcat.UpdateFastly(&set, body)
		if err != nil {
			log.Fatalf("Unable to parse Fastly IP ranges: %s", err)
		}
	}

	if *updateAkamai {
		body, err := ipcat.DownloadAkamai()
		if err != nil {
			log.Fatalf("Unable to download Akamai IP ranges: %s", err)
		}
		err = ipcat.UpdateAkamai(&set, body)
		if err != nil {
			log.Fatalf("Unable to parse Akamai IP ranges: %s", err)
		}
	}

	if *updateDigitalOcean {
		body, err := ipcat.DownloadDigitalOcean()
		if err != nil {
			log.Fatalf("Unable to download DigitalOcean IP ranges: %s", err)
		}
		err = ipcat.UpdateDigitalOcean(&set, body)
		if err != nil {
			log.Fatalf("Unable to parse DigitalOcean IP ranges: %s", err)
		}
	}

	if *addCIDR != "" {
		t := strings.Split(*addCIDR, ",")
		if len(t) != 3 {
			log.Fatal("range must be in format: CIDR,name,url")
		}
		err := set.AddCIDR(t[0], t[1], t[2])
		if err != nil {
			log.Fatalf("Could not add range: %v", err)
		}
		log.Println("Range added successfully")
	}

	//  Make sure output dir exists
	const outputPath = "/tmp/ipcat"
	err = os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		log.Fatalf("Could not create output directory: %s. Reason: %s", outputPath, err)
	}

	if *statsfile != "" {
		fileout, err := os.OpenFile(*statsfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalf("Unable to open file to write: %s", err)
		}
		list := set.RankBySize()
		fileout.WriteString("Datacenter Name, Total IPs\n")
		for _, val := range list {
			name := val.Name
			if strings.Contains(name, ",") {
				name = fmt.Sprintf("%q", val.Name)
			}
			fileout.WriteString(fmt.Sprintf("%s,%d\n", name, val.Size))
		}
		fileout.Close()

		basename := filepath.Base(*statsfile)
		dest := filepath.Join(outputPath, basename)
		CopyFile(*statsfile, dest)
	}

	fileout, err := os.OpenFile(*datafile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Unable to open file to write: %s", err)
	}
	err = set.ExportCSV(fileout)
	if err != nil {
		log.Fatalf("Unable to export: %s", err)
	}
	fileout.Close()

	basename := filepath.Base(*datafile)
	dest := filepath.Join(outputPath, basename)
	CopyFile(*datafile, dest)
}
