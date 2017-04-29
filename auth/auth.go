package auth

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"runtime"
	"strings"
	"time"

	"github.com/bjwbell/renfish/conf"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func googleOAuth2Config() *oauth2.Config {
	appConf := conf.Config()
	conf := &oauth2.Config{
		ClientID:     appConf.GoogleClientId,
		ClientSecret: appConf.GoogleClientSecret,
		RedirectURL:  "postmessage",
		Scopes:       []string{},
		Endpoint:     google.Endpoint,
	}
	return conf
}

func readHttpBody(response *http.Response) string {
	fmt.Println("Reading body")
	bodyBuffer := make([]byte, 5000)
	var str string
	count, err := response.Body.Read(bodyBuffer)
	for ; count > 0; count, err = response.Body.Read(bodyBuffer) {
		if err != nil {

		}
		str += string(bodyBuffer[:count])
	}
	return str
}

func Oauth2callback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if code == "" {
		log.Print("oauth2callback - NO CODE")
		email := "dummy@dummy.com"
		w.Write([]byte(email))
		return
	}
	log.Print("oauth2callback - url: " + r.URL.RawQuery)
	log.Print("oauth2callback - code: " + code)

	newAccount := r.FormValue("new_account")
	//email := getGPlusEmail(code)
	token := getGPlusToken(r)
	email := getGPlusEmail(token)
	if newAccount == "true" {
		log.Print("oauth2callback - NEW ACCOUNT")
		InterestedUser(email, "oauth2callback")
		//dbCreate(email)
		//dbInsert(email, "#1", "", 0, "", "", "", "", "", "", "", "")
	}
	w.Write([]byte(email))
}

func getGPlusToken(r *http.Request) oauth2.Token {
	accessToken := r.FormValue("access_token")
	if accessToken == "" {
		log.Fatal("getGPlusToken - NO ACCESS TOKEN")
	}
	return oauth2.Token{AccessToken: accessToken,
		TokenType:    "Bearer",
		RefreshToken: "",
		Expiry:       time.Now().Add(time.Hour)}
}

func getGPlusEmail(tok oauth2.Token) string {
	conf := googleOAuth2Config()
	client := conf.Client(oauth2.NoContext, &tok)
	response, err := client.Get("https://www.googleapis.com/plus/v1/people/me?fields=emails")
	// handle err. You need to change this into something more robust
	// such as redirect back to home page with error message
	if err != nil {
		log.Print("getGPlusEmail - COULDN'T GET PROFILE INFO WITH TOK, ERR:")
		log.Fatal(err)
	}
	str := readHttpBody(response)
	type Email struct {
		Value string
		Type  string
	}
	type OAuth2Response struct {
		Kind   string
		Emails []Email
		Id     string
	}
	log.Print("getemail - response: " + str)
	dec := json.NewDecoder(strings.NewReader(str))
	var m OAuth2Response
	if err := dec.Decode(&m); err != nil {
		logError(fmt.Sprintf("COULDN'T DECODE OAUTH2 RESPONSE, ERR: %v", err))
		log.Fatal(err)
	}
	for _, v := range m.Emails {
		log.Print("getemail - email (value, type): " + v.Value + ", " + v.Type)
	}

	email := "dummy@dummy.com"
	if len(m.Emails) != 1 {
		log.Print("getemail - NO VALID EMAIL OR TOO MANY")

	} else {
		email = m.Emails[0].Value
	}
	return email
}

func GetGPlusEmailHandler(w http.ResponseWriter, r *http.Request) {
	email := getGPlusEmail(getGPlusToken(r))
	w.Write([]byte(email))
}

func CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	loginMethod := r.FormValue("loginmethod")
	if email == "" {
		logError("CREATE ACCOUNT HANDLER - NO EMAIL")
		success := "false"
		w.Write([]byte(success))
		return
	}
	log.Printf("createAccountHandler - NEW ACCOUNT: %s", email)
	InterestedUser(email, loginMethod)
	//dbCreate(email)
	//dbInsert(email, "#1", "", 0, "", "", "", "", "", "", "", "")
	success := "true"
	w.Write([]byte(success))
}

func SigninFormHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("signinform - begin")
	t, _ := template.ParseFiles("templates/signinform.html")
	log.Print("signinform - execute")
	t.Execute(w, struct{ Conf conf.Configuration }{conf.Config()})
}

func SendAdminEmail(emailAddress string, subject string, body string) {
	configuration := conf.Config()
	content :=
		"To: " + configuration.GmailAddress + "\r\n" +
			"Subject: " + subject + "\r\n\r\n" +
			body
	auth := smtp.PlainAuth("", configuration.GmailAddress, configuration.GmailPassword, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, emailAddress,
		[]string{configuration.GmailAddress}, []byte(content))
	if err != nil {
		log.Print("sendEmail - ERROR:")
		log.Fatal(err)
	}
}

func InterestedUser(emailAddress string, loginMethod string) {
	subject := "Renfish Notification Clicked (" + emailAddress + ")"
	body := "Interested user: " + emailAddress + ".\r\n" +
		"Login method: " + loginMethod + "."
	SendAdminEmail(emailAddress, subject, body)
}

func LogErrorHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("logerrorhandler - begin")
	error := r.FormValue("error")
	SendAdminEmail(conf.Config().GmailAddress, "Renfish JS Error", error)
}

// The logError function prints the error to log.Print and also emails it to the email
// address, GmailAddress, in the config file
func logError(error string) {
	buf := make([]byte, 1<<16)
	runtime.Stack(buf, true)
	trace := fmt.Sprintf("%s", buf)
	msg := "Go Error\r\nError Message: " + error + "\r\n\r\nStack Trace:\r\n" + trace
	log.Print("logError:\n" + error)
	SendAdminEmail(conf.Config().GmailAddress, "Renfish Go Error", msg)
}
