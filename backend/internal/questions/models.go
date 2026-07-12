package questions

import "time"

type OptionInput struct {
	Key  string `json:"key" binding:"required"`
	Text string `json:"text" binding:"required"`
}

type ImportInput struct {
	Question    string        `json:"question" binding:"required"`
	Type        string        `json:"type"`
	Options     []OptionInput `json:"options"`
	Answer      string        `json:"answer" binding:"required"`
	AnswerRaw   string        `json:"answerRaw"`
	Platform    string        `json:"platform"`
	Subject     string        `json:"subject"`
	Source      string        `json:"source"`
	CollectedAt *time.Time    `json:"collectedAt"`
}

type BatchImportInput struct {
	Items []ImportInput `json:"items" binding:"required,min=1,max=1000"`
}

type Option struct {
	Key      string `json:"key"`
	Text     string `json:"text"`
	Position int    `json:"position"`
}

type Answer struct {
	Text     string `json:"text"`
	Raw      string `json:"raw"`
	Position int    `json:"position"`
}

type Question struct {
	ID          uint64     `json:"id"`
	Question    string     `json:"question"`
	Type        string     `json:"type"`
	Platform    string     `json:"platform"`
	Subject     string     `json:"subject"`
	Source      string     `json:"source"`
	CollectedAt *time.Time `json:"collectedAt,omitempty"`
	Options     []Option   `json:"options"`
	Answers     []Answer   `json:"answers"`
}
