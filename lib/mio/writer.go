package mio

import (
	"fmt"
)

type Writer struct {
}

func NewWriter() (w *Writer) {
	w = &Writer{}
	return
}

func (w *Writer) Print() {
	fmt.Println("...")
}
