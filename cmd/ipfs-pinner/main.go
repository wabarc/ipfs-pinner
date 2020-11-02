package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	Pinner "github.com/wabarc/ipfs-pinner"
)

var (
	pinner string
	apikey string
	secret string
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n\n")
		fmt.Fprintf(os.Stderr, "  ipfs-pinner [options] [file1] ... [fileN]\n\n")

		flag.PrintDefaults()
	}
	var basePrint = func() {
		fmt.Print("A CLI tool for pin files to IPFS.\n\n")
		flag.Usage()
		fmt.Fprint(os.Stderr, "\n")
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
			fmt.Print("Pinata require IPFS_PINNER_PINATA_API_KEY and IPFS_PINNER_PINATA_SECRET_API_KEY environment variables.\n\n")
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
	files := flag.Args()

	for _, path := range files {
		if _, err := os.Stat(path); err != nil {
			fmt.Fprintf(os.Stderr, "%s: no such file%v", path, "\n")
			continue
		}

		handle := Pinner.Config{Pinner: pinner, Apikey: apikey, Secret: secret}
		cid, err := handle.Pin(path)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v%v", path, err, "\n")
		} else {
			fmt.Fprintf(os.Stdout, "%s: %s%v", path, cid, "\n")
		}
	}
}
