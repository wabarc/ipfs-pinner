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
		usage := `A CLI tool for pin files or directory to IPFS.

Usage:

  ipfs-pinner [flags] [path]...

Flags:
`
		fmt.Fprintln(os.Stdout, usage)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stdout, "")
	}

	flag.StringVar(&target, "t", "infura", "IPFS pinner, supports pinners: infura, pinata, nftstorage, web3storage.")
	flag.StringVar(&apikey, "u", "", "Pinner apikey or username.")
	flag.StringVar(&secret, "p", "", "Pinner sceret or password.")
	flag.Parse()

	files := flag.Args()
	target = strings.ToLower(target)
	switch target {
	case pinner.Pinata:
		if apikey == "" {
			apikey = os.Getenv("IPFS_PINNER_PINATA_API_KEY")
		}
		if secret == "" {
			secret = os.Getenv("IPFS_PINNER_PINATA_SECRET_API_KEY")
		}
	case pinner.NFTStorage, pinner.Web3Storage:
		if apikey == "" {
			fmt.Println(target + " requires an apikey.")
			os.Exit(1)
		}
	case pinner.Infura:
		// Permit request without authorization
	default:
		flag.Usage()
		os.Exit(0)
	}
	if len(files) < 1 {
		flag.Usage()
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
