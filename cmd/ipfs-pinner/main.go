package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	pinner "github.com/wabarc/ipfs-pinner"
)

func main() {
	var (
		target string
		apikey string
		secret string
	)

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

	flag.StringVar(&target, "t", "infura", "IPFS pinner, supports pinners: infura, pinata, nftstorage, web3storage.")
	flag.StringVar(&apikey, "u", "", "Pinner apikey or username.")
	flag.StringVar(&secret, "p", "", "Pinner sceret or password.")

	flag.Parse()

	files := flag.Args()
	target = strings.ToLower(target)

	switch target {
	case "pinata":
		apikey = os.Getenv("IPFS_PINNER_PINATA_API_KEY")
		secret = os.Getenv("IPFS_PINNER_PINATA_SECRET_API_KEY")
		if apikey == "" || secret == "" {
			fmt.Println("Pinata require IPFS_PINNER_PINATA_API_KEY and IPFS_PINNER_PINATA_SECRET_API_KEY environment variables.")
			os.Exit(1)
		}
	case "nftstorage", "web3storage":
		if apikey == "" {
			fmt.Println(target + " requires an apikey.")
			os.Exit(1)
		}
	case "infura":
		// Permit request without authorization
	default:
		basePrint()
		os.Exit(0)
	}
	if len(files) < 1 {
		fmt.Println("file path is missing.")
		os.Exit(1)
	}

	for _, path := range files {
		if _, err := os.Stat(path); err != nil {
			fmt.Fprintf(os.Stderr, "ipfs-pinner: %s: no such file or directory\n", path)
			continue
		}

		handle := pinner.Config{Pinner: target, Apikey: apikey, Secret: secret}
		cid, err := handle.Pin(path)

		if err != nil {
			fmt.Fprintf(os.Stderr, "ipfs-pinner: %v\n", err)
		} else {
			fmt.Fprintf(os.Stdout, "%s  %s\n", cid, path)
		}
	}
}
