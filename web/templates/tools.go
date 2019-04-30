package templates

// courtesy to https://hackernoon.com/golang-template-2-template-composition-and-how-to-organize-template-files-4cb40bcdf8f6

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"upsilon_cities_go/config"

	"github.com/oxtoacart/bpool"
)

var mainTmpl = `{{define "main" }} {{ template "base" . }} {{ end }}`

type TemplateConfig struct {
	TemplateLayoutPath  string
	TemplateIncludePath string
}

var templateConfig TemplateConfig

var templates map[string]*template.Template
var layouts map[string][]string
var bufpool *bpool.BufferPool

func loadConfiguration() {
	templateConfig.TemplateLayoutPath = config.WEB_LAYOUTS
	templateConfig.TemplateIncludePath = config.WEB_TEMPLATES
}

// LoadTemplates initiates available templates.
func LoadTemplates() {
	if templates == nil {
		loadConfiguration()
		templates = make(map[string]*template.Template)
		layouts = make(map[string][]string)
	}

	mainTemplate := template.New("main")
	mainTemplate, err := mainTemplate.Parse(mainTmpl)

	err = filepath.Walk(templateConfig.TemplateLayoutPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("Templates: prevent panic by handling failure accessing a path %q: %v\n", templateConfig.TemplateLayoutPath, err)
			return err
		}
		if strings.HasSuffix(info.Name(), ".tmpl") {
			log.Printf("Templates: Added shared of file : %s\n", path)

			layoutfullname := strings.TrimLeft(strings.Replace(path, templateConfig.TemplateLayoutPath, "", 1), "\\")
			layoutbase := strings.TrimRight(strings.Replace(layoutfullname, info.Name(), "", 1), "/")

			layouts[layoutbase] = append(layouts[layoutbase], path)

		}

		return nil
	})

	if err != nil {
		log.Fatalf("Templates: Failed to load shared templates: %s\n", err)
	}

	err = filepath.Walk(templateConfig.TemplateIncludePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("Templates: prevent panic by handling failure accessing a path %q: %v\n", templateConfig.TemplateIncludePath, err)
			return err
		}
		if strings.HasSuffix(info.Name(), ".tmpl") {

			templatefullname := strings.Replace(strings.TrimLeft(strings.Replace(path, templateConfig.TemplateIncludePath, "", 1), "\\"), ".html.tmpl", "", 1)
			templatebase := strings.TrimRight(strings.Replace(templatefullname, info.Name(), "", 1), "/")

			templates[templatefullname], err = mainTemplate.Clone()
			if err != nil {
				log.Fatalf("Templates: Failed to clone mainTemplate: %s\n", err)
			}

			files := append(layouts[""], append(layouts[templatebase], path)...)
			templates[templatefullname] = template.Must(templates[templatefullname].ParseFiles(files...))

			log.Printf("Templates: Loaded template : %s as %s\n", path, templatefullname)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Templates: Failed to load templates: %s", err)
	}

	log.Println("Templates: Loading successful")

	bufpool = bpool.NewBufferPool(64)
	log.Println("Templates: buffer allocation successful")
}

// RenderTemplate render provided templates name. Template name must match path eg: garden/index
func RenderTemplate(w http.ResponseWriter, name string, data interface{}) {
	tmpl, ok := templates[name]
	if !ok {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		log.Fatalf("Templates: The template %s does not exist. Can't render", name)
		return
	}

	buf := bufpool.Get()
	defer bufpool.Put(buf)

	err := tmpl.Execute(buf, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}
