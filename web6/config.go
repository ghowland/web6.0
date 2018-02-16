package web6

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"

	"github.com/ghowland/yudien/yudien"
	"github.com/ghowland/yudien/yudiendata"
	"github.com/ghowland/yudien/yudienutil"
)

const configFile = "/etc/web6/web6.json"

type Web6Config struct {
	Ldap  yudien.LdapConfig  `json:"ldap"`
	Opsdb yudiendata.OpsdbConfig `json:"opsdb"`
}

var Config *Web6Config = &Web6Config{}

func Start() {
	LoadConfig()

	yudien.Configure(&Config.Ldap, &Config.Opsdb)

	if false {
		yudiendata.ImportSchemaJson("data/schema.json")
		yudiendata.GenerateSchemaJson("data/schema_out.json")

		// Test data in same format (ordering/sorting)
		text := yudieutil.ReadPathData("data/schema.json")
		data_str, _ := yudieutil.JsonLoadMap(text)
		data  := yudieutil.JsonDump(data_str)
		yudieutil.WritePathData("data/schema_in.json", data)

		fmt.Printf("\n\nEnsure DB\n\n")

		yudiendata.DatamanEnsureDatabases(yudiendata.Opsdb.ConnectOptions, yudiendata.Opsdb.Database, "data/schema.json", "data/schema_out.json")

	}

	fmt.Printf("Finished starting...\n")

	//go RunJobWorkers()
}

func LoadConfig() {
	config := configFile
	_, err := os.Stat(configFile)
	if err != nil {
		usr, _ := user.Current()
		homedir := usr.HomeDir
		config = fmt.Sprintf("%s/secure/web6.json", homedir)
		_, err := os.Stat(config)
		if err != nil {
			panic(fmt.Sprintf("Cound not find web6.json in /etc/web6 or %s", config))
		}
	}

	fmt.Printf("Found web6 config at %s\n", config)

	config_str, err := ioutil.ReadFile(config)
	if err != nil {
		log.Panic(fmt.Sprintf("Cannot read config file: %s: %s\n", config, err.Error()))
	}
	err = json.Unmarshal([]byte(config_str), Config)
	if err != nil {
		log.Panic(fmt.Sprintf("Cannot parse JSON config file: %s: %s\n", config, err.Error()))
	}
}
