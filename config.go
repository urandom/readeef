package readeef

type Config struct {
	DB struct {
		Driver  string
		Connect string
	}
	Auth struct {
		Secret          string
		IgnoreURLPrefix []string `gcfg:"ignore-url-prefix"`
	}
}
