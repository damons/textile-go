package core

import (
	"context"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// apiVersion is the api version
const apiVersion = "v0"

// apiHost is the instance used by the daemon
var apiHost *api

// api is a limited HTTP REST API for the cmd tool
type api struct {
	addr   string
	server *http.Server
	node   *Textile
}

// StartApi starts the host instance
func (t *Textile) StartApi(addr string) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = t.writer
	apiHost = &api{addr: addr, node: t}
	apiHost.Start()
}

// StopApi starts the host instance
func (t *Textile) StopApi() error {
	return apiHost.Stop()
}

// ApiAddr returns the api address
func (t *Textile) ApiAddr() string {
	if apiHost == nil {
		return ""
	}
	return apiHost.addr
}

// Start starts the http api
func (a *api) Start() {
	// setup router
	router := gin.Default()
	router.GET("/", func(g *gin.Context) {
		g.JSON(http.StatusOK, gin.H{
			"cafe_version": apiVersion,
			"node_version": Version,
		})
	})
	router.GET("/health", func(g *gin.Context) {
		g.Writer.WriteHeader(http.StatusNoContent)
	})

	// v0 routes
	v0 := router.Group("/api/v0")
	{
		v0.GET("/peer", a.peer)
		v0.GET("/address", a.address)
		v0.GET("/ping", a.ping)

		v0.POST("/threads", a.addThreads)
		v0.GET("/threads", a.lsThreads)
		v0.GET("/threads/:id", a.getThreads)
		v0.DELETE("/threads/:id", a.rmThreads)
		v0.GET("/threads/:id/updates", a.streamThreads)
		v0.POST("/invite/create", a.inviteThreads)
		v0.POST("/invite/join", a.joinThreads)

		v0.POST("/images", a.addImages)

		v0.POST("/cafes", a.addCafes)
		v0.GET("/cafes", a.lsCafes)
		v0.GET("/cafes/:id", a.getCafes)
		v0.DELETE("/cafes/:id", a.rmCafes)
		v0.POST("/cafes/check_mail", a.checkMailCafes)
	}
	a.server = &http.Server{
		Addr:    a.addr,
		Handler: router,
	}

	// start listening
	errc := make(chan error)
	go func() {
		errc <- a.server.ListenAndServe()
		close(errc)
	}()
	go func() {
		for {
			select {
			case err, ok := <-errc:
				if err != nil && err != http.ErrServerClosed {
					log.Errorf("api error: %s", err)
				}
				if !ok {
					log.Info("api was shutdown")
					return
				}
			}
		}
	}()
	log.Infof("api listening at %s\n", a.server.Addr)
}

// Stop stops the http api
func (a *api) Stop() error {
	// Use timeout to force a deadline
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	if err := a.server.Shutdown(ctx); err != nil {
		log.Errorf("error shutting down api: %s", err)
		cancel() // TODO: Not sure if this is right, linter was complaining
		return err
	}
	cancel()
	return nil
}

// -- UTILITY ENDPOINTS -- //

func (a *api) peer(g *gin.Context) {
	pid, err := a.node.PeerId()
	if err != nil {
		a.abort500(g, err)
		return
	}
	g.String(http.StatusOK, pid.Pretty())
}

func (a *api) address(g *gin.Context) {
	addr, err := a.node.Address()
	if err != nil {
		a.abort500(g, err)
		return
	}
	g.String(http.StatusOK, addr)
}

func (a *api) ping(g *gin.Context) {
	args, err := a.readArgs(g)
	if err != nil {
		a.abort500(g, err)
		return
	}
	if len(args) == 0 {
		g.String(http.StatusBadRequest, "missing peer id")
		return
	}
	pid, err := peer.IDB58Decode(args[0])
	if err != nil {
		g.String(http.StatusBadRequest, err.Error())
		return
	}
	status, err := a.node.Ping(pid)
	if err != nil {
		a.abort500(g, err)
		return
	}
	g.String(http.StatusOK, string(status))
}

func (a *api) readArgs(g *gin.Context) ([]string, error) {
	header := g.Request.Header.Get("X-Textile-Args")
	var args []string
	for _, a := range strings.Split(header, ",") {
		arg := strings.TrimSpace(a)
		if arg != "" {
			args = append(args, arg)
		}
	}
	return args, nil
}

func (a *api) readOpts(g *gin.Context) (map[string]string, error) {
	header := g.Request.Header.Get("X-Textile-Opts")
	opts := make(map[string]string)
	for _, o := range strings.Split(header, ",") {
		opt := strings.TrimSpace(o)
		if opt != "" {
			parts := strings.Split(opt, "=")
			if len(parts) == 2 {
				// If we've already seen this argument, create an array of args?
				// TODO: Is this a good idea? Maybe better way to deal with list-based args?
				if _, ok := opts[parts[0]]; ok {
					opts[parts[0]] += "," + parts[1]
				} else {
					opts[parts[0]] = parts[1]
				}
			}
		}
	}
	return opts, nil
}

func (a *api) abort500(g *gin.Context, err error) {
	g.String(http.StatusInternalServerError, err.Error())
}
