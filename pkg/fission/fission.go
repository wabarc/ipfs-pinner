package fission

const FISSION_URI = "https://runfission.com/ipfs"

type Fission struct {
	Username string
	Password string
}

func (p *Fission) PinFile(filepath string) (string, error) {
	return "", nil
}

func (p *Fission) PinHash(hash string) (bool, error) {
	return true, nil
}
