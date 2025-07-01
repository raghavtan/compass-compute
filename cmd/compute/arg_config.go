package compute

var (
	verbose bool
	config  *Config
)

type Config struct {
	ComponentName string
	Verbose       bool
}
