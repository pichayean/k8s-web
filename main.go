package main

import (
    "html/template"
    "log"
    "net/http"
    "sync"
    "time"
)

type PageData struct {
    Deployments []DeploymentData
    Ingresses   []IngressData
}

type DeploymentData struct {
    Name     string
    Replicas int32
}

type IngressData struct {
    Name string
    Host string
    Path string
}

// 🔁 Caching
var (
    cachedData     PageData
    lastFetched    time.Time
    cacheMutex     sync.Mutex
    cacheDuration  = 15 * time.Second
)

func getCachedData() PageData {
    cacheMutex.Lock()
    defer cacheMutex.Unlock()

    if time.Since(lastFetched) < cacheDuration {
        return cachedData
    }

    deps, _ := listDeployments()
    ingrs, _ := listIngresses()

    var depData []DeploymentData
    for _, d := range deps {
        depData = append(depData, DeploymentData{
            Name:     d.Name,
            Replicas: d.Status.Replicas,
        })
    }

    var ingData []IngressData
    for _, i := range ingrs {
        for _, rule := range i.Spec.Rules {
            for _, path := range rule.HTTP.Paths {
                ingData = append(ingData, IngressData{
                    Name: i.Name,
                    Host: rule.Host,
                    Path: path.Path,
                })
            }
        }
    }

    cachedData = PageData{Deployments: depData, Ingresses: ingData}
    lastFetched = time.Now()
    return cachedData
}

func main() {
    tmpl := template.Must(template.ParseFiles("templates/index.html"))

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        data := getCachedData()
        tmpl.Execute(w, data)
    })

    log.Println("Listening on :4000")
    http.ListenAndServe(":4000", nil)
}
