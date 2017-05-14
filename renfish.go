package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bjwbell/renfish/auth"
	"github.com/bjwbell/renfish/conf"
	"github.com/bjwbell/renfish/submit"
	"github.com/bjwbell/renroll/src/renroll"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("indexhandler - start")
	index := struct{ Conf conf.Configuration }{conf.Config()}
	t, e := template.ParseFiles("idx.html", "templates/header.html", "templates/topbar.html", "templates/bottombar.html")
	if e != nil {
		panic(e)
	}
	log.Print("indexhandler - execute")
	if e = t.Execute(w, index); e != nil {
		panic(e)
	}
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	about := struct{ Conf conf.Configuration }{conf.Config()}
	t, _ := template.ParseFiles(
		"about.html",
		"templates/header.html",
		"templates/topbar.html",
		"templates/bottombar.html")
	t.Execute(w, about)
}

func unreleasedHandler(w http.ResponseWriter, r *http.Request) {
	conf := struct{ Conf renroll.Configuration }{renroll.Config()}
	conf.Conf.GPlusSigninCallback = "gSettings"
	conf.Conf.FacebookSigninCallback = "fbSettings"
	t, _ := template.ParseFiles(
		"unreleased.html",
		"templates/header.html",
		"templates/topbar.html",
		"templates/bottombar.html")
	t.Execute(w, conf)
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	conf := struct{ Conf conf.Configuration }{conf.Config()}
	conf.Conf.GPlusSigninCallback = "gSettings"
	conf.Conf.FacebookSigninCallback = "fbSettings"
	t, _ := template.ParseFiles(
		"settings.html",
		"templates/header.html",
		"templates/topbar.html",
		"templates/bottombar.html")
	t.Execute(w, conf)
}

func redir(w http.ResponseWriter, req *http.Request) {
	host := req.Host
	httpsPort := "443"
	if strings.Index(host, ":8080") != -1 {
		httpsPort = "8443"
	}
	host = strings.TrimSuffix(host, ":8080")
	host = strings.TrimSuffix(host, ":80")
	if httpsPort == "443" {
		http.Redirect(w, req, "https://"+host+req.RequestURI, http.StatusMovedPermanently)
	} else {
		http.Redirect(w, req, "https://"+host+":"+httpsPort+req.RequestURI, http.StatusMovedPermanently)
	}
}

func main() {
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/auth/getemail", auth.GetGPlusEmailHandler)
	http.HandleFunc("/createaccount", auth.CreateAccountHandler)
	http.HandleFunc("/index", indexHandler)
	http.HandleFunc("/logerror", auth.LogErrorHandler)
	http.HandleFunc("/oauth2callback", auth.Oauth2callback)
	http.HandleFunc("/settings", settingsHandler)
	http.HandleFunc("/signinform", auth.SigninFormHandler)
	http.HandleFunc("/submit", submit.SubmitHandler)
	http.HandleFunc("/unreleased", unreleasedHandler)

	http.Handle("/", http.FileServer(http.Dir("./")))
	go func() {
		err := http.ListenAndServe(":80", http.HandlerFunc(redir))
		if err != nil {
			log.Print("HTTP ListenAndServe :8080", err)
			log.Print("Trying HTTP ListenAndServe :8080.")
			panic(http.ListenAndServe(":8080", http.HandlerFunc(redir)))

		}
	}()

	cert := "/etc/letsencrypt/live/renfish.com/cert.pem"
	privkey := "/etc/letsencrypt/live/renfish.com/privkey.pem"
	if _, err := os.Stat(cert); os.IsNotExist(err) {
		log.Print("cert: ", err)
		cert = "./generate_cert/cert.pem"
	}
	if _, err := os.Stat(privkey); os.IsNotExist(err) {
		log.Print("cert: ", err)
		privkey = "./generate_cert/key.pem"
	}
	err := http.ListenAndServeTLS(":443", cert, privkey, nil)
	if err != nil {
		log.Print("HTTPS ListenAndServe :8443")
		err = http.ListenAndServeTLS(":8443", cert, privkey, nil)
		if err != nil {
			log.Print("HTTPS ListenAndServe: ", err)
		}
	}
}
