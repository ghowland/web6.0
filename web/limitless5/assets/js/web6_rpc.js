// Global variable to store data as a global dictionary
__js_data = new Object();


// RPC: Remote Procedure Call, automated back to path/rpc/FunctionName
//
// On return will automatically fill in Object() key results into tag IDs of
//    the same name.  If "__js" key is found, assumes this is a array of strings
//    and performs an eval() on each after updating all the tag IDs
//
function RPC(url, input_data, on_complete_function) {
    //alert('RPC: ' + url + ': ' + input_data.toSource())  // Use to test

    // AJAX Code To Submit Form.
    $.ajax({
        //type: "POST",
        type: "GET",
        url: url,
        data: input_data,
        cache: false,
        success: function(data)
        {
            ProcessRPCData(data);
            if (typeof on_complete_function != 'undefined') { on_complete_function(); }
        }
    });
}


function RPCUrl(url, data) {
    $.getJSON(url, function(data) {
        ProcessRPCData(data);
    });
}


// Process the RPC response data, can be done without using the RPC call as well
function ProcessRPCData(data) {
    var js_execute = undefined;
    var reload_page = undefined;
    var load_page = undefined;

    //alert(data);
    data = JSON.parse(data);
    //alert(data);

    // Process the HTML sections, skip __js and __js_data
    for (var key in data)
    {
        // Non-Javascript data gets put into ID elements, if they exist
        if (key != '_js' && key != '_js_data' && key != '_reload_page' && key != '_load_page' && key != '_success' && key != '_failure') {
            //TODO(g): Is it worth checking if the ID exists in the DOM?  I dont think so, but think about it...
            // Start by clearing the existing data and freeing references
            $("#" + key).empty();
            //alert('Procesing: ' + key + ' :: ' + data[key]);
            $("#" + key).html(data[key]);
        }
        // Save our Javascript array until later so we can deal with it then
        else if (key == '_js') {
            js_execute = data[key];        //TODO(g): Test this.  Havent yet...
        }
        // Save our Javascript array until later so we can deal with it then
        else if (key == '_js_data') {
            __js_data = data[key];
        }
        // Else, if this is a key to reload the page (self or somewhere else)
        else if (key == '_reload_page') {
            reload_page = data[key];
        }
        // Else, if this is a key to load a page
        else if (key == '_load_page') {
            load_page = data[key];
        }
        // Else, this is a success message
        else if (key == '_success') {
            swal({
                title: "Success!",
                text: data[key],
                confirmButtonColor: "#66BB6A",
                type: "success"
            });
        }
        // Else, this is a success message
        else if (key == '_failure') {
            swal({
                title: "Fail...",
                text: data[key],
                confirmButtonColor: "#EF5350",
                type: "error"
            });
        }
    }

    // If we had JS data, eval() it now.  This is JS code that is not related to any specific
    //    element, and takes place after all elements have been updated (above)
    if (js_execute != undefined) {
        //alert(js_execute);    // Debug
        eval(js_execute);     // Execute arbitrary Javascript, to control the page from the server
    }

    // Reload the page
    if (reload_page != undefined) {
        location.reload();
    }
    // Load a page
    else if (load_page != undefined) {
        window.location = load_page;
    }
}

