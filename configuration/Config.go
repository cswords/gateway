package configuration

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

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

// LoadConfig TODO
func LoadConfig(location string, loaders ...(func(string) []byte)) *Config {
	c := &(Config{})
	var b []byte
	if len(loaders) == 0 {
		loaders = []func(string) []byte{FromLocal}
	}
	for _, loader := range loaders {
		b = loader(location)
		if b != nil {
			break
		}
	}
	return c.loadBinary(b)
}

// FromLocal TODO
func FromLocal(loc string) []byte {

	if !strings.HasPrefix(loc, "./") {
		return nil
	}

	folder, err := filepath.Abs(filepath.Dir("."))
	if err != nil {
		log.Println("load binary from local err", err)
		return nil
	}
	data, err := ioutil.ReadFile(folder + strings.TrimPrefix(loc, "."))
	if err != nil {
		log.Println("load binary from local  err", err)
		return nil
	}
	return data
}
