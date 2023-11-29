package context

import (
	"encoding/json"
	"os"
)

type Language struct {
	Name    string     `json:"name"`
	Install [][]string `json:"install"`
}

var (
	Go = &Language{
		Name: "go",
		Install: [][]string{
			{"go", "mod", "tidy"},
			{"go", "build"},
			{"go", "install"},
		},
	}

	Python = &Language{
		Name: "python",
		Install: [][]string{
			{"pip", "install", "-r", "requirements.txt"},
			{"python", "setup.py", "build"},
			{"python", "setup.py", "install"},
		},
	}
)

func LanguageFromName(name string) *Language {
	switch name {
	case "go":
		return Go
	case "python":
		return Python
	default:
		return nil
	}
}

type Device struct {
	Name     string   `json:"name"`
	Requires []string `json:"requires"`
	URL      string   `json:"url"`
	Start    string   `json:"start"`
	Language string   `json:"language"`
}

type Context struct {
	Devices map[string]*Device `json:"devices"`
}

var InitialContext = &Context{
	Devices: make(map[string]*Device),
}

var ErrDeviceExists = os.ErrExist

func (c *Context) AddDevice(d *Device) error {
	if c.Devices[d.URL] != nil {
		return ErrDeviceExists
	}
	if c.Devices == nil {
		c.Devices = make(map[string]*Device)
	}
	c.Devices[d.URL] = d
	return nil
}

func FromJson(filename string) (*Context, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	var ctx Context
	err = json.NewDecoder(f).Decode(&ctx)
	if err != nil {
		return nil, err
	}
	return &ctx, nil
}
