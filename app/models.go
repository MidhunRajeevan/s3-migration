package app

import "time"

type DirectoryRecord struct {
	ID          int64     `json:"id"`
	Did         string    `json:"did"`
	Totalfiles  int64     `json:"totalFiles"`
	Status      string    `json:"status"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt time.Time `json:"completedAt"`
}
