package conf

import (
	"encoding/json"
	"log"
	"os"
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
	file, _ := os.Open("../conf.json")
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatal(err)
	}
	return configuration
}
