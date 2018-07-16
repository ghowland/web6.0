package web6

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
	"testing"

	"github.com/ghowland/yudien/yudien"
	"github.com/ghowland/yudien/yudiendata"
)

func TestUDN(t *testing.T) {
	LoadConfig("")
	// Set the proper database schema path
	Config.DefaultDatabase.Schema = "../data/schema.json"
	yudien.Configure(&Config.DefaultDatabase, Config.Databases, &Config.Logging, &Config.Authentication)

	// DB
	db, err := sql.Open("postgres", "user=postgres dbname=opsdb password='password' host=localhost sslmode=disable")
	if err != nil {
		t.Fatalf("Cannot connect to DB: %s", err)
	}
	defer db.Close()

	web_site_id := int64(1) // Default is 1

	sql := fmt.Sprintf("SELECT * FROM web_site WHERE _id = %d", web_site_id)
	web_site_result := yudiendata.Query(db, sql)
	if web_site_result == nil || len(web_site_result) == 0 {
		t.Fatalf("Cannot find main web_site: %s", err)
	}

	web_site_row := web_site_result[0]
	web_site := web_site_row
	request_body := bytes.NewReader([]byte(""))
	header_map := make(map[string][]string)

	// Get starting UDN data
	udn_data := GetStartingUdnData(db, db, web_site, make(map[string]interface{}), "", "", request_body, make(map[string]interface{}), header_map, nil)

	// Test the UDN Processor
	udn_schema := yudien.PrepareSchemaUDN(db)
	//fmt.Printf("\n\nUDN Schema: %v\n\n", udn_schema)

	// Fetch test cases from db
	sql = "SELECT _id, name FROM test ORDER BY name;"

	result := yudiendata.Query(db, sql)

	for _, test := range result {
		t.Run(test["name"].(string), func(t *testing.T) {
			executed_udn := fmt.Sprintf("[[[\"__function.run_test.%s\"]]]", strconv.FormatInt(test["_id"].(int64), 10))

			result := yudien.ProcessSchemaUDNSet(db, udn_schema, executed_udn, udn_data)

			// Error if result does not match expected
			if result.(map[string]interface{})["success"] != "1" {
				t.Errorf("Mismatch of actual and expected result.\nExpected Result: %s\nActual Result: %s\nExpected Test Memory: %s\nActual Test Memory: %s",
					result.(map[string]interface{})["result"].(map[string]interface{})["expected_result"],
					result.(map[string]interface{})["result"].(map[string]interface{})["actual_result"],
					result.(map[string]interface{})["result"].(map[string]interface{})["expected_ending_memory"],
					result.(map[string]interface{})["result"].(map[string]interface{})["actual_ending_memory"])
			}

		})
	}
}
