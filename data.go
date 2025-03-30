package yamaih

import (
	"strconv"
	"strings"
	"time"
)

type User struct {
	Key      string // Key user's api key
	LastTime int64  // LastTime last visit time
	LastIP   string // LastIP last visit IP
	Count    int64  // Count total visit count
}

type Visit struct {
	ID        *int
	UserKey   string
	Time      int64
	WaitMilli int64
	Code      int
	IP        string
	Method    string
	Path      string
	Query     string
	Request   []byte
	Response  []byte
}

func (v *Visit) String() string {
	sb := strings.Builder{}
	sb.WriteByte('[')
	sb.WriteString(strconv.FormatInt(v.WaitMilli, 10))
	sb.WriteString("ms] ")
	sb.WriteString(v.IP)
	sb.WriteByte(' ')
	sb.WriteString(v.Method)
	sb.WriteByte(' ')
	sb.WriteString(v.Path)
	return sb.String()
}

func (g *Gemini) initdb() error {
	err := g.logdb.Open(time.Hour)
	if err != nil {
		return err
	}
	_, err = g.logdb.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return err
	}
	err = g.logdb.Create("user", &User{})
	if err != nil {
		return err
	}
	return g.logdb.Create("visit", &Visit{},
		"FOREIGN KEY(UserKey) REFERENCES user(Key)",
	)
}

func (g *Gemini) visit(v *Visit) error {
	v.ID = nil
	g.dbmu.Lock()
	defer g.dbmu.Unlock()
	u := User{}
	_ = g.logdb.Find("user", &u, "WHERE Key=?", v.UserKey)
	u.Key = v.UserKey
	u.LastTime = v.Time
	u.LastIP = v.IP
	u.Count++
	err := g.logdb.Insert("user", &u)
	if err != nil {
		return err
	}
	return g.logdb.Insert("visit", v)
}
