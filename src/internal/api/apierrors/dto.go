package apierrors

type ErrorResponseDTO struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	TraceID     string                 `json:"traceId"`
	Timestamp   string                 `json:"timestamp"`
	Path        string                 `json:"path"`
	Details     map[string]interface{} `json:"details,omitempty"`
	FieldErrors []FieldErrorDTO        `json:"fieldErrors,omitempty"`
}

type FieldErrorDTO struct {
	Field         string      `json:"field"`
	Issue         string      `json:"issue"`
	RejectedValue interface{} `json:"rejectedValue,omitempty"`
}
