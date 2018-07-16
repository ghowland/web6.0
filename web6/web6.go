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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ghowland/yudien/yudiencore"
)

const (
	type_int          = iota
	type_float        = iota
	type_string       = iota
	type_string_force = iota // This forces it to a string, even if it will be ugly, will print the type of the non-string data too.  Testing this to see if splitting these into 2 yields better results.
	type_array        = iota // []interface{} - takes: lists, arrays, maps (key/value tuple array, strings (single element array), ints (single), floats (single)
	type_map          = iota // map[string]interface{}
)

const ( // order matters for log levels
	log_off   = iota
	log_error = iota
	log_warn  = iota
	log_info  = iota
	log_debug = iota
	log_trace = iota
)

// Core Web Page Handler.  All other routing occurs inside this function.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Defer-recover for panics by returning a 500 internal server error (until we can guarantee no panics)
	// This way there is no connection reset for the user
	// Note that this will not recover from go routines when we implement concurrency in the future
	defer recoverError_500(w, r)

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
			log.Panic(err)
		}

		// If this isnt a directory
		if !file_info.IsDir() {
			is_static = true

			size := file_info.Size()

			data := make([]byte, size)
			_, err := file.Read(data)
			if err != nil {
				log.Panic(err)
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

// Set cookies against the HTTP Request
func SetCookies(cookie_map map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	for key, value := range cookie_map {
		//TODO(g):REMOVE: Testing only...
		new_cookie := http.Cookie{}
		new_cookie.Name = key
		new_cookie.Value = fmt.Sprintf("%v", value)
		new_cookie.Path = "/"
		http.SetCookie(w, &new_cookie)

		yudiencore.UdnLogLevel(nil, log_info, "** Setting COOKIE: %s = %s\n", key, value)
	}
}

// Get the params of the HTTP request
func GetHTTPParams(r *http.Request) map[string]interface{} {

	// Check the web protocol action - for POST/PUT requests, params are found in the body
	param_map := make(map[string]interface{})

	web_protocol_action := r.Method
	http_header := r.Header.Get("Content-Type")

	// For POST & PUT requests, we need to return the body of the request
	if web_protocol_action == "POST" || web_protocol_action == "PUT" {
		// Parse the body different depending on the type of the body (ex: JSON, form data, etc.)
		if http_header == "application/json" {
			// Read the body of the request (json)
			if body_bytes, err := ioutil.ReadAll(r.Body); err == nil {
				err = json.Unmarshal(body_bytes, &param_map)

				if err != nil {
					param_map = nil
				}
			}
		} else { // POST request where the body is not JSON and is rather form data
			err := r.ParseForm()

			// ParseFrom() returns map[string][]string - need to convert it to map[string]interface{}
			if err == nil {
				param_map_strings := r.PostForm

				for key, value := range param_map_strings {
					param_map[key] = value[0]
				}
			}
		}
	} else { // GET and other requests
		param_map_strings := r.URL.Query()

		// r.URL.Query() returns map[string][]string - need to convert it to map[string]interface{}
		for key, value := range param_map_strings {
			param_map[key] = value[0]
		}
	}

	return param_map
}
