package configuration

import (
	"log"

	"gopkg.in/yaml.v2"
)

// Config represent the root node of the structures
type Config struct {
	Server Server `yaml:"server"`
}

// Server TODO
type Server struct {
	Port    string   `yaml:"port"`
	Routers []Router `yaml:"routers"`
}

// Router TODO
type Router struct {
	Prefix      string       `yaml:"prefix"`
	Middlewares []Middleware `yaml:"middlewares"`
	Handlers    []Handler    `yaml:"handlers"`
}

// Middleware TODO
type Middleware struct {
	Type   string            `yaml:"type"`
	Config map[string]string `yaml:"config"`
}

// Handler TODO
type Handler struct {
	Path   string            `yaml:"path"`
	Type   string            `yaml:"type"`
	Config map[string]string `yaml:"config"`
}

func (c *Config) loadBinary(b []byte) *Config {

	err := yaml.Unmarshal(b, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}

// Load TODO
func (c *Config) Load(loaders ...(func() []byte)) *Config {
	var b []byte
	for _, loader := range loaders {
		b = loader()
		if b != nil {
			break
		}
	}
	return c.loadBinary(b)
}
