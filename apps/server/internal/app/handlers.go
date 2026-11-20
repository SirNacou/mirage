package app

import "github.com/stephenafamo/bob"

type Handlers struct {
	CommitFile        *CommitFileHandler
	DownloadBlock     *DownloadBlockHandler
	FindMissingBlocks *CheckBlocksHandler
	ListChanges       *ListChangesHandler
	UploadBlock       *UploadBlockHandler
}

func NewHandlers(exec bob.Executor, dataDir string) (*Handlers, error) {
	uploadBlock, err := NewUploadBlockHandler(exec, dataDir)
	if err != nil {
		return nil, err
	}
	downloadBlock := NewDownloadBlockHandler(uploadBlock.blockDir)

	return &Handlers{
		CommitFile:        NewCommitFileHandler(exec),
		DownloadBlock:     downloadBlock,
		FindMissingBlocks: NewCheckBlocksHandler(exec),
		ListChanges:       NewListChangesHandler(exec),
		UploadBlock:       uploadBlock,
	}, nil
}
