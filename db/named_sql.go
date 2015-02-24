package db

var (
	namedSQL = make(map[string]string)
)

func SetSQL(name, content string) {
	namedSQL[name] = content
}

func SQL(name string) string {
	return namedSQL[name]
}
