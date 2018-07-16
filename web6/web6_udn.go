package web6

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/ghowland/yudien/yudiencore"
	"github.com/ghowland/yudien/yudiendata"
	"github.com/segmentio/ksuid"
)

// Set up UDN data for an HTTP request
func GetStartingUdnData(db_web *sql.DB, db *sql.DB, web_site map[string]interface{}, web_site_page map[string]interface{}, uri string, web_protocol_action string, body io.Reader, param_map map[string]interface{}, header_map map[string][]string, cookie_array []*http.Cookie) map[string]interface{} {

	// Data pool for UDN
	udn_data := make(map[string]interface{})

	// Prepare the udn_data with it's fixed pools of data
	//udn_data["widget"] = *NewTextTemplateMap()
	udn_data["web_protocol_action"] = web_protocol_action
	udn_data["data"] = make(map[string]interface{})
	udn_data["temp"] = make(map[string]interface{})
	udn_data["output"] = make(map[string]interface{}) // Staging output goes here, can share them with appending as well.
	//TODO(g): Make args accessible at the start of every ExecuteUdnPart after getting the args!
	udn_data["arg"] = make(map[string]interface{})          // Every function call blows this away, and sets the args in it's data, so it's accessable
	udn_data["function_arg"] = make(map[string]interface{}) // Function arguments, from Stored UDN Function __function, sets the incoming function args
	udn_data["page"] = make(map[string]interface{})         //TODO(g):NAMING: __widget is access here, and not from "widget", this can be changed, since thats what it is...

	udn_data["set_api_result"] = make(map[string]interface{})   // If this is an API call, set values in here, which will be encoded in JSON and sent back to the client on return
	udn_data["set_cookie"] = make(map[string]interface{})       // Set Cookies.  Any data set in here goes into a cookie.  Will use standard expiration and domain for now.
	udn_data["set_header"] = make(map[string]interface{})       // Set HTTP Headers.
	udn_data["set_http_options"] = make(map[string]interface{}) // Any other things we want to control from UDN, we put in here to be processed.  Can be anything, not based on a specific standard.
	udn_data["http_response_code"] = 200                        // Default

	//TODO(g): Move this so we arent doing it every page load

	// Get the params: map[string]interface{}
	udn_data["param"] = make(map[string]interface{})

	for key, value := range param_map {
		//fmt.Printf("\n----KEY: %s  VALUE:  %s\n\n", key, value[0])
		//TODO(g): Decide what to do with the extra headers in the array later, we may not want to allow this ever, but thats not necessarily true.  Think about it, its certainly not the typical case, and isnt required
		udn_data["param"].(map[string]interface{})[key] = value
	}

	// Get the JSON Body, if it exists, from an API-style call in
	udn_data["api_input"] = make(map[string]interface{})
	json_body := make(map[string]interface{})
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&json_body)
	// If we got it, then add all the keys to api_input
	if err == nil {
		for body_key, body_value := range json_body {
			udn_data["api_input"].(map[string]interface{})[body_key] = body_value
		}
	}

	// Get the cookies: map[string]interface{}
	udn_data["cookie"] = make(map[string]interface{})
	for _, cookie := range cookie_array {
		udn_data["cookie"].(map[string]interface{})[cookie.Name] = cookie.Value
	}

	// Get the headers: map[string]interface{}
	udn_data["header"] = make(map[string]interface{})
	for header_key, header_value_array := range header_map {
		//TODO(g): Decide what to do with the extra headers in the array later, these will be useful and are necessary to be correct
		udn_data["header"].(map[string]interface{})[header_key] = header_value_array[0]
	}

	// Verify that this user is logged in, render the login page, if they arent logged in
	udn_data["session"] = make(map[string]interface{})
	udn_data["user"] = make(map[string]interface{})
	udn_data["user_data"] = make(map[string]interface{})
	udn_data["web_site"] = web_site
	udn_data["web_site_page"] = web_site_page
	if session_value, ok := udn_data["cookie"].(map[string]interface{})["opsdb_session"]; ok {
		session_sql := fmt.Sprintf("SELECT * FROM web_user_session WHERE web_site_id = %d AND name = '%s'", web_site["_id"], yudiendata.SanitizeSQL(session_value.(string)))
		session_rows := yudiendata.Query(db_web, session_sql)
		if len(session_rows) == 1 {
			session := session_rows[0]
			user_id := session["user_id"]

			yudiencore.UdnLogLevel(nil, log_info, "Found User ID: %d  Session: %v\n\n", user_id, session)

			// Load session from json_data
			target_map := make(map[string]interface{})
			if session["data_json"] != nil {
				err := json.Unmarshal([]byte(session["data_json"].(string)), &target_map)
				if err != nil {
					log.Panic(err)
				}
			}

			yudiencore.UdnLogLevel(nil, log_debug, "Session Data: %v\n\n", target_map)

			udn_data["session"] = target_map

			// Load the user data too
			user_sql := fmt.Sprintf("SELECT * FROM \"user\" WHERE _id = %d", user_id)
			user_rows := yudiendata.Query(db_web, user_sql)
			target_map_user := make(map[string]interface{})
			if len(user_rows) == 1 {
				// Set the user here
				udn_data["user"] = user_rows[0]

				// Load from user data from json_data
				if user_rows[0]["data_json"] != nil {
					err := json.Unmarshal([]byte(user_rows[0]["data_json"].(string)), &target_map_user)
					if err != nil {
						log.Panic(err)
					}
				}
			}
			yudiencore.UdnLogLevel(nil, log_debug, "User Data: %v\n\n", target_map_user)

			udn_data["user_data"] = target_map_user
		}
	}

	// Get the UUID for this request
	id := ksuid.New()
	udn_data["uuid"] = id.String()

	return udn_data
}
