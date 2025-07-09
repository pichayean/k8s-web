package main

import (
    "log"
    "net/http"
    "text/template"

    "github.com/gorilla/mux"
    "k8s-dashboard/handlers" // 👈 เพิ่มตรงนี้
)

var templates *template.Template

func main() {
    templates = template.Must(template.ParseGlob("templates/*.html"))

    r := mux.NewRouter()
    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    r.HandleFunc("/", indexHandler)
    r.HandleFunc("/deployments", handlers.DeploymentsHandler) // 👈 ใช้จาก package
    r.HandleFunc("/deployments/{name}", handlers.DeploymentGraphHandler)

    log.Println("Server running on http://localhost:4000")
    http.ListenAndServe(":4000", r)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    templates.ExecuteTemplate(w, "index.html", nil)
}
