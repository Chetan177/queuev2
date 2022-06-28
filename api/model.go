package api

type TaskResponse struct {
	TaskID string `json:"task_id"`
}

type Task struct {
	TaskID   string `json:"task_id"`
	Priority uint8  `json:"priority" validate:"required"`
	QueueID  string `json:"queue_id" validate:"required"`
}
