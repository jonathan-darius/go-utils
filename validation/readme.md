Makes use of four field tags: `validate`, `id`, `process`, `json`

`json` is used to get the field name when processing error details. If this tag is empty, or "-", the name defaults to the lowercased struct field name. Returns 400 on error.

`validate` tags is passed to go-validator. Errors from go-validator is parsed into validation.ErrorDetails.

`"id"` tags has three possible values: `"required"`, `"valid"` and `"allow-zero"`. `"required"` validates that the field is not empty, and is valid, sets the status code to 400 when empty and 422 when invalid. `"valid"` behaves similarly to `"required"` but returns 200 when empty. `"allow-zero"` allows zero value in id. When used on a slice, the tags will be applied to each element.

`"process"` behaves in the same way as validate but returns 422 on error.

`validation.ErrorDetails` is a map[string]string with an additional method `Add(key, value string)`. Appends detail with "|" as separator. This type can be directly passed to `rest.ResponseError` as details.
