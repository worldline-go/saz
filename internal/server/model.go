package server

type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type Info struct {
	Databases []string `json:"databases"`
	Version   string   `json:"version"`
}
