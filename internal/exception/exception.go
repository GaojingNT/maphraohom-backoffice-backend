package exception

type ErrorResponse struct {
	Code    string           `json:"code"`
	Message string           `json:"message,omitempty"`
	Errors  []ParameterError `json:"errors,omitempty"`
}

type ParameterError struct {
	FailedField string `json:"field"`
	Tag         string `json:"tag"`
	Value       string `json:"value"`
}
