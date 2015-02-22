package sql

type NamedSQL struct {
	sql map[string]string
}

func NewNamedSQL() NamedSQL {
	return NamedSQL{sql: make(map[string]string)}
}

func (n *NamedSQL) SetSQL(name, content string) {
	n.sql[name] = content
}

func (n NamedSQL) SQL(name string) string {
	return n.sql[name]
}
