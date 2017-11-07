// Web6.0 is a dynamic integrated web server, that runs off a Database
package web6

/*

TODO:

- Make the accessors, that run off the last output
	- Terminate the function, so that accessors can start, using ".__access."
	- `__sql.dbselect.'SELECT * FROM table WHERE id = 5'.__.0.json_data_field.fieldname.10.anotherfieldname.etc`
- Change the quotes from single to double-quotes, so that we can write raw SQL commands, and still have quoting work in them
- `__query.1.__slice.-5,-1` - get the last 5 elements
- `__query.1.__sort.fieldname1.fieldname2` sort on multiple fieldnames

*/

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	. "github.com/ghowland/yudien/yudien"
	. "github.com/ghowland/yudien/yudiendata"
	. "github.com/ghowland/yudien/yudienutil"
	_ "github.com/lib/pq"
	"github.com/segmentio/ksuid"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
)

const (
	type_int          = iota
	type_float        = iota
	type_string       = iota
	type_string_force = iota // This forces it to a string, even if it will be ugly, will print the type of the non-string data too.  Testing this to see if splitting these into 2 yields better results.
	type_array        = iota // []interface{} - takes: lists, arrays, maps (key/value tuple array, strings (single element array), ints (single), floats (single)
	type_map          = iota // map[string]interface{}
)

// Core Web Page Handler.  All other routing occurs inside this function.
func Handler(w http.ResponseWriter, r *http.Request) {

	//url := fmt.Sprintf("%s", r.URL)

	url := r.URL.RequestURI()

	parts := strings.SplitN(url, "?", 2)

	uri := parts[0]

	relative_path := "./web/limitless5" + uri

	//log.Println("Testing path:", relative_path)

	is_static := false

	file, err := os.Open(relative_path)
	if err == nil {
		defer file.Close()

		file_info, err := file.Stat()
		if err != nil {
			log.Fatal(err)
		}

		// If this isnt a directory
		if !file_info.IsDir() {
			is_static = true

			size := file_info.Size()

			data := make([]byte, size)
			_, err := file.Read(data)
			if err != nil {
				log.Fatal(err)
			}

			if strings.HasSuffix(relative_path, ".css") {
				w.Header().Set("Content-Type", "text/css")
			} else if strings.HasSuffix(relative_path, ".js") {
				w.Header().Set("Content-Type", "text/javascript")
			} else if strings.HasSuffix(relative_path, ".jpg") {
				w.Header().Set("Content-Type", "image/jpg")
			} else if strings.HasSuffix(relative_path, ".png") {
				w.Header().Set("Content-Type", "image/png")
			} else if strings.HasSuffix(relative_path, ".woff2") {
				w.Header().Set("Content-Type", "font/woff2")
			} else {
				w.Header().Set("Content-Type", "text/html")
			}

			// Write the file into the body
			w.Write(data)
		}
	}

	// If this is not dynamic, then it's static
	if !is_static {
		// Handle all dynamic pages
		dynamicPage(uri, w, r)
	}
}

