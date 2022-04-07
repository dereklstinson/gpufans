package pipe

import (
	"io"
	"log"
	"os"
	"syscall"
)

type Named struct {
	rfile, wfile string
	rpipe, wpipe *os.File
	Read         <-chan string
	Write        chan<- string
}

func CreateNamedPipe(rfile, wfile string) (n *Named, err error) {
	n = new(Named)
	if err1 := syscall.Mkfifo(rfile, 0666); err1 != nil {
		log.Printf("Error in making rctrl.fifo: %v\n", err1)

	} else {
		n.rpipe, err = os.OpenFile(rfile, os.O_RDWR, os.ModeNamedPipe)
		if err != nil {
			return nil, err
		}

	}
	if err1 := syscall.Mkfifo(wfile, 0666); err1 != nil {
		log.Printf("Error in making wctrl.fifo: %v\n", err1)
	} else {
		n.wpipe, err = os.OpenFile(wfile, os.O_RDWR, os.ModeNamedPipe)
		if err != nil {
			if err1 := n.rpipe.Close(); err1 != nil {
				log.Printf("Error in closeing n.rpipe: %v\n", err1)
			}
			return nil, err

		}

	}

	return n, nil
}
func (n *Named) ReadPipe() string {

	data, err := io.ReadAll(n.rpipe)
	if err != nil {
		return ""
	}
	return string(data)
}
func (n *Named) Close() {

	if err := n.rpipe.Close(); err != nil {
		log.Printf("Error in n.rpipe.Close(): %v\n", err)
	}
	if err := os.Remove(n.rfile); err != nil {
		log.Printf("Error in os.Remove(n.rfile): %v\n", err)
	}
	if err := n.wpipe.Close(); err != nil {
		log.Printf("Error in n.rpipe.Close(): %v\n", err)
	}
	if err := os.Remove(n.wfile); err != nil {
		log.Printf("Error in os.Remove(n.rfile): %v\n", err)
	}

}
