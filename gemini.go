package yamaih

import (
	"net/http"
	"sync"

	sql "github.com/FloatTech/sqlite"
)

const api = "https://generativelanguage.googleapis.com"

type Gemini struct {
	dbmu     sync.Mutex
	endpoint string
	apiver   string // apiver usually v1beta
	logdb    sql.Sqlite
	mux      *http.ServeMux
}

func NewGemini(endpoint, logfile, apiver string) *Gemini {
	g := &Gemini{
		endpoint: endpoint,
		apiver:   apiver,
		logdb:    sql.New(logfile),
		mux:      http.NewServeMux(),
	}
	err := g.initdb()
	if err != nil {
		panic(err)
	}
	g.mux.HandleFunc("/", g.handler)
	return g
}

func (g *Gemini) RunBlocking() error {
	return http.ListenAndServe(g.endpoint, g.mux)
}