func dynamicPage(uri string, w http.ResponseWriter, r *http.Request) {
	// DB
	db, err := sql.Open("postgres", PgConnect)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// DB Web
	db_web, err := sql.Open("postgres", PgConnect)
	if err != nil {
		log.Fatal(err)
	}
	defer db_web.Close()

	web_site_id := 1

	//TODO(g): Get the web_site_domain from host header
	//web_site_domain_id := 1

	// Get the path to match from the DB
	sql := fmt.Sprintf("SELECT * FROM web_site WHERE _id = %d", web_site_id)
	web_site_result := Query(db_web, sql)
	if web_site_result == nil {
		panic("Failed to load website")
	}

	fmt.Printf("Type: %T\n\n", web_site_result)

	web_site_row := web_site_result[0]
	web_site := web_site_row

	fmt.Printf("\n\nGetting Web Site Page from URI: %s\n\n", uri)

	// Get the path to match from the DB
	sql = fmt.Sprintf("SELECT * FROM web_site_page WHERE web_site_id = %d AND name = '%s'", web_site_id, SanitizeSQL(uri))
	fmt.Printf("\n\nQuery: %s\n\n", sql)
	web_site_page_result := Query(db_web, sql)
	fmt.Printf("\n\nWeb Page Results: %v\n\n", web_site_page_result)

	// Check if this is a match for an API call
	found_api := false
	web_site_api_result := make([]map[string]interface{}, 0)
	if web_site["api_prefix_path"] == nil || strings.HasPrefix(uri, web_site["api_prefix_path"].(string)) {
		short_path := uri
		if web_site["api_prefix_path"] != nil {
			short_path = strings.Replace(uri, web_site["api_prefix_path"].(string), "", -1)
		}

		// Get the type of request if it exists (GET/POST/PUT/DELETE/etc)
		web_protocol_action := r.Method

		sql = fmt.Sprintf("SELECT _id FROM web_protocol_action WHERE name = '%s'", web_protocol_action)
		web_protocol_action_id := Query(db_web, sql)[0]["_id"]

		// Get the path to match from the DB - check for specific web protocol
		sql = fmt.Sprintf("SELECT * FROM web_site_api WHERE web_site_id = %d AND name = '%s' AND web_protocol_action_id = '%d'", web_site_id, SanitizeSQL(short_path), web_protocol_action_id)
		fmt.Printf("\n\nQuery: %s\n\n", sql)
		web_site_api_result = Query(db_web, sql)

		if len(web_site_api_result) > 0 {
			found_api = true
		} else {
			// Check if there is a general web_site_api entry without specified web protocol
			sql = fmt.Sprintf("SELECT * FROM web_site_api WHERE web_site_id = %d AND name = '%s' AND web_protocol_action_id IS NULL", web_site_id, SanitizeSQL(short_path))
			fmt.Printf("\n\nQuery: %s\n\n", sql)
			web_site_api_result = Query(db_web, sql)

			if len(web_site_api_result) > 0 {
				found_api = true
			}
		}
	}

	// If we found a matching page
	if found_api {
		fmt.Printf("\n\nFound API: %v\n\n", web_site_api_result[0])
		dynamicPage_API(db_web, db, web_site, web_site_api_result[0], uri, w, r)
	} else if len(web_site_page_result) > 0 {
		fmt.Printf("\n\nFound Dynamic Page: %v\n\n", web_site_page_result[0])
		dynamePage_RenderWidgets(db_web, db, web_site, web_site_page_result[0], uri, w, r)
	} else {
		fmt.Printf("\n\nPage not found: 404: %v\n\n", web_site_page_result)

		dynamicPage_404(uri, w, r)
	}
}

// Set up UDN data for an HTTP request
func GetStartingUdnData(db_web *sql.DB, db *sql.DB, web_site map[string]interface{}, web_site_page map[string]interface{}, uri string, web_protocol_action string, body io.Reader, param_map map[string][]string, header_map map[string][]string, cookie_array []*http.Cookie) map[string]interface{} {

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

	//TODO(g): Move this so we arent doing it every page load

	// Get the params: map[string]interface{}
	udn_data["param"] = make(map[string]interface{})

	for key, value := range param_map {
		//fmt.Printf("\n----KEY: %s  VALUE:  %s\n\n", key, value[0])
		//TODO(g): Decide what to do with the extra headers in the array later, we may not want to allow this ever, but thats not necessarily true.  Think about it, its certainly not the typical case, and isnt required
		udn_data["param"].(map[string]interface{})[key] = value[0]
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
		session_sql := fmt.Sprintf("SELECT * FROM web_user_session WHERE web_site_id = %d AND name = '%s'", web_site["_id"], SanitizeSQL(session_value.(string)))
		session_rows := Query(db_web, session_sql)
		if len(session_rows) == 1 {
			session := session_rows[0]
			user_id := session["user_id"]

			fmt.Printf("Found User ID: %d  Session: %v\n\n", user_id, session)

			// Load session from json_data
			target_map := make(map[string]interface{})
			if session["data_json"] != nil {
				err := json.Unmarshal([]byte(session["data_json"].(string)), &target_map)
				if err != nil {
					log.Panic(err)
				}
			}

			fmt.Printf("Session Data: %v\n\n", target_map)

			udn_data["session"] = target_map

			// Load the user data too
			user_sql := fmt.Sprintf("SELECT * FROM \"user\" WHERE _id = %d", user_id)
			user_rows := Query(db_web, user_sql)
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
			fmt.Printf("User Data: %v\n\n", target_map_user)

			udn_data["user_data"] = target_map_user
		}
	}

	// Get the UUID for this request
	id := ksuid.New()
	udn_data["uuid"] = id.String()

	return udn_data
}

