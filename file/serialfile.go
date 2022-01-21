package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Node represents a serial files.
type Node struct {
	base  string
	path  string
	files []os.FileInfo
	stat  os.FileInfo
}

// NewSerialFile adopts serial files and returns a Node represents a file,
// directory, or special file.
func NewSerialFile(path string) (node *Node, err error) {
	node = new(Node)
	node.path = path
	stat, err := os.Stat(path)
	if err != nil {
		return node, fmt.Errorf("lookup path failed: %v", err)
	}
	node.stat = stat
	switch mode := stat.Mode(); {
	case mode.IsRegular():
		node.files = append(node.files, stat)
	case mode.IsDir():
		contents, err := ioutil.ReadDir(path)
		if err != nil {
			return node, fmt.Errorf("read directory failed: %v", err)
		}
		node.files = contents
	default:
		return node, fmt.Errorf("unrecognized file type for %s: %s", path, mode.String())
	}
	return
}

// MapDirectory sets up a new target directory by given path.
func (n *Node) MapDirectory(path string) {
	if n.stat.IsDir() {
		n.base = path
	}
}

// Mode returns a os.FileMode of Node
func (n *Node) Mode() os.FileMode {
	return n.stat.Mode()
}

// Size returns the file size of the Node.
func (n *Node) Size() (du int64, err error) {
	if len(n.files) == 0 {
		return 0, fmt.Errorf("node is empty")
	}

	for _, fi := range n.files {
		if fi.Mode().IsRegular() {
			du += fi.Size()
		}
		if fi.Mode().IsDir() {
			err = filepath.Walk(fi.Name(), func(p string, fi os.FileInfo, err error) error {
				if err != nil || fi == nil {
					return err
				}

				if fi.Mode().IsRegular() {
					du += fi.Size()
				}

				return nil
			})
			if err != nil {
				return 0, fmt.Errorf("walk directory failed: %v", err)
			}
		}
	}

	return du, err
}
