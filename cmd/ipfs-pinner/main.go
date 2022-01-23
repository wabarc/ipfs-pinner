package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ipfs/go-cid"

	pinner "github.com/wabarc/ipfs-pinner"
)

type pin struct {
	path  string
	isCid bool
}

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

	pins := []pin{}
	for _, p := range files {
		pins = append(pins, pin{
			path:  p,
			isCid: isCid(p),
		})
	}

	mustExist(pins)

	handler := pinner.Config{
		Pinner: target,
		Apikey: apikey,
		Secret: secret,
	}
	var cid string
	var err error
	for _, p := range pins {
		if p.isCid {
			cid, err = handler.PinHash(p.path)
		} else {
			cid, err = handler.Pin(p.path)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "ipfs-pinner: %v\n", err)
		} else {
			fmt.Fprintf(os.Stdout, "%s  %s\n", cid, p.path)
		}
	}
}

func mustExist(path []pin) {
	b := &bytes.Buffer{}
	for _, p := range path {
		if p.isCid {
			continue
		}
		_, err := os.Stat(p.path)
		if _, ok := err.(*os.PathError); ok {
			fmt.Fprintln(b, err)
		}
	}
	if b.Len() > 0 {
		fmt.Println(b.String())
		os.Exit(1)
	}
}

func isCid(s string) bool {
	_, err := cid.Parse(s)
	if err != nil {
		return false
	}
	return true
}