// Set cookies against the HTTP Request
func SetCookies(cookie_map map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	for key, value := range cookie_map {
		//TODO(g):REMOVE: Testing only...
		new_cookie := http.Cookie{}
		new_cookie.Name = key
		new_cookie.Value = fmt.Sprintf("%v", value)
		new_cookie.Path = "/"
		http.SetCookie(w, &new_cookie)

		fmt.Printf("** Setting COOKIE: %s = %s\n", key, value)
	}
}

// Get the params of the HTTP request
func GetHTTPParams(r *http.Request) map[string][]string {

	// Check the web protocol action - for POST/PUT requests, params are found in the body
	param_map := make(map[string][]string)

	web_protocol_action := r.Method
	http_header := r.Header.Get("Content-Type")

	if web_protocol_action == "POST" || web_protocol_action == "PUT" {
		// Parse the body different depending on the type of the body (ex: JSON, form data, etc.)
		if http_header == "application/json" {
			// Read the body of the request (json)
			if body_bytes, err := ioutil.ReadAll(r.Body); err == nil {

				var data map[string]interface{}

				// Convert the bytestream of the body to JSON
				if err = json.Unmarshal(body_bytes, &data); err == nil {

					// param_map is map[string][]string -> need to convert var data from map[string]interface{} to map[string][]string
					for key, value := range data {
						if value_string, err := json.Marshal(value); err == nil {
							param_map[key] = []string{string(value_string)}
						}
					}
				}
			}
		} else if http_header == "application/x-www-form-urlencoded"{
			err := r.ParseForm()

			if err == nil {
				param_map = r.PostForm
			}
		}
	} else {  // GET and other requests
		param_map = r.URL.Query()
	}

	return param_map
}

