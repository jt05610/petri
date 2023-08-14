package codegen

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/prisma"
	"github.com/jt05610/petri/yaml"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed template
var templateDir embed.FS

type Language string

type Params struct {
	Language
	Port         int
	OutDir       string
	DeviceID     string
	RabbitMQURI  string
	AMQPExchange string
}

type DevParams struct {
	*device.Device
	*Params
}
type Generator struct {
	*Params
	*prisma.DeviceClient
	dev *DevParams
}

type TypeMap map[string]string

type LangTypeMap map[Language]TypeMap

var langTypeMap = LangTypeMap{
	"go": TypeMap{
		"string":  "string",
		"number":  "float64",
		"boolean": "bool",
	},
	"python": TypeMap{
		"string":  "str",
		"number":  "float",
		"boolean": "bool",
	},
	"ts": TypeMap{
		"string":  "string",
		"number":  "number",
		"boolean": "boolean",
	},
}

func langType(l Language) func(ft labeled.FieldType) string {
	return func(ft labeled.FieldType) string {
		return langTypeMap[l][strings.ToLower(string(ft))]
	}
}

func (g *Generator) Generate(ctx context.Context) error {
	if err := g.Connect(); err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	defer func() {
		err := g.Disconnect()
		if err != nil {
			fmt.Printf("error disconnecting from database: %v", err)
		}
	}()
	steps := []struct {
		Text string
		f    func(ctx context.Context) error
	}{
		{
			Text: "Loading device",
			f:    g.loadDev,
		},
		{
			Text: "Making instance",
			f:    g.makeInstance,
		},
		{
			Text: "Saving device",
			f:    g.saveDev,
		},
		{
			Text: "Making directory tree",
			f:    g.makeDirTree,
		},
	}

	for i, step := range steps {
		fmt.Printf("Step %d: %s\n", i, step.Text)
		err := step.f(ctx)
		if err != nil {
			return fmt.Errorf("error in step %d: %v", i, err)
		}
	}
	return nil
}

func sentenceToPascalCase(s string) string {
	splitString := strings.Split(s, " ")
	for i, word := range splitString {
		splitString[i] = strings.ToUpper(word[0:1]) + word[1:]
	}
	return strings.Join(splitString, "")
}

func sentenceToSnakeCase(s string) string {
	splitString := strings.Split(s, " ")
	for i, word := range splitString {
		splitString[i] = strings.ToLower(word)
	}
	return strings.Join(splitString, "_")
}

func sentenceToCamelCase(s string) string {
	splitString := strings.Split(s, " ")
	for i, word := range splitString {
		if i == 0 {
			splitString[i] = strings.ToLower(word)
		} else {
			splitString[i] = strings.ToUpper(word[0:1]) + word[1:]
		}
	}
	return strings.Join(splitString, "")
}

func (g *Generator) loadDev(ctx context.Context) error {
	if g.DeviceID == "" {
		devices, err := g.List(ctx)
		if err != nil {
			return fmt.Errorf("device ID must be provided.\n  error listing devices: %v", err)
		}
		msg := "device ID must be provided.\n  devices:\n"
		msg += "    ID: Name\n"
		for _, dev := range devices {
			msg += fmt.Sprintf("    %s: %s\n", dev.ID, dev.Name)
		}
		return errors.New(msg)
	}
	dev, err := g.Load(ctx, g.DeviceID, nil)
	if err != nil {
		return fmt.Errorf("error loading device: %v", err)
	}
	g.dev = &DevParams{
		Device: dev,
		Params: g.Params,
	}
	return nil
}

func (g *Generator) makeInstance(ctx context.Context) error {
	g.dev.Instance = &device.Instance{
		ID:       g.dev.ID,
		Name:     g.dev.Name,
		Language: string(g.Language),
		Address:  "localhost",
		Port:     g.Port,
	}
	return nil
}

