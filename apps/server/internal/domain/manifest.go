package domain

import (
	"fmt"

	"github.com/SirNacou/mirage/packages/sync-core/chunker"
)

func ValidateManifest(totalSize int64, isDeleted bool, chunks []chunker.ChunkMeta) error {
	if isDeleted {
		if len(chunks) > 0 {
			return fmt.Errorf("deleted file must not contain chunks")
		}
		return nil
	}

	if totalSize < 0 {
		return fmt.Errorf("file size cannot be negative: %d", totalSize)
	}

	if totalSize == 0 {
		if len(chunks) > 0 {
			return fmt.Errorf("empty file must have 0 chunks")
		}
		return nil
	}

	var expectedOffset int64
	for i, c := range chunks {
		if _, err := ParseBlockHash(c.Hash); err != nil {
			return fmt.Errorf("chunk[%d]: %w", i, err)
		}
		if c.Length <= 0 {
			return fmt.Errorf("chunk[%d]: invalid length %d", i, c.Length)
		}
		if int64(c.Offset) != expectedOffset {
			return fmt.Errorf("chunk[%d]: expected offset %d, got %d", i, expectedOffset, c.Offset)
		}
		expectedOffset += int64(c.Length)
	}

	if expectedOffset != totalSize {
		return fmt.Errorf(
			"manifest byte sum %d does not match file size %d",
			expectedOffset,
			totalSize,
		)
	}
	return nil
}
