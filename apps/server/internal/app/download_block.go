package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/SirNacou/mirage/apps/server/internal/domain"
)

type DownloadBlockResponse struct {
	BlockStream BlockStream
}
type BlockStream struct {
	Body io.ReadCloser
	Size int64
}
type DownloadBlockHandler struct {
	blockDir string
}

func NewDownloadBlockHandler(blockDir string) *DownloadBlockHandler {
	return &DownloadBlockHandler{
		blockDir: blockDir,
	}
}

func (h *DownloadBlockHandler) DownloadBlock(
	ctx context.Context,
	rawHash string,
) (*DownloadBlockResponse, error) {
	hash, err := domain.ParseBlockHash(rawHash)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(hash.DiskPath(h.blockDir))
	if err != nil {
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("stat block file: %w", err)
	}

	return &DownloadBlockResponse{
		BlockStream: BlockStream{
			Body: f,
			Size: info.Size(),
		},
	}, nil
}
