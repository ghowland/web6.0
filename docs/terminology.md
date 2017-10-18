# Terminology

## Web 6.0 User Interface

`Data Widget (DW)` - A collection of Data Widget Instances, grouped by common data.  Frequently I make this based on a Database Table, but it doesn't have to be restricted to that.

`Data Widget Instance (DWI)` - A single widget instance, of a Data Widget (DW).  The DWI is the most common element to populate on a web page, as it uses the generic Widget Instance (WI) to populate with specific data (DWI).

`Widget Instance (WI)` - A generic widget instance (ex: Table, Form, Profile View, etc).  These are used by DWIs to go from specific to generic.  This is a combination of code (UDN), static data (JSON), inputs and widgets.

`Widget` - A snippet of HTML/CSS/JS for templating.  Can be used directly, but more commonly used inside Widget Instances (WI).

