package rest

// this code is used strictly for swagger documentation purpose only
// some error message may differ from this struct layout

// helper uuid message struct
type message struct {
	Message string `json:"message" example:"00000000-0000-0000-0000-000000000000"`
}

// Generic error message
type GenericErrorResponse struct {
	Error string `json:"error" example:"Error"`
	message
}

// 200 OK
type OKResponse struct {
	Message string `json:"message" example:"Success"`
}

// 201 Created
type CreatedResponse struct {
	Message string `json:"message" example:"Successfully Inserted Data"`
}

// 400 Bad Request
type BadRequestResponse struct {
	Error string `json:"error" example:"Bad Request"`
	message
}

// 401 Unauthorized
type UnauthorizedResponse struct {
	Error string `json:"error" example:"Unauthorized"`
	message
}

// 403 Forbidden
type ForbiddenResponse struct {
	Error string `json:"error" example:"Forbidden"`
	message
}

// 404 Not Found
type NotFoundResponse struct {
	Error string `json:"error" example:"Not Found"`
	message
}

// 409 Resource Conflict
type ResourceConflictResponse struct {
	Error string `json:"error" example:"Resource Conflict"`
	message
}

// 410 Gone
type GoneResponse struct {
	Error string `json:"error" example:"Resource Expired"`
	message
}

// 422 Unprocessable Entity
type UnprocessableEntityResponse struct {
	Error string `json:"error" example:"Unprocessable Entity"`
	message
}

// 423 Locked
type LockedResponse struct {
	Error string `json:"error" example:"Locked"`
	message
}

// 500 Internal Server Error
type InternalServerErrorResponse struct {
	Error string `json:"error" example:"Internal Server Error"`
}
