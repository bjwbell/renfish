package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/bjwbell/renfish/auth"
	"github.com/bjwbell/renfish/conf"
	"github.com/bjwbell/renfish/submit"
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

func contactHandler(w http.ResponseWriter, r *http.Request) {
	about := struct{ Conf conf.Configuration }{conf.Config()}
	t, _ := template.ParseFiles(
		"contact.html",
		"templates/header.html",
		"templates/topbar.html",
		"templates/bottombar.html")
	t.Execute(w, about)
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

func main() {
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/auth/getemail", auth.GetGPlusEmailHandler)
	http.HandleFunc("/contact", contactHandler)
	http.HandleFunc("/createaccount", auth.CreateAccountHandler)
	http.HandleFunc("/index", indexHandler)
	http.HandleFunc("/logerror", auth.LogErrorHandler)
	http.HandleFunc("/oauth2callback", auth.Oauth2callback)
	http.HandleFunc("/settings", settingsHandler)
	http.HandleFunc("/signinform", auth.SigninFormHandler)
	http.HandleFunc("/submit", submit.SubmitHandler)

	http.Handle("/", http.FileServer(http.Dir("./")))
	go func() {
		err := http.ListenAndServe(":80", nil)
		if err != nil {
			panic(http.ListenAndServe(":8080", nil))
		}
	}()

	cert := "/etc/letsencrypt/live/renfish.com/cert.pem"
	privkey := "/etc/letsencrypt/live/renfish.com/privkey.pem"
	err := http.ListenAndServeTLS(":443", cert, privkey, nil)
	if err != nil {
		cert = "./generate_cert/cert.pem"
		privkey = "./generate_cert/key.pem"
		err = http.ListenAndServeTLS(":10443", cert, privkey, nil)
		if err != nil {
			log.Print("HTTPS ListenAndServe: ", err)
		}
	}
}
