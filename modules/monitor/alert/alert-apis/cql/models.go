package cql

// AlertHistory .
type AlertHistory struct {
	GroupID    string `db:"group_id"`
	Timestamp  int64  `db:"timestamp"`
	AlertState string `db:"alert_state"`
	Title      string `db:"title"`
	Content    string `db:"content"`
	DisplayURL string `db:"display_url"`
}
