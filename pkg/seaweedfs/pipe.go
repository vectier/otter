package seaweedfs

import (
	"sync"

	"github.com/rs/zerolog/log"
)

var (
	maxPipeBufferSize = 64 * 1024 // 64 kB
)

type pipeBufferPool struct {
	pool *sync.Pool
}

func newPipeBufferPool() *pipeBufferPool {
	pool := sync.Pool{
		New: func() any {
			log.Debug().Msg("create a new pipe buffer in the pool")
			buf := make(pipeBuffer, maxPipeBufferSize)
			return &buf
		},
	}
	return &pipeBufferPool{pool: &pool}
}

func (p *pipeBufferPool) Get() *pipeBuffer {
	return p.pool.Get().(*pipeBuffer)
}

func (p *pipeBufferPool) Release(b *pipeBuffer) {
	b.Reset()
	p.pool.Put(b)
}

type pipeBuffer []byte

func (b *pipeBuffer) Reset() {
	*b = (*b)[:maxPipeBufferSize]
}
