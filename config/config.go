package config

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

const ( // order matters for log levels
	log_off   = iota
	log_error = iota
	log_warn  = iota
	log_info  = iota
	log_debug = iota
	log_trace = iota
)

// This is the location of the Production configuration file.  The development file is in ~/secure/web6.json
const ConfigFile = "/etc/web6/web6.json"

type Web6Config struct {
	//Ldap  yudien.LdapConfig  `json:"ldap"`
	DefaultDatabase yudiendata.DatabaseConfig            `json:"default_database"`
	Databases       map[string]yudiendata.DatabaseConfig `json:"databases"`
	LdapOverride    yudiendata.DatabaseConfig            `json:"ldap_override"`
	Authentication  yudien.AuthenticationConfig          `json:"authentication"`
	Logging         yudien.LoggingConfig                 `json:"logging"`
}

var Config Web6Config

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

	fmt.Printf("Load web6 config: %s\n\n", config)

	config_str, err := ioutil.ReadFile(config)
	if err != nil {
		log.Panic(fmt.Sprintf("Cannot read config file: %s: %s\n", config, err.Error()))
	}
	err = json.Unmarshal([]byte(config_str), &Config)
	if err != nil {
		log.Panic(fmt.Sprintf("Cannot parse JSON config file: %s: %s\n", config, err.Error()))
	}

	// Format the ConnectOptions based on all our data, for the default database
	Config.DefaultDatabase.ConnectOptions = yudienutil.TemplateInterface(Config.DefaultDatabase, Config.DefaultDatabase.ConnectOptionsTemplate)
	fmt.Printf("Database Connect String: %s: %s\n\n", Config.DefaultDatabase.Name, Config.DefaultDatabase.ConnectOptions)

	// Format the ConnectOptions based on all our data, for all secondary databases
	for db_key, db_config := range Config.Databases {
		// Cannot modify structs behind maps because it's not a normal point, as it's behind a hashmap, so copying, changing and copying back
		real_db_config := Config.Databases[db_key]
		real_db_config.ConnectOptions = yudienutil.TemplateInterface(db_config, db_config.ConnectOptionsTemplate)
		Config.Databases[db_key] = real_db_config
	}
}
