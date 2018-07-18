package web6

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"
	"text/template"

	"github.com/ghowland/web6.0/config"
	"github.com/ghowland/yudien/yudien"
	"github.com/ghowland/yudien/yudiencore"
	"github.com/ghowland/yudien/yudiendata"
	"github.com/ghowland/yudien/yudienutil"
	_ "github.com/lib/pq"
)

func dynamicPage(uri string, w http.ResponseWriter, r *http.Request) {
	//yudiencore.UdnLogLevel(nil, log_info, "Connecting to: %s\n", Config.DefaultDatabase.ConnectOptions)

	// DB
	db, err := sql.Open("postgres", config.Config.DefaultDatabase.ConnectOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// DB Web
	db_web, err := sql.Open("postgres", config.Config.DefaultDatabase.ConnectOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer db_web.Close()

	web_site_id := int64(1) // Default is 1

	// Get the web_site_domain from host header
	host := r.Host

	// Get the domain - splice off the suffix (port_number) or prefix (http://) if present
	host_parts := strings.Split(host, "http://") // split off http:// prefix if it exists
	host = host_parts[len(host_parts)-1]
	host_parts = strings.Split(host, ":") // split off port number if it exists
	host = host_parts[0]

	// Match the domain name to an entry in web_site_domain
	sql := fmt.Sprintf("SELECT * FROM web_site_domain WHERE name = '%s'", host)
	web_site_host_result := yudiendata.Query(db_web, sql)

	if web_site_host_result == nil || len(web_site_host_result) == 0 {
		yudiencore.UdnLogLevel(nil, log_error, "Failed to load website domain: %d\n", host)
		// Rendor 404
		dynamicPage_404(uri, w, r)
		return
	} else {
		// set the web_site_id from the web_site_domain table
		web_site_id = web_site_host_result[0]["web_site_id"].(int64)
	}

	// Get the path to match from the DB
	sql = fmt.Sprintf("SELECT * FROM web_site WHERE _id = %d", web_site_id)
	web_site_result := yudiendata.Query(db_web, sql)
	if web_site_result == nil || len(web_site_result) == 0 {
		yudiencore.UdnLogLevel(nil, log_error, "Failed to load website: %d\n", web_site_id)
		// Rendor 404
		dynamicPage_404(uri, w, r)
		return
	}

	yudiencore.UdnLogLevel(nil, log_debug, "Type: %T\n\n", web_site_result)

	web_site_row := web_site_result[0]
	web_site := web_site_row

	yudiencore.UdnLogLevel(nil, log_info, "\n\nGetting Web Site Page from URI: %s\n\n", uri)

	// Get the path to match from the DB
	sql = fmt.Sprintf("SELECT * FROM web_site_page WHERE web_site_id = %d AND name = '%s'", web_site_id, yudiendata.SanitizeSQL(uri))
	web_site_page_result := yudiendata.Query(db_web, sql)
	yudiencore.UdnLogLevel(nil, log_info, "\n\nWeb Page Results: %v\n\n", web_site_page_result)

	// Check if this is a match for an API call
	found_api := false
	web_site_api_result := make([]map[string]interface{}, 0)
	web_site_api_entry := make(map[string]interface{})

	if web_site["api_prefix_path"] == nil || strings.HasPrefix(uri, web_site["api_prefix_path"].(string)) {
		short_path := uri
		if web_site["api_prefix_path"] != nil {
			short_path = strings.Replace(uri, web_site["api_prefix_path"].(string), "", -1)
		}

		// Get the type of request if it exists (GET/POST/PUT/DELETE/etc)
		web_protocol_action := r.Method

		sql = fmt.Sprintf("SELECT _id FROM web_protocol_action WHERE name = '%s'", web_protocol_action)
		web_protocol_action_id := yudiendata.Query(db_web, sql)[0]["_id"]

		// Track the path name to provide use in possible regex in the path name
		name := yudiendata.SanitizeSQL(short_path)

		// Get the path to match from the DB - check for specific web protocol
		sql = fmt.Sprintf("SELECT * FROM web_site_api WHERE web_site_id = %d AND name = '%s' AND (web_protocol_action_id = '%d' OR web_protocol_action_id IS NULL)", web_site_id, name, web_protocol_action_id)
		web_site_api_result = yudiendata.Query(db_web, sql)

		if len(web_site_api_result) > 0 {
			found_api = true
			web_site_api_entry = web_site_api_result[0]
			web_site_api_entry["path"] = name
		} else {
			// Cannot find any exact matches in web_site_api table - check if we have any "*" matches in the web_site_api_table
			//Note(z): ("*" becomes ".*" for regex in api names)
			//TODO(z): implement full Regex?
			// ex: /api/*/graph could have any string that substitutes the * like .*
			grep_pattern := "%*%"
			sql = fmt.Sprintf("SELECT * FROM web_site_api WHERE web_site_id = %d AND name LIKE '%s' AND (web_protocol_action_id = '%d' OR web_protocol_action_id IS NULL)", web_site_id, grep_pattern, web_protocol_action_id)

			// Go through each of the search results and check if there are any matching regex expressions in the web_site_api table
			web_site_api_result = yudiendata.Query(db_web, sql)

			for _, element := range web_site_api_result {
				current_exp := strings.Replace(element["name"].(string), "*", ".*", -1)
				r, err := regexp.Compile(current_exp)

				if err != nil {
					continue
				}

				if r.MatchString(name) && r.FindString(name) == name {
					found_api = true
					web_site_api_entry = element
					web_site_api_entry["path"] = name
					//break // Match only the first result found
				}
			}
		}
	}

	// If we found a matching page
	if found_api {
		yudiencore.UdnLogLevel(nil, log_info, "\n\nFound API: %v\n\n", web_site_api_entry)
		dynamicPage_API(db_web, db, web_site, web_site_api_entry, uri, w, r)
	} else if len(web_site_page_result) > 0 {
		yudiencore.UdnLogLevel(nil, log_info, "\n\nFound Dynamic Page: %v\n\n", web_site_page_result[0])
		dynamePage_RenderWidgets(db_web, db, web_site, web_site_page_result[0], uri, w, r)
	} else {
		yudiencore.UdnLogLevel(nil, log_info, "\n\nPage not found: 404: %v\n\n", web_site_page_result)

		dynamicPage_404(uri, w, r)
	}
}

func dynamicPage_API(db_web *sql.DB, db *sql.DB, web_site map[string]interface{}, web_site_api map[string]interface{}, uri string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Check if there are any headers to be set for http.ResponseWriter specified by the api function (ex: grafana cors headers)
	if web_site_api["header_data_json"] != nil {

		header_data_bytes := []byte(web_site_api["header_data_json"].(string))

		var header_data_map interface{}

		err := json.Unmarshal(header_data_bytes, &header_data_map)

		if err == nil { // if there is an error, do nothing
			for key, value := range header_data_map.(map[string]interface{}) {
				w.Header().Set(key, value.(string))
			}
		}
	}

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
		yudiencore.UdnLogLevel(nil, log_debug, "Starting UDN Data: %s\n\n", yudienutil.SnippetData(udn_data, 120))

		yudiencore.UdnLogLevel(nil, log_debug, "Params: %s\n\n", yudienutil.SnippetData(param_map, 600))
	}

	// Get the base widget
	sql := fmt.Sprintf("SELECT * FROM web_widget")
	all_widgets := yudiendata.Query(db_web, sql)

	// Save all our base web_widgets, so we can access them anytime we want
	udn_data["base_widget"] = yudienutil.MapArrayToMap(all_widgets, "name")

	// Get UDN schema per request
	//TODO(g): Dont do this every request
	udn_schema := yudien.PrepareSchemaUDN(db_web)

	// Make sure messages are output to screen and logged when it is allowed to do so
	udn_schema["allow_logging"] = udn_data["web_site_page"].(map[string]interface{})["allow_logging"].(bool)

	// If we are being told to debug, do so
	if param_map["__debug"] != nil {
		udn_schema["udn_debug"] = true
	} else if yudiencore.Debug_Udn_Api == true {
		// API calls are harder to change than web page requests, so made a separate in code var to toggle debugging
		udn_schema["udn_debug"] = true
	}

	// Process the UDN, which updates the pool at udn_data
	if web_site_api["udn_data_json"] != nil {
		yudien.ProcessSchemaUDNSet(db_web, udn_schema, web_site_api["udn_data_json"].(string), udn_data)
	} else {
		yudiencore.UdnLogLevel(nil, log_debug, "UDN Execution: API: %s: None\n\n", web_site_api["name"])
	}

	// Set Cookies
	SetCookies(udn_data["set_cookie"].(map[string]interface{}), w, r)

	// Write whatever is in the API result map, as a JSON result
	var buffer bytes.Buffer
	body, _ := json.Marshal(udn_data["set_api_result"])
	buffer.Write(body)

	yudiencore.UdnLogLevel(nil, log_info, "Writing API body: %s\n\n", body)

	// Write out our output as HTML
	html_path := yudiencore.UdnDebugWriteHtml(udn_schema)

	if udn_schema["allow_logging"].(bool) {
		yudiencore.UdnLogLevel(nil, log_info, "UDN Debug HTML Log: %s\n", html_path)
	}

	// Write out the final page
	w.WriteHeader(udn_data["http_response_code"].(int))
	w.Write([]byte(buffer.String()))

	//yudiencore.UdnLogLevel(nil, log_trace, "API Result:\n\n%s\n", buffer.String())
}

func dynamePage_RenderWidgets(db_web *sql.DB, db *sql.DB, web_site map[string]interface{}, web_site_page map[string]interface{}, uri string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	sql := fmt.Sprintf("SELECT * FROM web_site_page_widget WHERE web_site_page_id = %d ORDER BY priority ASC", web_site_page["_id"])
	web_site_page_widgets := yudiendata.Query(db_web, sql)

	// Get the base web site widget.  Only where priority is non-negative.  Negative priority items are considered to be explicit rendering only, so we will render them when they are invoked.
	sql = fmt.Sprintf("SELECT * FROM web_site_page_widget WHERE _id = %d", web_site_page["base_page_web_site_page_widget_id"])
	base_page_widgets := yudiendata.Query(db_web, sql)

	// If we couldnt find the base_page_widget, we cannot render the any page for the website - return internal server error
	if len(base_page_widgets) < 1 {
		yudiencore.UdnLogLevel(nil, log_error, "No base page widgets found, returning internal server error 500\n")

		dynamicPage_500("Base page cannot be found.", w, r)
		return
	}

	base_page_widget := base_page_widgets[0]

	// Get the base widget
	sql = fmt.Sprintf("SELECT * FROM web_widget WHERE _id = %d", base_page_widget["web_widget_id"])
	base_widgets := yudiendata.Query(db_web, sql)

	base_page_html := base_widgets[0]["html"].(string)

	// Get UDN starting data values
	web_protocol_action := r.Method
	request_body := r.Body
	param_map := GetHTTPParams(r)
	header_map := r.Header
	cookie_array := r.Cookies()

	// Get our starting UDN data
	udn_data := GetStartingUdnData(db_web, db, web_site, web_site_page, uri, web_protocol_action, request_body, param_map, header_map, cookie_array)

	yudiencore.UdnLogLevel(nil, log_debug, "Starting UDN Data: %v\n\n", udn_data)

	// Get the base widget
	sql = fmt.Sprintf("SELECT * FROM web_widget")
	all_widgets := yudiendata.Query(db_web, sql)

	// Save all our base web_widgets, so we can access them anytime we want
	udn_data["base_widget"] = yudienutil.MapArrayToMap(all_widgets, "name")

	//fmt.Printf("Base Widget: base_list2_header: %v\n\n", udn_data["base_widget"].(map[string]interface{})["base_list2_header"])

	// We need to use this as a variable, so make it accessible to reduce casting
	page_map := udn_data["page"].(map[string]interface{})

	//TODO(g):HARDCODED: Im just forcing /login for now to make bootstrapping faster, it can come from the data source, think about it
	if uri != "/login" {
		if udn_data["user"].(map[string]interface{})["_id"] == nil {
			login_page_id := web_site["login_web_site_page_id"].(int64)
			login_page_sql := fmt.Sprintf("SELECT * FROM web_site_page WHERE _id = %d", login_page_id)
			login_page_rows := yudiendata.Query(db_web, login_page_sql)
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
	udn_schema := yudien.PrepareSchemaUDN(db_web)

	// If we are being told to debug, do so
	if param_map["__debug"] != nil {
		udn_schema["udn_debug"] = true
	}

	// Loop over the page widgets, and template them
	for _, site_page_widget := range web_site_page_widgets {
		// If we dont have any priority, dont render it on the normal pass.  It must be explicitly asked to be rendered.
		if site_page_widget["priority"] != nil && site_page_widget["priority"].(int64) < 1 {
			yudiencore.UdnLogLevel(nil, log_trace, "Skipping Page Widget: %s\n", site_page_widget["name"])
			continue
		}

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
		if site_page_widget["static_data_json"] != nil && site_page_widget["static_data_json"] != "" {
			err := json.Unmarshal([]byte(site_page_widget["static_data_json"].(string)), &widget_static)
			if err != nil {
				yudiencore.UdnLogLevel(nil, log_trace, "Static JSON Data: %s\n", site_page_widget["static_data_json"].(string))
				//log.Panic(err)
			}
		}

		// If we have web_widget specified, use it
		if site_page_widget["web_widget_id"] != nil {

			// Get the base widget
			sql = fmt.Sprintf("SELECT * FROM web_widget WHERE _id = %d", site_page_widget["web_widget_id"])
			page_widgets := yudiendata.Query(db_web, sql)
			page_widget = page_widgets[0]

			yudiencore.UdnLogLevel(nil, log_debug, "Page Widget: %s: %s\n", site_page_widget["name"], page_widget["name"])

			// wigdet_map has all the UDN operations we will be using to embed child-widgets into this widget
			//TODO(g): We need to use the page_map data here too, because we need to template in the sub-widgets.  Think about this after testing it as-is...
			if site_page_widget["data_json"] != nil {
				err := json.Unmarshal([]byte(site_page_widget["data_json"].(string)), &widget_map)
				if err != nil {
					log.Panic(err)
				}
			}

			udn_data["web_widget"] = page_widget

			// Processing UDN: which updates the data pool at udn_data
			if site_page_widget["udn_data_json"] != nil {
				yudien.ProcessSchemaUDNSet(db_web, udn_schema, site_page_widget["udn_data_json"].(string), udn_data)
			} else {
				yudiencore.UdnLogLevel(nil, log_debug, "UDN Execution: %s: None\n\n", site_page_widget["name"])
			}

			// Process the Widget's Rendering UDN statements (singles)
			for widget_key, widget_value := range widget_map {
				//fmt.Printf("\n\nWidget Key: %s:  Value: %v\n\n", widget_key, widget_value)

				// Force the UDN string into a string
				//TODO(g): Not the best way to do this, fix later, doing now for dev speed/simplicity
				widget_udn_string := []string{fmt.Sprintf("%v", widget_value)}

				// Process the UDN with our new method.  Only uses Source, as we are getting, but not setting in this phase
				widget_udn_result := yudien.ProcessUDN(db, udn_schema, widget_udn_string, udn_data)

				widget_map[widget_key] = fmt.Sprintf("%v", yudienutil.GetResult(widget_udn_result, type_string))

				//fmt.Printf("Widget Key Result: %s   Result: %s\n\n", widget_key, SnippetData(widget_map[widget_key], 600))
			}

			//fmt.Printf("Title: %s\n", widget_map.Map["title"])

			item_html := page_widget["html"].(string)
			/*
				item_html, err := ioutil.ReadFile(page_widget["path"].(string))
				if err != nil {
					log.Panic(err)
				}*/

			//TODO(g): Replace reading from the "path" above with the "html" stored in the DB, so it can be edited and displayed live
			//item_html := page_widget.Map["html"].(string)

			//fmt.Printf("Page Widget: %s   HTML: %s\n", page_widget["name"], SnippetData(page_widget["html"], 600))

			item_template := template.Must(template.New("text").Parse(string(item_html)))

			widget_map_template := yudienutil.NewTextTemplateMap()
			widget_map_template.Map = widget_map

			//fmt.Printf("  Templating with data: %v\n\n", SnippetData(widget_map, 600))

			item := yudienutil.StringFile{}
			err := item_template.Execute(&item, widget_map_template)
			if err != nil {
				log.Panic(err)
			}

			// Append to our total forum_list_string
			key := site_page_widget["name"]

			//fmt.Printf("====== Finalized Template: %s\n%s\n\n", key, item.String)

			//fmt.Printf("=-=-=-=-= UDN Data: Output:\n%v\n\n", udn_data["output"])

			page_map[key.(string)] = item.String

		} else if site_page_widget["web_widget_instance_id"] != nil {
			// Render the Widget Instance
			udn_update_map := make(map[string]interface{})
			yudien.RenderWidgetInstance(db_web, udn_schema, udn_data, site_page_widget, udn_update_map)

		} else if site_page_widget["web_data_widget_instance_id"] != nil {
			// Render the Widget Instance, from the web_data_widget_instance
			udn_update_map := make(map[string]interface{})
			yudien.RenderWidgetInstance(db_web, udn_schema, udn_data, site_page_widget, udn_update_map)

		} else {
			yudiencore.UdnLogLevel(nil, log_error, "No web_widget_id, web_widget_instance_id, web_data_widget_instance_id.  Site Page Widgets need at least one of these.")
			dynamicPage_404(uri, w, r)
			return
		}

	}

	// Get base page widget items.  These were also processed above, as the base_page_widget was included with the page...
	base_page_widget_map := yudienutil.NewTextTemplateMap()
	err := json.Unmarshal([]byte(base_page_widget["data_json"].(string)), &base_page_widget_map.Map)
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
			widget_udn_result := yudien.ProcessUDN(db, udn_schema, value_str, udn_data)

			if widget_udn_result != nil {
				page_map[key] = fmt.Sprintf("%v", yudienutil.GetResult(widget_udn_result, type_string))
			} else {
				// Use the base page widget, without any processing, because we got back nil
				page_map[key] = value_str
			}

			//// Set the value, static text
			//page_map[key] = value
		}
	}

	yudiencore.UdnLogLevel(nil, log_info, "Rendering base page\n")

	// Put them into the base page
	base_page_template := template.Must(template.New("text").Parse(string(base_page_html)))

	// Set up the TextTemplateMap for page_map, now that it is map[string]interface{}
	page_map_text_template_map := yudienutil.NewTextTemplateMap()
	page_map_text_template_map.Map = page_map

	// Write the base page
	base_page := yudienutil.StringFile{}
	err = base_page_template.Execute(&base_page, page_map_text_template_map)
	if err != nil {
		log.Fatal(err)
	}

	// Set Cookies
	SetCookies(udn_data["set_cookie"].(map[string]interface{}), w, r)

	// Write out our output as HTML
	html_path := yudiencore.UdnDebugWriteHtml(udn_schema)
	yudiencore.UdnLogLevel(nil, log_debug, "UDN Debug HTML Log: %s\n", html_path)

	// Write out the final page
	w.WriteHeader(udn_data["http_response_code"].(int))
	w.Write([]byte(base_page.String))

}

func dynamicPage_404(uri string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// DB Web
	db_web, err := sql.Open("postgres", config.Config.DefaultDatabase.ConnectOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer db_web.Close()

	// Get the base widget
	sql := fmt.Sprintf("SELECT * FROM web_widget WHERE name = 'error_404'")
	base_widgets := yudiendata.Query(db_web, sql)

	base_html := base_widgets[0]["html"].(string)

	/*
		base_html, err := ioutil.ReadFile("web/limitless5/error_404.html")
		if err != nil {
			log.Panic(err)
		}
	*/

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(base_html))
}

func dynamicPage_500(error_msg string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	//TODO(z): Create a formatted 500 Internal Server Error html page like dynamicPage_404 and return that instead of text
	error_response := fmt.Sprintf("500 Internal Error. %v", error_msg)

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(error_response))
}

func recoverError_500(w http.ResponseWriter, r *http.Request) {
	// Recover from internal panics until we could guarantee no panics
	if recover := recover(); recover != nil {
		//TODO(z): Add config option to include recover stack msg if needed
		error_message := fmt.Sprintf("Internal panic: %v", recover)

		yudiencore.UdnLogLevel(nil, log_error, "Error: %s\n", error_message)
		yudiencore.UdnLogLevel(nil, log_trace, "Error stack: %s\n", string(debug.Stack()))

		dynamicPage_500("", w, r)
	}
}
