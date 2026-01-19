package job

import "time"

type Job struct {
	ID        string    `json:"id"`
	Filename  string    `json:"filename"`
	FilePath  string    `json:"filepath"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func New(id, filename, filepath string) *Job {
	return &Job{
		ID:        id,
		Filename:  filename,
		FilePath:  filepath,
		Status:    "Pending",
		CreatedAt: time.Now(),
	}
}
