package configloader

import (
	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Event describes a config file change without exposing fsnotify.
type Event struct {
	Name string
	Op   string
}

// Loader wraps config file parsing and watching behind project-owned types.
type Loader struct {
	v *viper.Viper
}

// New creates an isolated config loader.
func New() *Loader {
	return &Loader{v: viper.New()}
}

// LoadEnv loads a dotenv file if present.
func LoadEnv(path string) {
	_ = godotenv.Load(path)
}

func (l *Loader) SetConfigFile(path string) {
	l.v.SetConfigFile(path)
}

func (l *Loader) ReadInConfig() error {
	return l.v.ReadInConfig()
}

func (l *Loader) AllSettings() map[string]any {
	return l.v.AllSettings()
}

func (l *Loader) Set(key string, value any) {
	l.v.Set(key, value)
}

func (l *Loader) Unmarshal(out any) error {
	return l.v.Unmarshal(out)
}

func (l *Loader) OnConfigChange(handler func(Event)) {
	l.v.OnConfigChange(func(e fsnotify.Event) {
		handler(Event{Name: e.Name, Op: e.Op.String()})
	})
}

func (l *Loader) WatchConfig() {
	l.v.WatchConfig()
}
