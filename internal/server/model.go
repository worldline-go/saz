package server

type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type ResponseQuery struct {
	Columns      []string `json:"columns,omitempty"`
	Rows         [][]any  `json:"rows,omitempty"`
	RowsAffected int64    `json:"rows_affected,omitempty"`
	Duration     string   `json:"duration,omitempty"`
}

type Info struct {
	Databases []string `json:"databases"`
	Version   string   `json:"version"`
}

type ResponseBackground struct {
	PID     string `json:"pid"`
	Message string `json:"message"`
}

type ResponseNoteCell struct {
	Description  string   `json:"description,omitempty"`
	Columns      []string `json:"columns,omitempty"`
	Rows         [][]any  `json:"rows,omitempty"`
	RowsAffected int64    `json:"rows_affected,omitempty"`
	Duration     string   `json:"duration,omitempty"`
	Status       string   `json:"status"`
	Error        string   `json:"error,omitempty"`
}

type RenderRequest struct {
	Content string `json:"content"`
	Data    any    `json:"data"`
}
