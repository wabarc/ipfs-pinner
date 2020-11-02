package pinner // import "github.com/wabarc/ipfs-pinner"

import (
	"fmt"
	"os"

	"github.com/wabarc/ipfs-pinner/pkg/infura"
	"github.com/wabarc/ipfs-pinner/pkg/pinata"
)

type Config struct {
	Pinner string
	Apikey string
	Secret string
}

// Pin file to pinning network using filepath
func (cfg *Config) Pin(filepath string) (string, error) {
	if _, err := os.Stat(filepath); err != nil {
		return "", fmt.Errorf("%s: no such file%v", filepath, "\n")
	}

	var cid string
	var err error

	switch cfg.Pinner {
	default:
		err = fmt.Errorf("%s", "unknow pinner")
	case "infura":
		cid, err = infura.PinFile(filepath)
	case "pinata":
		pnt := pinata.Pinata{Apikey: cfg.Apikey, Secret: cfg.Secret}
		cid, err = pnt.PinFile(filepath)
	}

	return cid, err
}
