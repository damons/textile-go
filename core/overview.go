package core

import (
	"fmt"

	"github.com/textileio/textile-go/repo"
)

// Overview is a wallet overview object
type Overview struct {
	AccountPeerCount int `json:"account_peer_cnt"`
	ThreadCount      int `json:"thread_cnt"`
	FileCount        int `json:"file_cnt"`
	ContactCount     int `json:"contact_cnt"`
}

// Overview returns an overview object
func (t *Textile) Overview() (*Overview, error) {
	threads := t.datastore.Threads().Count()
	files := t.datastore.Blocks().Count(fmt.Sprintf("type=%d", repo.FilesBlock))
	contacts := t.datastore.Contacts().Count()

	return &Overview{
		AccountPeerCount: 0,
		ThreadCount:      threads,
		FileCount:        files,
		ContactCount:     contacts,
	}, nil
}
