package rules

type Config struct {
	HandSize int
}

func DefaultConfig() Config {
	return Config{HandSize: 7}
}
