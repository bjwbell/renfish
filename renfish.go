package main

import (
	"fmt"
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

func robotsHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("robothandler - start")
	index := struct{ Conf conf.Configuration }{conf.Config()}
	t, e := template.ParseFiles("robots.txt")
	if e != nil {
		panic(e)
	}
	log.Print("robothandler - execute")
	if e = t.Execute(w, index); e != nil {
		panic(e)
	}
}

func googleAdwordsVerifyHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("adwordsVerifyHandler - start")
	index := struct{ Conf conf.Configuration }{conf.Config()}
	t, e := template.ParseFiles("google41fd03a6c9348593.html")
	if e != nil {
		panic(e)
	}
	log.Print("adwordsVerifyHandler - execute")
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

func createsiteHandler(w http.ResponseWriter, r *http.Request) {
	conf := struct {
		Conf     renroll.Configuration
		Email    string
		SiteName string
	}{renroll.Config(), "", ""}
	conf.Conf.GPlusSigninCallback = "gSettings"
	conf.Conf.FacebookSigninCallback = "fbSettings"
	if err := r.ParseForm(); err != nil {
		auth.LogError(fmt.Sprintf("ERROR PARSEFORM, ERR: %v", err))
		t, _ := template.ParseFiles(
			"setuperror.html",
			"templates/header.html",
			"templates/topbar.html",
			"templates/bottombar.html")
		if err := t.Execute(w, conf); err != nil {
			auth.LogError(fmt.Sprintf("ERROR t.EXECUTE, ERR: %v", err))
		}
	}
	email := r.Form.Get("email")
	siteName := r.Form.Get("sitename")
	conf.Email = email
	conf.SiteName = "https://" + siteName + "." + r.Host
	if email == "" || siteName == "" {
		auth.LogError(fmt.Sprintf("MiSSING EMAIL or SITENAME, email: %v, sitename: %v", email, siteName))
		t, _ := template.ParseFiles(
			"setuperror.html",
			"templates/header.html",
			"templates/topbar.html",
			"templates/bottombar.html")
		if err := t.Execute(w, conf); err != nil {
			auth.LogError(fmt.Sprintf("ERROR t.EXECUTE, ERR: %v", err))
		}
	} else {
		t, _ := template.ParseFiles(
			"setup.html",
			"templates/header.html",
			"templates/topbar.html",
			"templates/bottombar.html")
		if err := t.Execute(w, conf); err != nil {
			auth.LogError(fmt.Sprintf("ERROR t.EXECUTE, ERR: %v", err))
		}
	}
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
	if err := t.Execute(w, conf); err != nil {
		auth.LogError(fmt.Sprintf("ERROR t.EXECUTE, ERR: %v", err))
	}
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
	http.HandleFunc("/createsite", createsiteHandler)

	http.HandleFunc("/index.html", indexHandler)
	http.HandleFunc("/robots.txt", robotsHandler)
	http.HandleFunc("/google41fd03a6c9348593.html", googleAdwordsVerifyHandler)

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css"))))
	http.Handle("/font-awesome-4.7.0/", http.StripPrefix("/font-awesome-4.7.0/", http.FileServer(http.Dir("./font-awesome-4.7.0"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("./fonts"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js"))))
	http.Handle("/screenshots/", http.StripPrefix("/screenshots/", http.FileServer(http.Dir("./screenshots"))))
	http.HandleFunc("/", indexHandler)

	// HTTP to HTTPS redirection
	// go func() {
	// 	err := http.ListenAndServe(":80", http.HandlerFunc(redir))
	// 	if err != nil {
	// 		log.Print("HTTP ListenAndServe :8080", err)
	// 		log.Print("Trying HTTP ListenAndServe :8080.")
	// 		panic(http.ListenAndServe(":8080", http.HandlerFunc(redir)))

	// 	}
	// }()

	go func() {
		err := http.ListenAndServe(":80", nil)
		if err != nil {
			log.Print("HTTP ListenAndServe :80, ", err)
			log.Print("Trying HTTP ListenAndServe :8080.")
			if err != nil {
				panic(http.ListenAndServe(":8080", nil))
			}

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
