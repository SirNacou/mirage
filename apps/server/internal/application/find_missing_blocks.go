package application

import (
	"context"

	"github.com/SirNacou/mirage/apps/server/internal/db/gen/models"
	"github.com/SirNacou/mirage/apps/server/internal/domain"
	"github.com/stephenafamo/bob"
)

type FindMissingBlocksQuery struct {
	RawHashes []string
}

type FindMissingBlocksHandler struct {
	exec bob.Executor
}

func NewFindMissingBlocksHandler(exec bob.Executor) *FindMissingBlocksHandler {
	return &FindMissingBlocksHandler{exec: exec}
}

func (h *FindMissingBlocksHandler) Handle(ctx context.Context, q FindMissingBlocksQuery) ([]string, error) {
	if len(q.RawHashes) == 0 {
		return []string{}, nil
	}

	// Validate and deduplicate input hashes
	seen := make(map[string]struct{}, len(q.RawHashes))
	validHashes := make([]string, 0, len(q.RawHashes))
	for _, raw := range q.RawHashes {
		parsed, err := domain.ParseBlockHash(raw)
		if err != nil {
			return nil, err
		}
		s := parsed.String()
		if _, exists := seen[s]; !exists {
			seen[s] = struct{}{}
			validHashes = append(validHashes, s)
		}
	}

	existingRows, err := models.Chunks.Query(
		models.SelectWhere.Chunks.Hash.In(validHashes...),
	).All(ctx, h.exec)
	if err != nil {
		return nil, err
	}

	existing := make(map[string]struct{}, len(existingRows))
	for _, row := range existingRows {
		existing[row.Hash] = struct{}{}
	}

	missing := make([]string, 0)
	for _, hStr := range validHashes {
		if _, ok := existing[hStr]; !ok {
			missing = append(missing, hStr)
		}
	}
	return missing, nil
}