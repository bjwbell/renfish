package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"

	"github.com/bjwbell/renfish/auth"
	"github.com/bjwbell/renfish/conf"
	"github.com/bjwbell/renfish/db"
	"github.com/bjwbell/renfish/submit"
)

type Configuration struct {
	GmailAddress           string
	GmailPassword          string
	GoogleClientId         string
	GoogleClientSecret     string
	GooglePlusScopes       string
	GPlusSigninCallback    string
	GoogleAnalyticsId      string
	FacebookScopes         string
	FacebookAppId          string
	FacebookSigninCallback string
}

func Config() Configuration {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatal(err)
	}
	return configuration
}

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
	conf := struct{ Conf Configuration }{Config()}
	conf.Conf.GPlusSigninCallback = "gSettings"
	conf.Conf.FacebookSigninCallback = "fbSettings"
	t, _ := template.ParseFiles(
		"unreleased.html",
		"templates/header.html",
		"templates/topbar.html",
		"templates/bottombar.html")
	t.Execute(w, conf)
}

func getNextIP() string {
	ips := db.DbGetIPs(db.DbName)
	return db.DbGetNextAvailableIP(ips)
}

func createSite(emailAddress, siteName string) {
	domain := siteName + "." + "renfish.com"
	// Add nginx conf file
	nginxConf := `server {
     listen 443 ssl;
    listen [::]:443 ssl;
    server_name  <site-name>;
    ssl_certificate     /etc/letsencrypt/live/<site-name>/cert.pem;
    ssl_certificate_key /etc/letsencrypt/live/<site-name>/privkey.pem;
    ssl_protocols       TLSv1 TLSv1.1 TLSv1.2;
    ssl_ciphers         HIGH:!aNULL:!MD5;
    location / {
             proxy_pass http://<ip-address>:8080;
            proxy_set_header Host $host;
    }
}
server {
    listen              80;
    server_name  <site-name>;
    location / {
            proxy_pass http://<ip-address>;
             proxy_set_header Host $host;
    }
}
`
	// START GOPHISH CONTAINER
	fmt.Println("STARTING GOPHISH CONTAINER")
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	imageName := "bjwbell/gophish-container"
	out3, err3 := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err3 != nil {
		panic(err3)
	}
	io.Copy(os.Stdout, out3)

	var nsconfig map[string]*network.EndpointSettings
	nsconfig = make(map[string]*network.EndpointSettings)
	nsconfig["gophish"] = nil
	networkConfig := network.NetworkingConfig{EndpointsConfig: nsconfig}
	resp, err3 := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
	}, nil, &networkConfig, "")
	if err3 != nil {
		panic(err3)
	}

	if err3 := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err3 != nil {
		panic(err3)
	}

	fmt.Println("CONTAINERID:", resp.ID)

	container, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		panic(err)
	}
	endpoint := container.NetworkSettings.Networks["gophish"]
	ipAddr := endpoint.IPAddress
	fmt.Println("CONTAINER IP ADDRESS:", ipAddr)
	fmt.Println("FINISHED STARTING CONTAINER")
	// END START CONTAINER

	nginxConf = strings.Replace(nginxConf, "<site-name>", domain, -1)
	nginxConf = strings.Replace(nginxConf, "<ip-address>", ipAddr, -1)
	fileName := "/etc/nginx/sites-available/" + siteName + "." + "renfish.com"
	if err := ioutil.WriteFile(fileName, []byte(nginxConf), 0644); err != nil {
		auth.LogError(fmt.Sprintf("ERROR WRITING NGINX CONF FILE, sitename: %v, filename: %v, err: %v", siteName, fileName, err))
		return
	}

	// create certificate
	staging := ""
	fmt.Println("flagStaging:", *flagStaging)
	if *flagStaging {
		staging = "--staging"
		out, err := exec.Command("certbot", "certonly", "-n", "-q", "--standalone", "--agree-tos", "--email", "bjwbell@gmail.com", staging, "--pre-hook", "service nginx stop", "--post-hook", "service nginx start", "-d", domain).CombinedOutput()
		if err != nil {
			auth.LogError(fmt.Sprintf("CERTBOT ERROR, err: %v, stdout: %v", err, string(out)))
			log.Fatal(err)
		} else {
			fmt.Println("CREATED CERTBOT CERTIFICATE")
		}

	} else {
		out, err := exec.Command("certbot", "certonly", "-n", "-q", "--standalone", "--agree-tos", "--email", "bjwbell@gmail.com", "--pre-hook", "service nginx stop", "--post-hook", "service nginx start", "-d", domain).CombinedOutput()
		if err != nil {
			auth.LogError(fmt.Sprintf("CERTBOT ERROR, err: %v, stdout: %v", err, string(out)))
			log.Fatal(err)
		} else {
			fmt.Println("CREATED CERTBOT CERTIFICATE")
		}

	}

	// Link nginx conf file to sites-enabled/
	symlink := "/etc/nginx/sites-enabled/" + siteName + "." + "renfish.com"
	if err := os.Symlink(fileName, symlink); err != nil {
		auth.LogError(fmt.Sprintf("ERROR CREATING NGINX CONF FILE SYMLINK, sitename: %v, filename: %v, symlink: %v, err: %v", siteName, fileName, symlink, err))
		return
	} else {
		fmt.Println("CREATED NGINX CONF FILE")
	}

	// Reload nginx conf
	out, err := exec.Command("nginx", "-s", "reload").CombinedOutput()
	if err != nil {
		auth.LogError(fmt.Sprintf("ERROR RELOADING NGINX CONF, err: %v, stdout: %v", err, string(out)))
		log.Fatal(err)
	} else {
		fmt.Println("RELOADED NGINX CONF")
	}

	// Save details to database
	if _, success := db.SaveSite(emailAddress, siteName, ipAddr); !success {
		auth.LogError(fmt.Sprintf("ERROR SAVING SITE TO DB email (%s), sitename (%s), ip (%s)",
			emailAddress, siteName, ipAddr))
		log.Fatal(nil)
	} else {
		fmt.Println(fmt.Sprintf("SAVED SITE TO DB email (%s), sitename (%s), ip (%s)", emailAddress, siteName, ipAddr))
	}
	return
}

func createsiteHandler(w http.ResponseWriter, r *http.Request) {
	conf := struct {
		Conf     Configuration
		Email    string
		SiteName string
	}{Config(), "", ""}
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
		createSite(email, siteName)
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

var flagStaging = flag.Bool("staging", false, "Pass --staging to certbot")

func main() {
	flag.Parse()
	log.Println("Staging: ", *flagStaging)
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

	if !db.Exists(db.DbName) {
		db.Create(db.DbName)
	}

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
