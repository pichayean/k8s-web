package handlers

import (
	"net/http"
	"text/template"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var tpl = template.Must(template.ParseFiles("templates/deployments.html"))

func DeploymentsHandler(w http.ResponseWriter, r *http.Request) {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		http.Error(w, "Failed to load kubeconfig", 500)
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		http.Error(w, "Failed to connect to cluster", 500)
		return
	}

	deployments, err := clientset.AppsV1().Deployments("default").List(r.Context(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, "Failed to list deployments", 500)
		return
	}

	tpl.Execute(w, deployments.Items)
}