func (g *Generator) genFromTemplate(outPath, tPath string) error {
	t, err := templateDir.ReadFile(tPath)
	if err != nil {
		return fmt.Errorf("error reading template: %v", err)
	}
	tmpl := template.Must(template.New(tPath).Funcs(template.FuncMap{
		"pascal":          sentenceToPascalCase,
		"snake":           sentenceToSnakeCase,
		"camel":           sentenceToCamelCase,
		"pascalFromSnake": toPascalFromSnake,
		"langType":        langType(g.Language),
	}).Parse(string(t)))
	outFile := strings.Replace(strings.TrimSuffix(outPath, filepath.Ext(outPath)), "{dot}", ".", 1)
	f, err := os.Create(outFile)
	defer func() {
		err := f.Close()
		if err != nil {
			fmt.Printf("error closing file: %v", err)
		}
	}()
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	err = tmpl.Execute(f, g.dev)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}
	return nil
}

func (g *Generator) crawlDir(ctx context.Context, outPath, tPath string, parDir fs.DirEntry) error {
	if !parDir.IsDir() {
		return nil
	}
	rPath := filepath.Join(outPath, parDir.Name())
	dPath := filepath.Join(tPath, parDir.Name())
	subDir, err := templateDir.ReadDir(dPath)
	if err != nil {
		return fmt.Errorf("error reading directory: %v", err)
	}
	for _, file := range subDir {
		if file.IsDir() {
			err := g.crawlDir(ctx, dPath, rPath, file)
			if err != nil {
				return err
			}
			continue
		}
		fName := strings.Replace(file.Name(), "{dot}", ".", 1)
		fPath := filepath.Join(rPath, fName)
		tfPath := filepath.Join(dPath, fName)
		err := g.genFromTemplate(fPath, tfPath)
		if err != nil {
			return err
		}
	}
	err = os.MkdirAll(rPath, 0755)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	return nil
}

func ToSnakeCaseFromSentence(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

func toPascalFromSnake(s string) string {
	splitString := strings.Split(s, "_")
	for i, word := range splitString {
		splitString[i] = strings.ToUpper(word[0:1]) + word[1:]
	}
	return strings.Join(splitString, "")
}

func (g *Generator) makeDirTree(ctx context.Context) error {
	if g.OutDir == "" || g.OutDir == "." {
		g.OutDir = ToSnakeCaseFromSentence(g.dev.Name)
	}
	err := os.MkdirAll(g.OutDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating output directory: %v", err)
	}
	tplPath := filepath.Join("template", g.dev.Device.Language)
	langDir, err := templateDir.ReadDir(tplPath)
	if err != nil {
		return fmt.Errorf("error reading directory: %v", err)
	}
	for _, file := range langDir {
		if file.IsDir() {
			err := g.crawlDir(ctx, g.OutDir, tplPath, file)
			if err != nil {
				return err
			}
			continue
		}
		fPath := filepath.Join(g.OutDir, file.Name())
		tfPath := filepath.Join(tplPath, file.Name())
		err := g.genFromTemplate(fPath, tfPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) saveDev(ctx context.Context) error {
	srv := yaml.Service{}
	err := os.MkdirAll(g.OutDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating output directory: %v", err)
	}
	nf, err := os.Create(filepath.Join(g.OutDir, "device.yaml"))
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer func() {
		err := nf.Close()
		if err != nil {
			fmt.Printf("error closing file: %v", err)
		}
	}()
	err = srv.Flush(nf, g.Seen[g.dev.ID])
	if err != nil {
		return err
	}
	instanceID, err := g.Flush(ctx, g.dev.Device)
	if err != nil {
		return fmt.Errorf("error saving device: %v", err)
	}
	if instanceID != g.dev.Instance.ID {
		g.dev.Instance.ID = instanceID
	}
	fmt.Printf("Saved device: %s\n", instanceID)
	return nil
}

func NewGenerator(client *prisma.DeviceClient, params *Params) *Generator {
	return &Generator{
		Params:       params,
		DeviceClient: client,
	}
}
