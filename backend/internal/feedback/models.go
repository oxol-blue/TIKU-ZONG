package feedback

type Input struct {
	RequestID    string `json:"requestId" binding:"required"`
	Question     string `json:"question" binding:"required"`
	FeedbackType string `json:"feedbackType" binding:"required"`
	Comment      string `json:"comment"`
}
