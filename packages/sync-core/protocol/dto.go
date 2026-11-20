package protocol

import "github.com/SirNacou/mirage/packages/sync-core/chunker"

// 1. POST /api/sync/blocks/check
type BlockCheckReq struct {
	Hashes []string `json:"hashes"`
}
type BlockCheckResp struct {
	Missing []string `json:"missing"`
}

// 2. POST /api/sync/files/commit
type CommitFileReq struct {
	RootID       string              `json:"root_id"`
	RelativePath string              `json:"relative_path"` // Normalized: "docs/notes.txt"
	Size         int64               `json:"size"`
	MtimeUnix    int64               `json:"mtime_unix"`
	IsDeleted    bool                `json:"is_deleted"`
	Chunks       []chunker.ChunkMeta `json:"chunks"`
}
type CommitFileResp struct {
	SyncSeq int64 `json:"sync_seq"`
}

// 3. GET /api/sync/changes?since={seq}
type FileChangeEvent struct {
	SyncSeq      int64               `json:"sync_seq"`
	RelativePath string              `json:"relative_path"`
	Size         int64               `json:"size"`
	MtimeUnix    int64               `json:"mtime_unix"`
	IsDeleted    bool                `json:"is_deleted"`
	Chunks       []chunker.ChunkMeta `json:"chunks"`
}
type SyncChangesResp struct {
	LatestSeq int64             `json:"latest_seq"`
	Changes   []FileChangeEvent `json:"changes"`
}
