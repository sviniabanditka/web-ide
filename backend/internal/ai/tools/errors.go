package tools

type ToolResult struct {
	OK    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Meta  *ResultMeta `json:"meta,omitempty"`
	Error *ToolError  `json:"error,omitempty"`
}

type ResultMeta struct {
	DurationMs int64  `json:"duration_ms"`
	Truncated  bool   `json:"truncated,omitempty"`
	SHA        string `json:"sha,omitempty"`
}

type ToolError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

const (
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeNotFound      = "FILE_NOT_FOUND"
	ErrCodePermission    = "PERMISSION_DENIED"
	ErrCodeTimeout       = "TOOL_TIMEOUT"
	ErrCodeSizeLimit     = "SIZE_LIMIT_EXCEEDED"
	ErrCodeUserRejected  = "USER_REJECTED"
	ErrCodeExecution     = "EXECUTION_ERROR"
	ErrCodeInvalidPath   = "INVALID_PATH"
	ErrCodeNotExecutable = "NOT_EXECUTABLE"
	ErrCodeAlreadyExists = "ALREADY_EXISTS"
)

func NewSuccessResult(data interface{}) ToolResult {
	return ToolResult{
		OK:   true,
		Data: data,
		Meta: &ResultMeta{},
	}
}

func NewSuccessResultWithMeta(data interface{}, meta ResultMeta) ToolResult {
	return ToolResult{
		OK:   true,
		Data: data,
		Meta: &meta,
	}
}

func NewErrorResult(code, message string, details interface{}) ToolResult {
	return ToolResult{
		OK: false,
		Error: &ToolError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}
