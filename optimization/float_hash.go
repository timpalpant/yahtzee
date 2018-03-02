package optimization

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"math"
	"sync"
)

var floatHasherPool = sync.Pool{
	New: func() interface{} {
		return newFloatHasher()
	},
}

type floatHasher struct {
	buf *bytes.Buffer
	enc *gob.Encoder
}

func newFloatHasher() *floatHasher {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	return &floatHasher{buf, enc}
}

func (fh *floatHasher) HashSlice(values []float32) string {
	fh.buf.Reset()
	fh.enc.Encode(values)
	h := sha256.Sum256(fh.buf.Bytes())
	return string(h[:])
}

func HashFloat32(v float32) string {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, math.Float32bits(v))
	return string(buf)
}
