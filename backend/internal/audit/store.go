package audit

import (
	"context"
	"database/sql"
	"strings"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) Create(ctx context.Context, item Entry) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO admin_audit_logs (admin_id, admin_email, action, resource, request_path, ip_address, http_status) VALUES (?, ?, ?, ?, ?, ?, ?)`, item.AdminID, item.AdminEmail, item.Action, item.Resource, item.RequestPath, item.IPAddress, item.HTTPStatus)
	return err
}

func (s *Store) List(ctx context.Context, page, pageSize int, search string) (Page, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	search = strings.TrimSpace(search)
	where, args := "", []any{}
	if search != "" {
		where = " WHERE admin_email LIKE ? OR resource LIKE ? OR request_path LIKE ?"
		like := "%" + search + "%"
		args = append(args, like, like, like)
	}
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM admin_audit_logs"+where, args...).Scan(&total); err != nil {
		return Page{}, err
	}
	queryArgs := append(args, pageSize, (page-1)*pageSize)
	rows, err := s.db.QueryContext(ctx, `SELECT id, admin_id, admin_email, action, resource, request_path, ip_address, http_status, created_at FROM admin_audit_logs`+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return Page{}, err
	}
	defer rows.Close()
	items := make([]Entry, 0)
	for rows.Next() {
		var item Entry
		if err := rows.Scan(&item.ID, &item.AdminID, &item.AdminEmail, &item.Action, &item.Resource, &item.RequestPath, &item.IPAddress, &item.HTTPStatus, &item.CreatedAt); err != nil {
			return Page{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return Page{}, err
	}
	return Page{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}
