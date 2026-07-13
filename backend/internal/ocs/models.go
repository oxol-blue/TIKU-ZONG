package ocs

import "time"

type SourceInput struct {
	Name         string            `json:"name" binding:"required"`
	Homepage     string            `json:"homepage"`
	URL          string            `json:"url" binding:"required"`
	Method       string            `json:"method"`
	Headers      map[string]string `json:"headers"`
	Data         map[string]any    `json:"data"`
	SuccessPath  string            `json:"successPath"`
	SuccessValue string            `json:"successValue"`
	QuestionPath string            `json:"questionPath"`
	AnswerPath   string            `json:"answerPath"`
	Priority     int               `json:"priority"`
	Enabled      bool              `json:"enabled"`
}

type Source struct {
	ID           uint64            `json:"id"`
	Name         string            `json:"name"`
	Homepage     string            `json:"homepage"`
	URL          string            `json:"url"`
	Method       string            `json:"method"`
	Headers      map[string]string `json:"headers"`
	Data         map[string]any    `json:"data"`
	SuccessPath  string            `json:"successPath"`
	SuccessValue string            `json:"successValue"`
	QuestionPath string            `json:"questionPath"`
	AnswerPath   string            `json:"answerPath"`
	Priority     int               `json:"priority"`
	Enabled      int               `json:"enabled"`
	CreatedAt    time.Time         `json:"createdAt"`
}

type Result struct {
	Question string
	Answer   string
	Source   string
	Elapsed  time.Duration
}
