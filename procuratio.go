package yamaih

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (g *Gemini) handler(w http.ResponseWriter, r *http.Request) {
	extractIP(r)
	if len(r.URL.Path) <= 1 {
		fmt.Println("[ERR]", r.RemoteAddr, "400 Invalid Path", r.URL.String())
		http.Error(w, "400 Invalid Path", http.StatusBadRequest)
		return
	}
	k := r.URL.Query().Get("key")
	if k == "" {
		fmt.Println("[ERR]", r.RemoteAddr, "400 Empty API Key", r.URL.String())
		http.Error(w, "400 Empty API Key", http.StatusBadRequest)
		return
	}
	v := &Visit{
		UserKey: k,
		Time:    time.Now().UnixMilli(),
		IP:      r.RemoteAddr,
		Method:  r.Method,
		Path:    r.URL.Path,
		Query:   r.URL.RawQuery,
	}
	respstr := ""
	defer func() {
		v.WaitMilli = time.Now().UnixMilli() - v.Time
		if respstr != "" {
			v.Response = []byte(respstr)
		}
		g.visit(v)
		fmt.Println(v)
	}()
	apiver, _, _ := strings.Cut(r.URL.Path[1:], "/")
	if apiver != g.apiver {
		respstr = "400 Invalid API Version"
		v.Code = 400
		http.Error(w, respstr, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		respstr = "400 Bad Request: " + err.Error()
		v.Code = 400
		http.Error(w, respstr, http.StatusBadRequest)
		return
	}
	v.Request = data
	req, err := http.NewRequest(
		r.Method, api+r.URL.String(), bytes.NewReader(data),
	)
	if err != nil {
		respstr = "400 Bad Request: " + err.Error()
		v.Code = 400
		http.Error(w, respstr, http.StatusBadRequest)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		respstr = "500 Do: " + err.Error()
		v.Code = 500
		http.Error(w, respstr, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	v.Code = resp.StatusCode
	h := w.Header()
	for k, vs := range resp.Header {
		if len(vs) == 0 {
			continue
		}
		h.Set(k, vs[0])
		for _, v := range vs[1:] {
			h.Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	var b []byte
	if resp.ContentLength > 0 {
		b = make([]byte, 0, resp.ContentLength)
	}
	buf := bytes.NewBuffer(b)
	_, _ = io.Copy(io.MultiWriter(w, buf), resp.Body)
	v.Response = buf.Bytes()
}

// extractIP parse real IP addr to r.RemoteAddr from proxy
func extractIP(r *http.Request) {
	raddr := r.RemoteAddr
	if strings.Contains(raddr, "127.0.0.1") ||
		strings.Contains(raddr, "localhost") ||
		strings.Contains(raddr, "@") {
		realr := r.Header.Get("X-Forwarded-For")
		if len(realr) > 0 && !strings.Contains(realr, "@") {
			raddr = realr
		} else {
			realr = r.Header.Get("X-Real-IP")
			if len(realr) > 0 && !strings.Contains(realr, "@") {
				raddr = realr
			}
		}
	}
	r.RemoteAddr = raddr
}
