package framing

import (
	"encoding/binary"
	"errors"
	"io"
)

// no need for locks since we will only have one goroutine reading and another
// one for writing on it for each peer. So no need for lock contetions.
// frameSize is set once in NewFraming and never mutated after construction
type Framing struct {
	frameSize uint32 // in bytes
}

func NewFraming(size uint32) *Framing {
	return &Framing{
		frameSize: size,
	}
}

func (f *Framing) Read(r io.Reader) ([]byte, error) {
	size := f.frameSize

	buf := make([]byte, size)

	_, err := io.ReadFull(r, buf)

	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	size = binary.BigEndian.Uint32(buf)

	payload := make([]byte, size)
	_, err = io.ReadFull(r, payload)
	if err != nil {
		return nil, err
	}

	return payload, err
}

func (f *Framing) Write(r io.Writer, payload []byte) error {
	size := f.frameSize

	sizeBuf := make([]byte, size)
	binary.BigEndian.PutUint32(sizeBuf, uint32(len(payload)))

	_, err := r.Write(sizeBuf)
	if err != nil {
		return err
	}

	if _, err := r.Write(payload); err != nil {
		return err
	}

	return nil
}