func dynamicPage_API(db_web *sql.DB, db *sql.DB, web_site map[string]interface{}, web_site_api map[string]interface{}, uri string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Get UDN starting data values
	web_protocol_action := r.Method
	request_body := r.Body
	param_map := GetHTTPParams(r)
	header_map := r.Header
	cookie_array := r.Cookies()

	// Get our starting UDN data
	udn_data := GetStartingUdnData(db_web, db, web_site, web_site_api, uri, web_protocol_action, request_body, param_map, header_map, cookie_array)

	// Output params if logging is allowed
	if udn_data["web_site_page"].(map[string]interface{})["allow_logging"].(bool) {
		fmt.Printf("Starting UDN Data: %v\n\n", udn_data)

		fmt.Printf("Params: %v\n\n", param_map)
	}

	// Get the base widget
	sql := fmt.Sprintf("SELECT * FROM web_widget")
	all_widgets := Query(db_web, sql)

	// Save all our base web_widgets, so we can access them anytime we want
	udn_data["base_widget"] = MapArrayToMap(all_widgets, "name")

	// Get UDN schema per request
	//TODO(g): Dont do this every request
	udn_schema := PrepareSchemaUDN(db_web)

	// Make sure messages are output to screen and logged when it is allowed to do so
	udn_schema["allow_logging"] = udn_data["web_site_page"].(map[string]interface{})["allow_logging"].(bool)

	// If we are being told to debug, do so
	if param_map["__debug"] != nil {
		udn_schema["udn_debug"] = true
	} else if Debug_Udn_Api == true {
		// API calls are harder to change than web page requests, so made a separate in code var to toggle debugging
		udn_schema["udn_debug"] = true
	}

	// Process the UDN, which updates the pool at udn_data
	if web_site_api["udn_data_json"] != nil {
		ProcessSchemaUDNSet(db_web, udn_schema, web_site_api["udn_data_json"].(string), udn_data)
	} else {
		fmt.Printf("UDN Execution: API: %s: None\n\n", web_site_api["name"])
	}

	// Set Cookies
	SetCookies(udn_data["set_cookie"].(map[string]interface{}), w, r)

	// Write whatever is in the API result map, as a JSON result
	var buffer bytes.Buffer
	body, _ := json.Marshal(udn_data["set_api_result"])
	buffer.Write(body)

	fmt.Printf("Writing API body: %s\n\n", body)

	// Write out our output as HTML
	html_path := UdnDebugWriteHtml(udn_schema)

	if udn_schema["allow_logging"].(bool) {
		fmt.Printf("UDN Debug HTML Log: %s\n", html_path)
	}

	// Write out the final page
	w.Write([]byte(buffer.String()))

}

