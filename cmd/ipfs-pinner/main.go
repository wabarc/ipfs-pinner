package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/wabarc/ipfs-pinner/pkg/infura"
	"github.com/wabarc/ipfs-pinner/pkg/pinata"
)

var (
	pinner string
	apikey string
	secret string
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n\n")
		fmt.Fprintf(os.Stderr, "  ipfs-pinner [options] file\n\n")

		flag.PrintDefaults()
	}
	var basePrint = func() {
		fmt.Printf("A CLI tool for pin files to IPFS.\n\n")
		flag.Usage()
		fmt.Fprintf(os.Stderr, "\n")
	}

	flag.StringVar(&pinner, "p", "infura", "IPFS pinner, supports pinners: infura, pinata.")
	// flag.StringVar(&apikey, "u", "", "Pinner apikey or username.")
	// flag.StringVar(&secret, "P", "", "Pinner sceret or password.")

	flag.Parse()

	args := flag.Args()

	if strings.ToLower(pinner) == "pinata" {
		apikey = os.Getenv("IPFS_PINNER_PINATA_API_KEY")
		secret = os.Getenv("IPFS_PINNER_PINATA_SECRET_API_KEY")
		if apikey == "" || secret == "" {
			fmt.Printf("Pinata require IPFS_PINNER_PINATA_API_KEY and IPFS_PINNER_PINATA_SECRET_API_KEY environment variables.\n\n")
			basePrint()
			os.Exit(0)
		}
	}
	if len(args) < 1 {
		basePrint()
		os.Exit(0)
	}

}

func main() {
	args := flag.Args()
	filepath := args[0]
	if _, err := os.Stat(filepath); err != nil {
		fmt.Fprint(os.Stderr, err, "\n")
		os.Exit(1)
	}

	var cid string
	var err error

	switch pinner {
	default:
		err = fmt.Errorf("%s", "unknow pinner")
	case "infura":
		cid, err = infura.PinFile(filepath)
	case "pinata":
		pnt := pinata.Pinata{Apikey: apikey, Secret: secret}
		cid, err = pnt.PinFile(filepath)
	}

	if err != nil {
		fmt.Fprint(os.Stderr, err, "\n")
		os.Exit(1)
	} else {
		fmt.Fprint(os.Stdout, cid, "\n")
		os.Exit(0)
	}
}
