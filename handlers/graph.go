package handlers

import (
	"context"
	"fmt"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var graphTpl = template.Must(template.ParseFiles("templates/graph.html"))

func deploymentGraphHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

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

	ctx := context.TODO()
	d, err := clientset.AppsV1().Deployments("default").Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, "Deployment not found", 404)
		return
	}

	selector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: d.Spec.Selector.MatchLabels})
	pods, _ := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{LabelSelector: selector})
	services, _ := clientset.CoreV1().Services("default").List(ctx, metav1.ListOptions{})
	ingresses, _ := clientset.NetworkingV1().Ingresses("default").List(ctx, metav1.ListOptions{})

	var matchedService *corev1.Service
	for _, s := range services.Items {
		if matchLabels(d.Spec.Selector.MatchLabels, s.Spec.Selector) {
			matchedService = &s
			break
		}
	}

	var matchedIngress *netv1.Ingress
	if matchedService != nil {
		for _, i := range ingresses.Items {
			for _, rule := range i.Spec.Rules {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service != nil && path.Backend.Service.Name == matchedService.Name {
						matchedIngress = &i
						break
					}
				}
			}
			if matchedIngress != nil {
				break
			}
		}
	}

	graph := "graph TD\n"
	if matchedIngress != nil {
		graph += fmt.Sprintf("  Ingress[\"Ingress: %s\"]\n", matchedIngress.Name)
		graph += fmt.Sprintf("  Ingress --> Service\n")
	}
	if matchedService != nil {
		graph += fmt.Sprintf("  Service[\"Service: %s\"]\n", matchedService.Name)
		graph += fmt.Sprintf("  Service --> Deploy\n")
	}
	graph += fmt.Sprintf("  Deploy[\"Deployment: %s\"]\n", d.Name)
	for _, p := range pods.Items {
		graph += fmt.Sprintf("  Deploy --> %s[\"Pod: %s\"]\n", p.Name, p.Name)
	}

	graphTpl.Execute(w, map[string]string{
		"Graph": graph,
		"Name":  name,
	})
}

func matchLabels(selector map[string]string, labels map[string]string) bool {
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}
