package api

type TaskResponse struct {
	TaskID string `json:"task_id"`
}

type Task struct {
	TaskID   string            `json:"task_id"`
	Priority uint8             `json:"priority" validate:"required"`
	QueueID  string            `json:"queue_id" validate:"required"`
	CallData map[string]string `json:"call_data" validate:"required"`
	Position int               `json:"position"`
}

type Queue struct {
	QueueID     string `json:"queue_id"`
	QueueName   string `json:"queue_name" validate:"required"`
	MaxPriority uint8  `json:"max_priority" validate:"required"`
}
