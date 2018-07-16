package web6

import (
	"flag"

	"github.com/ghowland/web6.0/config"
	"github.com/ghowland/yudien/yudien"
	"github.com/ghowland/yudien/yudiencore"
	"github.com/ghowland/yudien/yudiendata"
	"github.com/ghowland/yudien/yudienutil"
)

func Start(pidFile *string) {
	// Vars for CLI arguments and flags
	config_path := ""
	log := ""

	// Process CLI arguments and flags
	flag.StringVar(&config_path, "config", config.ConfigFile, "Configuration file path (web6.json)")
	flag.StringVar(&log, "log", "", "Level for logging purposes")
	flag.StringVar(pidFile, "pid", "", "Specify PID path")
	var import_database = flag.Bool("import-database", false, "Import databases into schema_table")
	flag.Parse()

	config.LoadConfig(config_path)

	//fmt.Printf("Configure post load: \n%s\n", yudienutil.JsonDump(Config))

	// If logging is specified in the flag, then override the config file
	if log != "" {
		config.Config.Logging.Level = log
	}

	yudien.Configure(&config.Config.DefaultDatabase, config.Config.Databases, &config.Config.Logging, &config.Config.Authentication)

	//TODO(g): Make this a CLI flag
	if import_database != nil && *import_database == true {
		// Import the default database
		ImportDatabase(yudiendata.DefaultDatabase)

		// Import all the other databases
		for _, db_config := range config.Config.Databases {
			ImportDatabase(&db_config)
		}
	}

	yudiencore.UdnLogLevel(nil, log_info, "Finished starting...\n")

	//go opsdb.RunJobWorkers()
}

func ImportDatabase(database_config *yudiendata.DatabaseConfig) {
	yudiencore.UdnLogLevel(nil, log_info, "\n\nEnsure Database: %s: %s\n\n", database_config.Name, yudienutil.JsonDump(database_config))

	yudiendata.ImportSchemaJson(database_config.Schema)

	//yudiendata.DatamanEnsureDatabases(*database_config, nil)
}

func ExportDatabase(path_out string, path_in_compare interface{}) {
	yudiendata.GenerateSchemaJson("data/schema_out.json")

	if path_in_compare != nil {
		// Test data in same format (ordering/sorting)
		text := yudienutil.ReadPathData(path_in_compare.(string))
		data_str, _ := yudienutil.JsonLoadMap(text)
		data := yudienutil.JsonDump(data_str)
		yudienutil.WritePathData("data/schema_out_compare.json", data)
	}
}
