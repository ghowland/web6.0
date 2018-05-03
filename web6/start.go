package web6

import (
	"github.com/ghowland/yudien/yudien"
	"github.com/ghowland/yudien/yudiencore"
	"github.com/ghowland/yudien/yudiendata"
	"github.com/ghowland/yudien/yudienutil"
	. "github.com/ghowland/web6.0/config"
	//"github.com/ghowland/opsdb/opsdb"
	"flag"
)

func Start(pidFile *string) {
	// Vars for CLI arguments and flags
	config_path := ""
	log := ""

	// Process CLI arguments and flags
	flag.StringVar(&config_path, "config", ConfigFile, "Configuration file path (web6.json)")
	flag.StringVar(&log, "log", "", "Level for logging purposes")
	flag.StringVar(pidFile, "pid", "", "pid from command line")
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

	//go opsdb.RunJobWorkers()
}
