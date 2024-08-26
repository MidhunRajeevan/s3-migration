package model

// Error response
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
