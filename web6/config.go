package web6

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"

	"github.com/ghowland/yudien/yudien"
	"github.com/ghowland/yudien/yudiencore"
	"github.com/ghowland/yudien/yudiendata"
	"github.com/ghowland/yudien/yudienutil"
	"flag"
)

// This is the location of the Production configuration file.  The development file is in ~/secure/web6.json
const configFile = "/etc/web6/web6.json"

type Web6Config struct {
	//Ldap  yudien.LdapConfig  `json:"ldap"`
	DefaultDatabase yudiendata.DatabaseConfig            `json:"default_database"`
	Databases       map[string]yudiendata.DatabaseConfig `json:"databases"`
	LdapOverride    yudiendata.DatabaseConfig            `json:"ldap_override"`
	Authentication  yudien.AuthenticationConfig          `json:"authentication"`
	Logging         yudien.LoggingConfig                 `json:"logging"`
}

var Config *Web6Config = &Web6Config{}

func Start() {
	// Vars for CLI arguments and flags
	config_path := ""
	log := ""
	pid := "" // For avoiding error in prod as pid flag is passed so we need an empty catcher

	// Process CLI arguments and flags
	flag.StringVar(&config_path, "config", configFile, "Configuration file path (web6.json)")
	flag.StringVar(&log, "log", "", "Level for logging purposes")
	flag.StringVar(&pid, "pid", "", "For avoiding error with prod - leave empty")
	flag.Parse()

	LoadConfig(config_path)

	// If logging is specified in the flag, then override the config file
	if log != "" {
		Config.Logging.Level = log
	}

	yudien.Configure(&Config.DefaultDatabase, Config.Databases, &Config.Logging, &Config.Authentication)

	if false {
		yudiendata.ImportSchemaJson("data/schema.json")
		yudiendata.GenerateSchemaJson("data/schema_out.json")

		// Test data in same format (ordering/sorting)
		text := yudienutil.ReadPathData("data/schema.json")
		data_str, _ := yudienutil.JsonLoadMap(text)
		data := yudienutil.JsonDump(data_str)
		yudienutil.WritePathData("data/schema_in.json", data)

		yudiencore.UdnLogLevel(nil, log_info, "\n\nEnsure DB\n\n")

		yudiendata.DatamanEnsureDatabases(yudiendata.DefaultDatabase.ConnectOptions, yudiendata.DefaultDatabase.Database, yudiendata.DefaultDatabase.Schema, "data/schema_out.json")

	}

	yudiencore.UdnLogLevel(nil, log_info, "Finished starting...\n")

	//go RunJobWorkers()
}

func LoadConfig(config string) {
	_, err := os.Stat(config)
	if err != nil {
		usr, _ := user.Current()
		homedir := usr.HomeDir
		// This is the developer version.  The configFile is the production version
		config = fmt.Sprintf("%s/secure/web6.json", homedir)
		_, err := os.Stat(config)
		if err != nil {
			panic(fmt.Sprintf("Cound not find web6.json in /etc/web6 or %s", config))
		}
	}

	yudiencore.UdnLogLevel(nil, log_info, "Found web6 config at %s\n", config)

	config_str, err := ioutil.ReadFile(config)
	if err != nil {
		log.Panic(fmt.Sprintf("Cannot read config file: %s: %s\n", config, err.Error()))
	}
	err = json.Unmarshal([]byte(config_str), Config)
	if err != nil {
		log.Panic(fmt.Sprintf("Cannot parse JSON config file: %s: %s\n", config, err.Error()))
	}
}
