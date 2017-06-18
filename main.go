package main

import (
	"fmt"
	"github.com/RangelReale/osin"
	"github.com/RangelReale/osin/example"
	"net/http"
	"net/url"
)

// サーバーを singleton で実装
var sharedServer *osin.Server = newServer()

func getServer() *osin.Server {
	return sharedServer
}
func newServer() *osin.Server {
	fmt.Printf("new server")
	cfg := osin.NewServerConfig()
	cfg.AllowGetAccessRequest = true
	cfg.AllowClientSecretInParams = true
	return osin.NewServer(cfg, NewStorage())
}

func main() {

	// Authorization code endpoint
	http.HandleFunc("/index", indexHandler)
	http.HandleFunc("/authorize", authorizeHandler)
	http.HandleFunc("/token", tokenHandler)
	http.HandleFunc("/info", infoHandler)
	http.HandleFunc("/app", appHandler)
	http.HandleFunc("/app_auth/code", appAuthCodeHandler)

	// Server Start
	http.ListenAndServe(":8888", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello derabon")
}

func authorizeHandler(w http.ResponseWriter, r *http.Request) {

	s := getServer()
	resp := getServer().NewResponse()
	defer resp.Close()

	if ar := s.HandleAuthorizeRequest(resp, r); ar != nil {
		// HANDLE LOGIN PAGE HERE
		ar.Authorized = true
		s.FinishAuthorizeRequest(resp, r, ar)
	}
	osin.OutputJSON(resp, w, r)
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	s := getServer()
	resp := s.NewResponse()
	defer resp.Close()

	if ar := s.HandleAccessRequest(resp, r); ar != nil {
		ar.Authorized = true
		s.FinishAccessRequest(resp, r, ar)
	}
	osin.OutputJSON(resp, w, r)
}

// Information endpoint
func infoHandler(w http.ResponseWriter, r *http.Request) {
	s := getServer()
	resp := s.NewResponse()
	defer resp.Close()

	if ir := s.HandleInfoRequest(resp, r); ir != nil {
		s.FinishInfoRequest(resp, r, ir)
	}
	osin.OutputJSON(resp, w, r)
}

func appHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<html><body>"))
	w.Write([]byte(fmt.Sprintf("<a href=\"/authorize?response_type=code&client_id=1234&state=xyz&scope=everything&redirect_uri=%s\">Login</a><br/>", url.QueryEscape("http://localhost:8888/app_auth/code"))))
	w.Write([]byte("</body></html>"))
}

func appAuthCodeHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	code := r.Form.Get("code")

	w.Write([]byte("<html><body>"))
	w.Write([]byte("APP AUTH - CODE<br/>"))
	defer w.Write([]byte("</body></html>"))

	if code == "" {
		w.Write([]byte("Nothing to do"))
		return
	}

	jr := make(map[string]interface{})

	// build access code url
	aUrl := fmt.Sprintf("/token?grant_type=authorization_code&client_id=1234&client_secret=aabbccdd&state=xyz&redirect_uri=%s&code=%s",
		url.QueryEscape("http://localhost:8888/app_auth/code"), url.QueryEscape(code))

	// if parse, download and parse json
	if r.Form.Get("doparse") == "1" {
		err := example.DownloadAccessToken(fmt.Sprintf("http://localhost:14000%s", aUrl),
			&osin.BasicAuth{"1234", "aabbccdd"}, jr)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.Write([]byte("<br/>"))
		}
	}

	// show json error
	if erd, ok := jr["error"]; ok {
		w.Write([]byte(fmt.Sprintf("ERROR: %s<br/>\n", erd)))
	}

	// show json access token
	if at, ok := jr["access_token"]; ok {
		w.Write([]byte(fmt.Sprintf("ACCESS TOKEN: %s<br/>\n", at)))
	}

	w.Write([]byte(fmt.Sprintf("FULL RESULT: %+v<br/>\n", jr)))

	// output links
	w.Write([]byte(fmt.Sprintf("<a href=\"%s\">Goto Token URL</a><br/>", aUrl)))

	cururl := *r.URL
	curq := cururl.Query()
	curq.Add("doParse", "1")
	cururl.RawQuery = curq.Encode()
	w.Write([]byte(fmt.Sprintf("<a href=\"%s\">Download Token</a><br/>", cururl.String())))
}
