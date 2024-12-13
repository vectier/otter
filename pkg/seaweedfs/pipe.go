package seaweedfs

import (
	"bytes"
	"io"
	"sync"

	"github.com/rs/zerolog/log"
)

var (
	maxPipeBufferSize = 64 * 1024 // 64 kB
)

type pipe struct {
	pool *sync.Pool
}

func newPipe() *pipe {
	pool := sync.Pool{
		New: func() interface{} {
			log.Debug().Msg("create a new pipe buffer in the pool")
			return new(bytes.Buffer)
		},
	}
	return &pipe{pool: &pool}
}

func (p *pipe) Pipe(dst io.Writer, src io.Reader) error {
	buf := p.get()
	defer p.release(buf)
	_, err := io.CopyBuffer(dst, src, buf.Bytes())
	return err
}

func (p *pipe) get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

func (p *pipe) release(b *bytes.Buffer) {
	b.Reset()
	p.pool.Put(b)
}
