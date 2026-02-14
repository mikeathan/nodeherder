package protocol

type ToolResponse struct {
	Status     string      `json:"status"` // success, error, ambiguous
	Data       any         `json:"data,omitempty"`
	Hint       string      `json:"hint,omitempty"` // Hint for LLM when results are empty
	Error      *ToolError  `json:"error,omitempty"`
	Candidates []Candidate `json:"candidates,omitempty"`
}

type ToolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Candidate struct {
	DeviceID string  `json:"device_id"`
	Name     string  `json:"name"`
	Score    float64 `json:"score"`
}

func NewSuccessResponse(data any) *ToolResponse {
	return &ToolResponse{
		Status: "success",
		Data:   data,
	}
}

func NewErrorResponse(code, message string) *ToolResponse {
	return &ToolResponse{
		Status: "error",
		Error: &ToolError{
			Code:    code,
			Message: message,
		},
	}
}

func NewAmbiguousResponse(candidates []Candidate) *ToolResponse {
	return &ToolResponse{
		Status:     "ambiguous",
		Candidates: candidates,
	}
}
