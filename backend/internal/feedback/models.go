package feedback

import "time"

type Input struct {
	RequestID    string `json:"requestId" binding:"required"`
	Question     string `json:"question" binding:"required"`
	FeedbackType string `json:"feedbackType" binding:"required"`
	Comment      string `json:"comment"`
}

type Item struct {
	ID           uint64    `json:"id"`
	RequestID    string    `json:"requestId"`
	QuestionHash string    `json:"questionHash"`
	FeedbackType string    `json:"feedbackType"`
	Comment      string    `json:"comment"`
	CreatedAt    time.Time `json:"createdAt"`
}

type AdminItem struct {
	Item
	UserID    uint64 `json:"userId"`
	UserEmail string `json:"userEmail"`
}

type Page struct {
	Items    []AdminItem `json:"items"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Total    int         `json:"total"`
}
