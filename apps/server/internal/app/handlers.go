package app

import "github.com/stephenafamo/bob"

type Handlers struct {
	CommitFile        *CommitFileHandler
	FindMissingBlocks *CheckBlocksHandler
	ListChanges       *ListChangesHandler
	UploadBlock       *UploadBlockHandler
}

func NewHandlers(exec bob.Executor, dataDir string) (*Handlers, error) {
	uploadBlock, err := NewUploadBlockHandler(exec, dataDir)
	if err != nil {
		return nil, err
	}

	return &Handlers{
		CommitFile:        NewCommitFileHandler(exec),
		FindMissingBlocks: NewCheckBlocksHandler(exec),
		ListChanges:       NewListChangesHandler(exec),
		UploadBlock:       uploadBlock,
	}, nil
}
