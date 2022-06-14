package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

//go:embed story_template.html
var storyTemplate string

type StoryArc struct {
	Title   string   `json:"title"`
	Story   []string `json:"story"`
	Options []struct {
		Text string `json:"text"`
		Arc  string `json:"arc"`
	} `json:"options"`
}

type Handler struct {
	arcNameToArc map[string]StoryArc
	storyTmpl    *template.Template
}

func MustNewHandler(arcNameToArc map[string]StoryArc) *Handler {
	storyTmpl, err := template.New("story-arc").Parse(storyTemplate)
	if err != nil {
		panic(fmt.Errorf("parse story template: %w", err))
	}
	return &Handler{
		arcNameToArc: arcNameToArc,
		storyTmpl:    storyTmpl,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	arcName := strings.TrimPrefix(r.URL.Path, "/")
	if arcName == "" {
		arcName = "intro"
	}

	arc, ok := h.arcNameToArc[arcName]
	if !ok {
		http.NotFound(w, r)
		return
	}

	if err := h.storyTmpl.Execute(w, arc); err != nil {
		log.Println(fmt.Errorf("execute story template for arc %q: %w", arcName, err))
	}
}
