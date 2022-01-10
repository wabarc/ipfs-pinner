# ipfs-pinner

[![Go Report Card](https://goreportcard.com/badge/github.com/wabarc/ipfs-pinner)](https://goreportcard.com/report/github.com/wabarc/ipfs-pinner)
[![Go Reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/wabarc/ipfs-pinner)
[![Releases](https://img.shields.io/github/v/release/wabarc/ipfs-pinner.svg?include_prereleases&color=blue)](https://github.com/wabarc/ipfs-pinner/releases)
[![ipfs-pinner](https://snapcraft.io/ipfs-pinner/badge.svg)](https://snapcraft.io/ipfs-pinner)

`ipfs-pinner` is a toolkit to help upload files or specific content id to IPFS pinning services.

Supported Golang version: See [.github/workflows/testing.yml](./.github/workflows/testing.yml)

## Installation

Via Golang package get command

```sh
go get -u github.com/wabarc/ipfs-pinner/cmd/ipfs-pinner
```

Using [Snapcraft](https://snapcraft.io/ipfs-pinner) (on GNU/Linux)

```sh
snap install ipfs-pinner
```

## Usage

### Supported Pinning Services

#### [Infura](https://infura.io)

Infura is a freemium pinning service that doesn't require any additional setup.
It's the default one used. Please bear in mind that Infura is a free service,
so there is probably a rate-limiting.

##### How to enable

Command-line:

Use flag `-p infura`.
<!-- markdownlint-disable-file MD010 -->
```sh
$ ipfs-pinner
A CLI tool for pin files to IPFS.

Usage:

  ipfs-pinner [options] [file1] ... [fileN]

  -p string
       IPFS pinner, supports pinners: infura, pinata. (default "infura")
```
<!-- markdownlint-enable-file MD010 -->

Go package:
```go
import (
        "fmt"

        "github.com/wabarc/ipfs-pinner/pkg/infura"
)

func main() {
        cid, err := infura.PinFile("file-to-path");
        if err != nil {
                fmt.Sprintln(err)
                return
        }
        fmt.Println(cid)
}
```

or requests with project authentication

```go
import (
        "fmt"

        "github.com/wabarc/ipfs-pinner/pkg/infura"
)

func main() {
        inf := &infura.Infura{ProjectID: "your-project-id", ProjectSecret: "your-project-secret"}
        cid, err := inf.PinFile("file-to-path");
        if err != nil {
                fmt.Sprintln(err)
                return
        }
        fmt.Println(cid)
}
```

#### [Pinata](https://pinata.cloud)

Pinata is another freemium pinning service. It gives you more control over
what's uploaded. You can delete, label and add custom metadata. This service
requires signup.

##### Environment variables

Unix*:
```sh
IPFS_PINNER_PINATA_API_KEY=<api key>
IPFS_PINNER_PINATA_SECRET_API_KEY=<secret api key>
```

Windows:
```sh
set IPFS_PINNER_PINATA_API_KEY=<api key>
set IPFS_PINNER_PINATA_SECRET_API_KEY=<secret api key>
```

##### How to enable

Command-line:

Use flag `-p pinata`.
```sh
ipfs-pinner -p pinata file-to-path
```

Go package:
```go
import (
        "fmt"

        "github.com/wabarc/ipfs-pinner/pkg/pinata"
)

func main() {
        pnt := pinata.Pinata{Apikey: "your api key", Secret: "your secret key"}
        cid, err := pnt.PinFile("file-to-path");
        if err != nil {
                fmt.Sprintln(err)
                return
        }
        fmt.Println(cid)
}
```
## License

Permissive GPL 3.0 license, see the [LICENSE](https://github.com/wabarc/ipfs-pinner/blob/main/LICENSE) file for details.
