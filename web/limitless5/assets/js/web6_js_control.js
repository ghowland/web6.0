/*
web6: JS Control.  Control JS DOM and code stuff automatically, without living in crazy town.

Requires JQuery

Author: Geoff Howland <geoff@gmail.com>
License: MIT
*/

var __web6_js_control_data_store = new Object()
var __web6_js_control_data_store_global = new Object()


function JsControl_Register(namespace, dom_id, control_data_list) {
    if (__web6_js_control_data_store[namespace] == undefined) {
        __web6_js_control_data_store[namespace] = new Object()
        __web6_js_control_data_store[namespace]['namespace'] = namespace

        __web6_js_control_data_store[namespace]['control_data_list'] = new Array()
        __web6_js_control_data_store[namespace]['string'] = new Object()
        __web6_js_control_data_store[namespace]['data'] = new Object()
        __web6_js_control_data_store[namespace]['var'] = new Object()
    }

    // Add all the control data items to the list
    for (item_count in control_data_list) {
        item = control_data_list[item_count]

        // Add the dom_id here, because we want this to be item specific, but we dont want to have to put it into the item spec, because we dont know the UUID in the spec as that is live data
        if (item['value_dom'] == undefined) {
            item['value_dom'] = dom_id
        }

        __web6_js_control_data_store[namespace]['control_data_list'].push(item)
    }

    JsControl_InitEventHandlers(namespace)
}

function JsControl_InitEventHandlers(namespace) {
    data_store = __web6_js_control_data_store[namespace]

    // Set up event handles for collecting vars data from the control_data_list, and get initial values
    for (var item_count in data_store['control_data_list']) {
        item = data_store['control_data_list'][item_count]

        // Var items -- On change they update the vars data store for this namespace, and then update the Strings and Data
        if (item['type'] === 'var') {
            data_store['var'][item['name']] = $('#' + item['value_dom']).val()

            $('#'+ item['dom_id']).ready().on('change', function (event) {
                __web6_js_control_data_store[namespace]['var'][item['name']] = $('#' + item['value_dom']).val()

                if (item['publish'] != undefined) {
                    __web6_js_control_data_store_global[item['publish']] = __web6_js_control_data_store[namespace]['var'][item['name']]
                }

                JsControl_UpdateStringsAndData(namespace)
            })
        }
        // RPC items -- For their specified event type, they execute the RPC
        else if (item['type'] === 'rpc') {
            $('#'+ item['dom_id']).ready().on(item['event'], function (event) {
                RPC(JsControl_Get(namespace, item['name']), item['data'])
            })
        }
        // Eval items -- For their specified event type, they execute the arbitrary code
        else if (item['type'] === 'eval') {
            $('#'+ item['dom_id']).ready().on(item['event'], function (event) {
                eval(item['eval'])
            })
        }
    }

    // Update all the strings, now that we initialized all the vars from the DOM elements
    JsControl_UpdateStringsAndData(namespace)
}

function JsControl_UpdateStringsAndData(namespace) {
    data_store = __web6_js_control_data_store[namespace]

    alert('Data Store: ' + namespace + '  Data: ' + JSON.stringify(data_store.toSource()))

    //TODO(g): Do data later.  I dont have a use-case for it yet, but it will definitely be implemented as it's useful
    //...

    // Populate the data from the control_data_list
    for (var item_count in data_store['control_data_list']) {
        item = data_store['control_data_list'][item_count]

        if (item['type'] === 'string') {
            // Initial format_string, we will iterate over all vars/globals, and replace any template keys in this string
            format_string = item['format']

            // Vars
            for (var var_item_access_key in data_store['var']) {
                var_item_value = data_store['var'][var_item_access_key]
                var_item_key = '[[[' + var_item_access_key + ']]]'

                alert('Var Replace: ' + var_item_key + '  With: ' + var_item_value)
                format_string = format_string.replace(var_item_key, var_item_value)
            }

            // Globals
            for (var var_item_access_key in __web6_js_control_data_store_global) {
                var_item_value = __web6_js_control_data_store_global[var_item_access_key]
                var_item_key = '[[[' + var_item_access_key + ']]]'

                alert('Global Replace: ' + var_item_key + '  With: ' + var_item_value)
                format_string = format_string.replace(var_item_key, var_item_value)
            }

            // After all the formatting, assign it into our formatted dict
            data_store['string'][item['name']] = format_string
        }
    }

    alert('Strings Formatted: ' + JSON.stringify(__web6_js_control_data_store[namespace]['string'].toSource()))
}

function JsControl_Get(namespace, key) {
    data_store = __web6_js_control_data_store[namespace]

    //TODO(g): Implement data once we have the use-case for it
    //...

    if (data_store['string'][key] !== undefined) {
        return data_store['string'][key]
    }
    else if (data_store['var'][key] !== undefined) {
        return data_store['var'][key]
    }
    else if (__web6_js_control_data_store_global[key] !== undefined) {
        return __web6_js_control_data_store_global[key]
    }

    return undefined
}
