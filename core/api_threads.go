package core

import (
	"crypto/rand"
	peer "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	libp2pc "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mr-tron/base58/base58"
	"github.com/textileio/textile-go/repo"
)

func (a *api) joinThreads(g *gin.Context) {
	args, err := a.readArgs(g)
	if err != nil {
		a.abort500(g, err)
		return
	}
	if len(args) == 0 {
		g.String(http.StatusBadRequest, "missing invite id")
		return
	}
	type RequestBody struct {
		Key string `form:"key"`
	}
	body := &RequestBody{}
	_ = g.Bind(body)
	key, err := base58.Decode(body.Key)
	if err != nil {
		a.abort500(g, err)
		return
	}
	hash, err := a.node.AcceptThreadInvite(args[0], key)
	if err != nil {
		a.abort500(g, err)
		return
	}
	g.String(http.StatusOK, hash.B58String())
}

func (a *api) inviteThreads(g *gin.Context) {
	id := g.Param("id")
	thrd := a.node.Thread(id)
	if thrd == nil {
		g.String(http.StatusNotFound, "thread not found")
		return
	}
	username := "cafe"
	peerID, err := a.node.PeerId()
	if err == nil {
		username = peerID.Pretty()
	}
	var invite InviteInfo
	type RequestBody struct {
		ID string `form:"id"`
	}
	reqBody := &RequestBody{}
	_ = g.Bind(reqBody)
	if reqBody.ID == "" {
		// add it
		hash, key, err := thrd.AddExternalInvite()
		if err != nil {
			a.abort500(g, err)
			return
		}
		// create a structured invite
		invite = InviteInfo{
			Id:      hash.B58String(),
			Key:     base58.FastBase58Encoding(key),
			Inviter: username,
		}
	} else {
		targetID, err := peer.IDB58Decode(reqBody.ID)
		if err != nil {
			a.abort500(g, err)
			return
		}
		// add it
		hash, err := thrd.AddInvite(targetID)
		if err != nil {
			a.abort500(g, err)
			return
		}
		// create a structured invite
		invite = InviteInfo{
			Id:      hash.B58String(),
			Inviter: username,
		}
	}

	g.JSON(http.StatusOK, invite)
}

func (a *api) addThreads(g *gin.Context) {
	args, err := a.readArgs(g)
	if err != nil {
		a.abort500(g, err)
		return
	}
	if len(args) == 0 {
		g.String(http.StatusBadRequest, "missing thread name")
		return
	}
	sk, _, err := libp2pc.GenerateEd25519Key(rand.Reader)
	if err != nil {
		a.abort500(g, err)
		return
	}
	thrd, err := a.node.AddThread(args[0], sk, true)
	if err != nil {
		a.abort500(g, err)
		return
	}
	info, err := thrd.Info()
	if err != nil {
		a.abort500(g, err)
		return
	}
	g.JSON(http.StatusCreated, info)
}

func (a *api) lsThreads(g *gin.Context) {
	infos := make([]*ThreadInfo, 0)
	for _, thrd := range a.node.Threads() {
		info, err := thrd.Info()
		if err != nil {
			a.abort500(g, err)
			return
		}
		infos = append(infos, info)
	}
	g.JSON(http.StatusOK, infos)
}

func (a *api) getThreads(g *gin.Context) {
	id := g.Param("id")
	thrd := a.node.Thread(id)
	if thrd == nil {
		g.String(http.StatusNotFound, "thread not found")
		return
	}
	info, err := thrd.Info()
	if err != nil {
		a.abort500(g, err)
		return
	}
	g.JSON(http.StatusOK, info)
}

func (a *api) rmThreads(g *gin.Context) {
	id := g.Param("id")
	thrd := a.node.Thread(id)
	if thrd == nil {
		g.String(http.StatusNotFound, "thread not found")
		return
	}
	if _, err := a.node.RemoveThread(id); err != nil {
		a.abort500(g, err)
		return
	}
	g.String(http.StatusOK, "ok")
}

func (a *api) streamThreads(g *gin.Context) {
	id := g.Param("id")
	thrd := a.node.Thread(id)
	if thrd == nil {
		g.String(http.StatusNotFound, "thread not found")
		return
	}

	g.Request.ParseForm()
	types := g.Request.Form["type"]

	go func() {
		block := repo.Block{
			Id:       "id",
			Date:     time.Now(),
			Parents:  []string{},
			ThreadId: "12D3KooW9vSfCJkR4PofQCHgifRUSXWNZ81FUHiUCytXiuNHTtwi",
			AuthorId: "authorid",
			Type:     repo.TextBlock,
		}
		update := ThreadUpdate{
			Block:      block,
			ThreadId:   "12D3KooW9vSfCJkR4PofQCHgifRUSXWNZ81FUHiUCytXiuNHTtwi",
			ThreadName: "threadname",
		}

		for i := 0; i < 50; i++ {
			a.node.sendThreadUpdate(update)
			time.Sleep(time.Second * 1)
		}
	}()
	g.Stream(func(w io.Writer) bool {
		if update, ok := <-a.node.ThreadUpdateCh(); ok {
			if update.ThreadId == thrd.Id {
				for _, val := range types {
					i, err := strconv.Atoi(val)
					if err != nil {
						// We'll just ignore these issues
						continue
					}
					if update.Block.Type == repo.BlockType(i) {
						g.IndentedJSON(http.StatusOK, update)
						return true
					}
				}
			}
		}
		return false
	})
}
