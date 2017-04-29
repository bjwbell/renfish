package submit

import (
	"html/template"
	"net/http"

	"github.com/bjwbell/renfish/conf"
)

func SubmitHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles(
		"submit.html",
		"templates/header.html",
		"templates/bottombar.html")
	t.Execute(w, struct{ Conf conf.Configuration }{conf.Config()})
}
