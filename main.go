package main

import (
	"html/template"
	"log"
	"net/http"
)

var templates = template.Must(template.ParseFiles("templates/index.html"))

func main() {
	http.HandleFunc("/", indexHandler)
	log.Println("Server running on http://localhost:4000")
	http.ListenAndServe(":4000", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	html, err := renderKubernetesOverview()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"HTML": template.HTML(html),
	})
}
