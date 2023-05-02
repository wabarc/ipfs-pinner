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
//
//nolint:gocyclo
func (cfg *Config) Pin(path interface{}) (cid string, err error) {
	// TODO using generics
	errPinner := errors.New("unknown pinner")
	switch v := path.(type) {
	case string:
		_, err := os.Lstat(v)
		if err != nil {
			return "", err
		}
		switch cfg.Pinner {
		default:
			err = errPinner
		case Infura:
			inf := &infura.Infura{Apikey: cfg.Apikey, Secret: cfg.Secret}
			cid, err = inf.PinFile(v)
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
		return cid, err
	case io.Reader:
		switch cfg.Pinner {
		default:
			err = errPinner
		case Infura:
			inf := &infura.Infura{Apikey: cfg.Apikey, Secret: cfg.Secret}
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
		return cid, err
	case []byte:
		switch cfg.Pinner {
		default:
			err = errPinner
		case Infura:
			inf := &infura.Infura{Apikey: cfg.Apikey, Secret: cfg.Secret}
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
		return cid, err
	}

	return cid, errPinner
}

// PinHash pins from any IPFS node, returns the original cid and an error.
func (cfg *Config) PinHash(cid string) (string, error) {
	ok := false
	err := errors.New("unsupported pinner")
	switch cfg.Pinner {
	case Infura:
		inf := &infura.Infura{Apikey: cfg.Apikey, Secret: cfg.Secret}
		ok, err = inf.PinHash(cid)
	case Pinata:
		pnt := &pinata.Pinata{Apikey: cfg.Apikey, Secret: cfg.Secret}
		ok, err = pnt.PinHash(cid)
	}
	if ok {
		return cid, nil
	}

	return "", err
}
