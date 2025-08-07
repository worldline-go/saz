package server

type Response struct {
	Message      string   `json:"message,omitempty"`
	Error        string   `json:"error,omitempty"`
	Data         any      `json:"data,omitempty"`
	Columns      []string `json:"columns,omitempty"`
	RowsAffected int64    `json:"rows_affected,omitempty"`
}

type Info struct {
	Databases []string `json:"databases"`
	Version   string   `json:"version"`
}
