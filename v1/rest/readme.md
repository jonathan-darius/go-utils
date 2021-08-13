# Contents

- [Response Helper](#response-helper)

# Response Helper

Each of the three functions listed below returns a logger containing the current context, and a uuid. This allows the function to be chained with .Log(msg string) to create a log in kibana. Alternatively, should there be a need for multiple logs, or further processing after the response, the logger could be stored in a variable to be used later.

##### Examples

Chaining

```go
rest.ResponseMessage(context, http.StatusInternalServerError).Log(
 "get bank accounts failed: " + err.Error())
```

Stored logger

```go
logger := rest.ResponseMessage(context, http.StatusOK)

err := someProcess()
if err != nil {
    logger.Log(err.Error()
}

err = moreProcess()
if err != nil {
    logger.Log(err.Error()
}
```

### Response Message

For non-2xx status codes, creates response in the form of

```json
{
    "error": uuid,
    "message": message
}
```

For 2xx status codes, the error uuid is omitted.

##### Example

Default message

```go
rest.ResponseMessage(context, http.StatusOK)
```

Custom Message

```go
rest.ResponseMessage(context, http.StatusOK, "payment success")
```

### Response Data

```json
"result": payload,
"message": message
```

Accepts payload of type interface{}

##### Example

```go
result, _ := service.DonationService.GetDonationByID(donationID)
rest.ResponseData(context, http.StatusOK, result)
```

### Response Error

```json
{
    "error": uuid,
    "message": message,
    "detail": {
        "field": error_detail
    }
}
```

ResponseError accepts an additional parameter of type `validator.ValidationErrors`, `map[string]string`, `validation.ErrorDetails`, or `string`.

Details in the form of `string`:

```json
"detail": {
    "error": detail
}
```

`validator.ValidationErrors`:

```json
"detail": {
    lowercase_error_field: error_tag
    ...
    lowercase_error_field: error_tag
}
```

`map[string]string`, `validation.ErrorDetails`:

```json
"detail": detail
```

##### Examples

`validator.ValidationErrors`:

```go
err = constants.Validator.Struct(requestBody)
if err != nil {
 rest.ResponseError(context, http.StatusBadRequest, err)
 return
}
```

`map[string]string`:

```go
rest.ResponseError(context, http.StatusBadRequest, map[string]string{"id": "invalid id"})
```

`validation.ErrorDetails`:

```go
det, code := validation.Validate(requestBody)
if det != nil {
    rest.ResponseError(context, code, det)
}
```
