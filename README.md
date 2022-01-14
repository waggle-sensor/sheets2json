# sheets2json
JSON API to serve google sheets table


## Config

See `config.yaml` for an example. Urls are for the google sheets API and have this form:

`https://sheets.googleapis.com/v4/spreadsheets/.../values/Sheet1?key=...`


The field `types` indicates if the first column in a spreadsheet is used to indicate the type of the columns fields. `bool` and `int` are supported, `string` is the default.