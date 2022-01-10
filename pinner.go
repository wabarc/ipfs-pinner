package pinner // import "github.com/wabarc/ipfs-pinner"

import (
	"errors"
	"io"
	"os"

	"github.com/wabarc/ipfs-pinner/pkg/infura"
	"github.com/wabarc/ipfs-pinner/pkg/pinata"
)

// Config represents pinner's configuration. Pinner is the identifier of
// the target IPFS service.
type Config struct {
	Pinner string
	Apikey string
	Secret string
}

// Pin pins a file to a network and returns a content id and an error. The file
// is an interface to access the file. It's contents may be either stored in
// memory or on disk. If stored on disk, it's underlying concrete type should
// be a file path. If it is in memory, it should be an *io.Reader or byte slice.
func (cfg *Config) Pin(file interface{}) (cid string, err error) {
	errPinner := errors.New("unknown pinner")
	switch v := file.(type) {
	case string:
		if _, err := os.Stat(v); err != nil {
			return "", err
		}
		switch cfg.Pinner {
		default:
			err = errPinner
		case "infura":
			cid, err = infura.PinFile(v)
		case "pinata":
			pnt := &pinata.Pinata{Apikey: cfg.Apikey, Secret: cfg.Secret}
			cid, err = pnt.PinFile(v)
		}
	case io.Reader:
		switch cfg.Pinner {
		default:
			err = errPinner
		case "infura":
			inf := infura.Infura{ProjectID: cfg.Apikey, ProjectSecret: cfg.Secret}
			cid, err = inf.PinWithReader(v)
		case "pinata":
			pnt := &pinata.Pinata{Apikey: cfg.Apikey, Secret: cfg.Secret}
			cid, err = pnt.PinWithReader(v)
		}
	case []byte:
		switch cfg.Pinner {
		default:
			err = errPinner
		case "infura":
			inf := infura.Infura{ProjectID: cfg.Apikey, ProjectSecret: cfg.Secret}
			cid, err = inf.PinWithBytes(v)
		case "pinata":
			pnt := &pinata.Pinata{Apikey: cfg.Apikey, Secret: cfg.Secret}
			cid, err = pnt.PinWithBytes(v)
		}
	default:
		return "", errors.New("unhandled file")
	}

	return cid, err
}