func dynamePage_RenderWidgets(db_web *sql.DB, db *sql.DB, web_site map[string]interface{}, web_site_page map[string]interface{}, uri string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	sql := fmt.Sprintf("SELECT * FROM web_site_page_widget WHERE web_site_page_id = %d ORDER BY priority ASC", web_site_page["_id"])
	web_site_page_widgets := Query(db_web, sql)

	// Get the base web site widget
	sql = fmt.Sprintf("SELECT * FROM web_site_page_widget WHERE _id = %d", web_site_page["base_page_web_site_page_widget_id"])
	base_page_widgets := Query(db_web, sql)

	// If we couldnt find the page, quit (404)
	if len(base_page_widgets) < 1 {
		fmt.Printf("No base page widgets found, going 404\n")

		dynamicPage_404(uri, w, r)
		return
	}

	base_page_widget := base_page_widgets[0]

	// Get the base widget
	sql = fmt.Sprintf("SELECT * FROM web_widget WHERE _id = %d", base_page_widget["web_widget_id"])
	base_widgets := Query(db_web, sql)

	base_page_html, err := ioutil.ReadFile(base_widgets[0]["path"].(string))
	if err != nil {
		log.Panic(err)
	}

	// Get UDN starting data values
	web_protocol_action := r.Method
	request_body := r.Body
	param_map := GetHTTPParams(r)
	header_map := r.Header
	cookie_array := r.Cookies()

	// Get our starting UDN data
	udn_data := GetStartingUdnData(db_web, db, web_site, web_site_page, uri, web_protocol_action, request_body, param_map, header_map, cookie_array)

	fmt.Printf("Starting UDN Data: %v\n\n", udn_data)

	// Get the base widget
	sql = fmt.Sprintf("SELECT * FROM web_widget")
	all_widgets := Query(db_web, sql)

	// Save all our base web_widgets, so we can access them anytime we want
	udn_data["base_widget"] = MapArrayToMap(all_widgets, "name")

	//fmt.Printf("Base Widget: base_list2_header: %v\n\n", udn_data["base_widget"].(map[string]interface{})["base_list2_header"])

	// We need to use this as a variable, so make it accessible to reduce casting
	page_map := udn_data["page"].(map[string]interface{})

	//TODO(g):HARDCODED: Im just forcing /login for now to make bootstrapping faster, it can come from the data source, think about it
	if uri != "/login" {
		if udn_data["user"].(map[string]interface{})["_id"] == nil {
			login_page_id := web_site["login_web_site_page_id"].(int64)
			login_page_sql := fmt.Sprintf("SELECT * FROM web_site_page WHERE _id = %d", login_page_id)
			login_page_rows := Query(db_web, login_page_sql)
			if len(login_page_rows) >= 1 {
				login_page := login_page_rows[0]

				// Render the Login Page
				//TODO(g): Verify we can only ever recurse once, this is the only time I do this, so far.  Think out whether this is a good idea...
				dynamePage_RenderWidgets(db_web, db, web_site, login_page, "/login", w, r)

				// Return, as the Login page has been rendered, so we abandon rendering the requested page
				return
			}
		}

	}

	// Get UDN schema per request
	//TODO(g): Dont do this every request
	udn_schema := PrepareSchemaUDN(db_web)

	// If we are being told to debug, do so
	if param_map["__debug"] != nil {
		udn_schema["udn_debug"] = true
	}

	// Loop over the page widgets, and template them
	for _, site_page_widget := range web_site_page_widgets {
		// Skip it if this is the base page, because we
		if site_page_widget["_id"] == web_site_page["base_page_web_site_page_widget_id"] {
			continue
		}

		// Put the Site Page Widget into the UDN Data, so we can operate on it
		udn_data["page_widget"] = site_page_widget

		widget_map := make(map[string]interface{})

		// Put the widget map into the UDN Data too
		udn_data["widget_map"] = widget_map

		// web_widget_id rendering widget -- single widget rendering
		var page_widget map[string]interface{}

		// Get any static content associated with this page widget.  Then we dont need to worry about quoting or other stuff
		widget_static := make(map[string]interface{})
		udn_data["widget_static"] = widget_static
		if site_page_widget["static_data_json"] != nil {
			err = json.Unmarshal([]byte(site_page_widget["static_data_json"].(string)), &widget_static)
			if err != nil {
				log.Panic(err)
			}
		}

		// If we have web_widget specified, use it
		if site_page_widget["web_widget_id"] != nil {

			// Get the base widget
			sql = fmt.Sprintf("SELECT * FROM web_widget WHERE _id = %d", site_page_widget["web_widget_id"])
			page_widgets := Query(db_web, sql)
			page_widget = page_widgets[0]

			fmt.Printf("Page Widget: %s: %s\n", site_page_widget["name"], page_widget["name"])

			// wigdet_map has all the UDN operations we will be using to embed child-widgets into this widget
			//TODO(g): We need to use the page_map data here too, because we need to template in the sub-widgets.  Think about this after testing it as-is...
			err = json.Unmarshal([]byte(site_page_widget["data_json"].(string)), &widget_map)
			if err != nil {
				log.Panic(err)
			}

			udn_data["web_widget"] = page_widget

			// Processing UDN: which updates the data pool at udn_data
			if site_page_widget["udn_data_json"] != nil {
				ProcessSchemaUDNSet(db_web, udn_schema, site_page_widget["udn_data_json"].(string), udn_data)
			} else {
				fmt.Printf("UDN Execution: %s: None\n\n", site_page_widget["name"])
			}

			// Process the Widget's Rendering UDN statements (singles)
			for widget_key, widget_value := range widget_map {
				//fmt.Printf("\n\nWidget Key: %s:  Value: %v\n\n", widget_key, widget_value)

				// Force the UDN string into a string
				//TODO(g): Not the best way to do this, fix later, doing now for dev speed/simplicity
				widget_udn_string := []string{fmt.Sprintf("%v", widget_value)}

				// Process the UDN with our new method.  Only uses Source, as we are getting, but not setting in this phase
				widget_udn_result := ProcessUDN(db, udn_schema, widget_udn_string, udn_data)

				widget_map[widget_key] = fmt.Sprintf("%v", GetResult(widget_udn_result, type_string))

				//fmt.Printf("Widget Key Result: %s   Result: %s\n\n", widget_key, SnippetData(widget_map[widget_key], 600))
			}

			//fmt.Printf("Title: %s\n", widget_map.Map["title"])

			item_html, err := ioutil.ReadFile(page_widget["path"].(string))
			if err != nil {
				log.Panic(err)
			}

			//TODO(g): Replace reading from the "path" above with the "html" stored in the DB, so it can be edited and displayed live
			//item_html := page_widget.Map["html"].(string)

			//fmt.Printf("Page Widget: %s   HTML: %s\n", page_widget["name"], SnippetData(page_widget["html"], 600))

			item_template := template.Must(template.New("text").Parse(string(item_html)))

			widget_map_template := NewTextTemplateMap()
			widget_map_template.Map = widget_map

			//fmt.Printf("  Templating with data: %v\n\n", SnippetData(widget_map, 600))

			item := StringFile{}
			err = item_template.Execute(&item, widget_map_template)
			if err != nil {
				log.Fatal(err)
			}

			// Append to our total forum_list_string
			key := site_page_widget["name"]

			//fmt.Printf("====== Finalized Template: %s\n%s\n\n", key, item.String)

			//fmt.Printf("=-=-=-=-= UDN Data: Output:\n%v\n\n", udn_data["output"])

			page_map[key.(string)] = item.String

		} else if site_page_widget["web_widget_instance_id"] != nil {
			// Render the Widget Instance
			udn_update_map := make(map[string]interface{})
			RenderWidgetInstance(db_web, udn_schema, udn_data, site_page_widget, udn_update_map)

		} else if site_page_widget["web_data_widget_instance_id"] != nil {
			// Render the Widget Instance, from the web_data_widget_instance
			udn_update_map := make(map[string]interface{})
			RenderWidgetInstance(db_web, udn_schema, udn_data, site_page_widget, udn_update_map)

		} else {
			panic("No web_widget_id, web_widget_instance_id, web_data_widget_instance_id.  Site Page Widgets need at least one of these.")
		}

	}

	// Get base page widget items.  These were also processed above, as the base_page_widget was included with the page...
	base_page_widget_map := NewTextTemplateMap()
	err = json.Unmarshal([]byte(base_page_widget["data_json"].(string)), &base_page_widget_map.Map)
	if err != nil {
		log.Panic(err)
	}

	// Add base_page_widget entries to page_map, if they dont already exist
	for key, value := range base_page_widget_map.Map {
		if _, ok := page_map[key]; ok {
			// Pass, already has this value
		} else {
			value_str := []string{fmt.Sprintf("%v", value)}

			// Process the UDN with our new method.  Only uses Source, as we are getting, but not setting in this phase
			widget_udn_result := ProcessUDN(db, udn_schema, value_str, udn_data)

			if widget_udn_result != nil {
				page_map[key] = fmt.Sprintf("%v", GetResult(widget_udn_result, type_string))
			} else {
				// Use the base page widget, without any processing, because we got back nil
				page_map[key] = value_str
			}

			//// Set the value, static text
			//page_map[key] = value
		}
	}

	fmt.Println("Rendering base page")

	// Put them into the base page
	base_page_template := template.Must(template.New("text").Parse(string(base_page_html)))

	// Set up the TextTemplateMap for page_map, now that it is map[string]interface{}
	page_map_text_template_map := NewTextTemplateMap()
	page_map_text_template_map.Map = page_map

	// Write the base page
	base_page := StringFile{}
	err = base_page_template.Execute(&base_page, page_map_text_template_map)
	if err != nil {
		log.Fatal(err)
	}

	// Set Cookies
	SetCookies(udn_data["set_cookie"].(map[string]interface{}), w, r)

	// Write out our output as HTML
	html_path := UdnDebugWriteHtml(udn_schema)
	fmt.Printf("UDN Debug HTML Log: %s\n", html_path)

	// Write out the final page
	w.Write([]byte(base_page.String))

}

func dynamicPage_404(uri string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	base_html, err := ioutil.ReadFile("web/limitless5/error_404.html")
	if err != nil {
		log.Panic(err)
	}

	w.Write([]byte(base_html))
}
