package web6

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"

	"github.com/ghowland/yudien/yudien"
)

const configFile = "/etc/web6/web6.json"

type Web6Config struct {
	Ldap  yudien.LdapConfig  `json:"ldap"`
	Opsdb yudien.OpsdbConfig `json:"opsdb"`
}

var Config *Web6Config = &Web6Config{}

func LoadConfig() {
	config := configFile
	_, err := os.Stat(configFile)
	if err != nil {
		usr, _ := user.Current()
		homedir := usr.HomeDir
		config = fmt.Sprintf("%s/secure/web6.json", homedir)
		_, err := os.Stat(config)
		if err != nil {
			panic("Could not find web6.json in /etc/web or ~/secure")
		}
	}

	fmt.Printf("Found web6 config at %s\n", config)

	config_str, err := ioutil.ReadFile(config)
	if err != nil {
		log.Panic(err)
	}
	err = json.Unmarshal([]byte(config_str), Config)
	if err != nil {
		log.Panic(err)
	}
}
