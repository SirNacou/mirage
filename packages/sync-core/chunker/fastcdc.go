package chunker

import (
	"encoding/hex"
	"io"

	chunkers "github.com/PlakarKorp/go-cdc-chunkers"
	"github.com/zeebo/blake3"
)

type ChunkMeta struct {
	Hash   string `json:"hash"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

const MB = 1024 * 1024

var cdcOpts = &chunkers.ChunkerOpts{
	MinSize:    1 * MB,
	MaxSize:    2 * MB,
	NormalSize: 4 * MB,
}

func SliceAndHash(r io.Reader, onChunk func(meta ChunkMeta, data []byte) error) ([]ChunkMeta, error) {
	ck, err := chunkers.NewChunker("fastcdc", r, cdcOpts)
	if err != nil {
		return nil, err
	}

	var manifest []ChunkMeta
	offset := 0

	for {
		data, err := ck.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		sum := blake3.Sum256(data)
		meta := ChunkMeta{
			Hash:   hex.EncodeToString(sum[:]),
			Offset: offset,
			Length: len(data),
		}
		manifest = append(manifest, meta)
		offset += len(data)

		if onChunk != nil {
			if err := onChunk(meta, data); err != nil {
				return nil, err
			}
		}
	}

	return manifest, nil
}
