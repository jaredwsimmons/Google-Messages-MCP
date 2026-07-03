package importer

import (
	"io"

	"github.com/jaredwsimmons/google-messages-mcp/internal/db"
)

// Importer reads messages from an external source and stores them.
type Importer interface {
	Import(store *db.Store, source io.Reader) (*ImportResult, error)
}

// ImportResult summarizes what happened during an import.
type ImportResult struct {
	ConversationsCreated int
	MessagesImported     int
	MessagesDuplicate    int
	Errors               []string
}
