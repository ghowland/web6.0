package web6

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
	"testing"

	. "github.com/ghowland/yudien/yudien"
	. "github.com/ghowland/yudien/yudiendata"
)

func BenchmarkUDN(b *testing.B) {
	LoadConfig("")
	// Set the proper database schema path
	Config.DefaultDatabase.Schema = "../data/schema.json"
	Configure(&Config.DefaultDatabase, Config.Databases, &Config.Logging, &Config.Authentication)

	// DB
	db, err := sql.Open("postgres", "user=postgres dbname=opsdb password='password' host=localhost sslmode=disable")
	if err != nil {
		b.Fatalf("Cannot connect to DB: %s", err)
	}
	defer db.Close()

	web_site_id := int64(1) // Default is 1

	sql := fmt.Sprintf("SELECT * FROM web_site WHERE _id = %d", web_site_id)
	web_site_result := Query(db, sql)
	if web_site_result == nil || len(web_site_result) == 0 {
		b.Fatalf("Cannot find main web_site: %s", err)
	}

	web_site_row := web_site_result[0]
	web_site := web_site_row
	request_body := bytes.NewReader([]byte(""))
	header_map := make(map[string][]string)

	// Get starting UDN data
	udn_data := GetStartingUdnData(db, db, web_site, make(map[string]interface{}), "", "", request_body, make(map[string]interface{}), header_map, nil)

	// Test the UDN Processor
	udn_schema := PrepareSchemaUDN(db)
	//fmt.Printf("\n\nUDN Schema: %v\n\n", udn_schema)

	// Fetch test cases from db
	sql = "SELECT * FROM test ORDER BY name;"

	result := Query(db, sql)

	for _, test := range result {
		b.Run(test["name"].(string), func(b *testing.B) {
			b.StopTimer() // don't measure initialization
			// Fetch the test udn_data_json and starting memory from the DB for the benchmark
			executed_udn := fmt.Sprintf("[[[\"__data_filter.test.{_id=['=',%s]}.__get_index.0.__set_temp.result\","+
				"\"__get_temp.result.udn_data_json.__set.test_udn\","+
				"\"__get_temp.result.starting_test_memory_data_json.__set.test_start_memory\"]]]", strconv.FormatInt(test["_id"].(int64), 10))

			ProcessSchemaUDNSet(db, udn_schema, executed_udn, udn_data)

			// Json decode the starting memory if needed
			executed_udn = fmt.Sprintf("[[[\"__if.(__get.test_start_memory).__get.test_start_memory.__json_decode.__set.test_start_memory.__else.__get_temp.null.__set.test_start_memory.__end_if\"]]]")
			ProcessSchemaUDNSet(db, udn_schema, executed_udn, udn_data)

			b.StartTimer()

			/*
				Note that the entire memory is not reset after each benchmark test and this could
				cause inaccuracies in certain benchmarks (unlikely). However, we do not care about the correctness of
				the result and resetting the entire udn memory slows down the benchmarking significantly.
				Tests should be unit tests and not reliant on existing memory results
			*/
			for n := 0; n < b.N; n++ {
				// Get starting memory for current benchmark run and execute the udn
				executed_udn = fmt.Sprintf("[[[\"__get.test_start_memory.__set.test.__execute.(__get.test_udn)\"]]]")

				ProcessSchemaUDNSet(db, udn_schema, executed_udn, udn_data)
			}
		})
	}
}
