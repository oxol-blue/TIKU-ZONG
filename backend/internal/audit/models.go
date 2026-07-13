package audit

import "time"

type Entry struct {
	ID          uint64    `json:"id"`
	AdminID     uint64    `json:"adminId"`
	AdminEmail  string    `json:"adminEmail"`
	Action      string    `json:"action"`
	Resource    string    `json:"resource"`
	RequestPath string    `json:"requestPath"`
	IPAddress   string    `json:"ipAddress"`
	HTTPStatus  int       `json:"httpStatus"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Page struct {
	Items    []Entry `json:"items"`
	Page     int     `json:"page"`
	PageSize int     `json:"pageSize"`
	Total    int     `json:"total"`
}
