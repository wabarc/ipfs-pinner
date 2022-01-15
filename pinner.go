package pinner // import "github.com/wabarc/ipfs-pinner"

import (
	"errors"
	"io"
	"os"

	"github.com/wabarc/ipfs-pinner/pkg/infura"
	"github.com/wabarc/ipfs-pinner/pkg/nftstorage"
	"github.com/wabarc/ipfs-pinner/pkg/pinata"
	"github.com/wabarc/ipfs-pinner/pkg/web3storage"
)

const (
	Infura      = "infura"
	Pinata      = "pinata"
	NFTStorage  = "nftstorage"
	Web3Storage = "web3storage"
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
	// TODO using generics
	errPinner := errors.New("unknown pinner")
	switch v := file.(type) {
	case string:
		if _, err := os.Stat(v); err != nil {
			return "", err
		}
		switch cfg.Pinner {
		default:
			err = errPinner
		case Infura:
			cid, err = infura.PinFile(v)
		case Pinata:
			pnt := &pinata.Pinata{Apikey: cfg.Apikey, Secret: cfg.Secret}
			cid, err = pnt.PinFile(v)
		case NFTStorage:
			nft := &nftstorage.NFTStorage{Apikey: cfg.Apikey}
			cid, err = nft.PinFile(v)
		case Web3Storage:
			web3 := &web3storage.Web3Storage{Apikey: cfg.Apikey}
			cid, err = web3.PinFile(v)
		}
	case io.Reader:
		switch cfg.Pinner {
		default:
			err = errPinner
		case Infura:
			inf := infura.Infura{ProjectID: cfg.Apikey, ProjectSecret: cfg.Secret}
			cid, err = inf.PinWithReader(v)
		case Pinata:
			pnt := &pinata.Pinata{Apikey: cfg.Apikey, Secret: cfg.Secret}
			cid, err = pnt.PinWithReader(v)
		case NFTStorage:
			nft := &nftstorage.NFTStorage{Apikey: cfg.Apikey}
			cid, err = nft.PinWithReader(v)
		case Web3Storage:
			web3 := &web3storage.Web3Storage{Apikey: cfg.Apikey}
			cid, err = web3.PinWithReader(v)
		}
	case []byte:
		switch cfg.Pinner {
		default:
			err = errPinner
		case Infura:
			inf := infura.Infura{ProjectID: cfg.Apikey, ProjectSecret: cfg.Secret}
			cid, err = inf.PinWithBytes(v)
		case Pinata:
			pnt := &pinata.Pinata{Apikey: cfg.Apikey, Secret: cfg.Secret}
			cid, err = pnt.PinWithBytes(v)
		case NFTStorage:
			nft := &nftstorage.NFTStorage{Apikey: cfg.Apikey}
			cid, err = nft.PinWithBytes(v)
		case Web3Storage:
			web3 := &web3storage.Web3Storage{Apikey: cfg.Apikey}
			cid, err = web3.PinWithBytes(v)
		}
	default:
		return "", errors.New("unhandled file")
	}

	return cid, err
}
