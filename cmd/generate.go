//go:generate go run generate.go

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gertd/go-pluralize"
	"github.com/stoewer/go-strcase"
)

// ConvertToPascalCase converts a string to PascalCase.
func ConvertToPascalCase(input string) string {
	return strcase.UpperCamelCase(input)
}

// Helper functions for the template engine.
var templateFuncs = template.FuncMap{
	"PascalCase": ConvertToPascalCase,
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run cmd/generate.go <generate-type [module, model]> <name>")
		return
	}

	if os.Args[1] == "module" {
		generateModule()
	} else if os.Args[1] == "model" {
		generateModel()
	} else {
		fmt.Println("Invalid generate type! [module|model]")
	}
}

func generateModel() {
	modelName := os.Args[2]
	pluralize := pluralize.NewClient()
	modelPluralName := pluralize.Plural(modelName)
	modelPath := filepath.Join("src", "models")
	templatesDir := filepath.Join("cmd", "gen", "model_templates")

	// Create model directory
	err := os.MkdirAll(modelPath, os.ModePerm)
	if err != nil {
		log.Printf("Failed to create model directory: %v\n", err)
		return
	}

	// Walk through the templates directory
	if err = filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if info.IsDir() {
			return nil
		}

		// Generate file
		fileName := strings.TrimSuffix(info.Name(), ".tpl")
		modelFileName := fmt.Sprintf("%s_%s", modelName, fileName)
		targetFile := filepath.Join(modelPath, modelFileName)

		tmplContent, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		tmpl, err := template.New(modelFileName).Funcs(templateFuncs).Parse(string(tmplContent))
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, map[string]string{
			"ModuleName":             modelName,
			"BracketModuleNameID":    fmt.Sprintf("{%s_id}", modelName),
			"ModulePluralName":       modelPluralName,
			"CamelModuleName":        strcase.LowerCamelCase(modelName),
			"PascalModuleName":       strcase.UpperCamelCase(modelName),
			"CamelModulePluralName":  strcase.LowerCamelCase(modelPluralName),
			"PascalModulePluralName": strcase.UpperCamelCase(modelPluralName),
		})
		if err != nil {
			return err
		}

		err = os.WriteFile(targetFile, buf.Bytes(), os.ModePerm)
		if err != nil {
			return err
		}

		log.Printf("Generated: %s\n", targetFile)
		return nil
	}); err != nil {
		log.Printf("Error generating model: %v\n", err)
	}
}

func generateModule() {
	moduleName := os.Args[2]
	pluralize := pluralize.NewClient()
	modulePluralName := pluralize.Plural(moduleName)
	modulePath := filepath.Join("src", "modules", fmt.Sprintf("%s_module", moduleName))

	// Create module directory
	err := os.MkdirAll(modulePath, os.ModePerm)
	if err != nil {
		log.Printf("Failed to create module directory: %v\n", err)
		return
	}

	// Create module dtos directory
	err = os.MkdirAll(filepath.Join(modulePath, "dtos"), os.ModePerm)
	if err != nil {
		log.Printf("Failed to create module dtos directory: %v\n", err)
		return
	}

	templatesDir := filepath.Join("cmd", "gen", "module_templates")
	err = filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if info.IsDir() {
			return nil
		}

		// Generate file
		fileName := strings.TrimSuffix(info.Name(), ".tpl")
		moduleFileName := fmt.Sprintf("%s_%s", moduleName, fileName)
		var targetFile string

		switch fileName {
		case "dtos.go":
			targetFile = filepath.Join(modulePath, "dtos", fileName)
		default:
			targetFile = filepath.Join(modulePath, moduleFileName)
		}

		tmplContent, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		tmpl, err := template.New(moduleFileName).Funcs(templateFuncs).Parse(string(tmplContent))
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, map[string]string{
			"ModuleName":             moduleName,
			"BracketModuleNameID":    fmt.Sprintf("{%s_id}", moduleName),
			"ModulePluralName":       modulePluralName,
			"CamelModuleName":        strcase.LowerCamelCase(moduleName),
			"PascalModuleName":       strcase.UpperCamelCase(moduleName),
			"CamelModulePluralName":  strcase.LowerCamelCase(modulePluralName),
			"PascalModulePluralName": strcase.UpperCamelCase(modulePluralName),
		})
		if err != nil {
			return err
		}

		err = os.WriteFile(targetFile, buf.Bytes(), os.ModePerm)
		if err != nil {
			return err
		}

		log.Printf("Generated: %s\n", targetFile)
		return nil
	})

	if err != nil {
		log.Printf("Error generating module: %v\n", err)
	}
}
