package main

import (
	"context"
	"fmt"
	"strings"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func renderKubernetesOverview() (string, error) {
	// config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	// }
	config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create clientset: %w", err)
	}

	ctx := context.TODO()
	deployments, err := clientset.AppsV1().Deployments("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list deployments: %w", err)
	}

	services, _ := clientset.CoreV1().Services("default").List(ctx, metav1.ListOptions{})
	ingresses, _ := clientset.NetworkingV1().Ingresses("default").List(ctx, metav1.ListOptions{})
	pods, _ := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})

	var sb strings.Builder
	sb.WriteString(`<table>
<thead><tr>
<th>Deployment</th>
<th>Pods</th>
<th>Service</th>
<th>Ingress</th>
<th>Namespace</th>
</tr></thead>
<tbody>`)

	for _, d := range deployments.Items {
		var matchedPods []corev1.Pod
		for _, p := range pods.Items {
			if matchLabels(d.Spec.Selector.MatchLabels, p.Labels) {
				matchedPods = append(matchedPods, p)
			}
		}

		var serviceName, ingressName string
		for _, s := range services.Items {
			if matchLabels(d.Spec.Selector.MatchLabels, s.Spec.Selector) {
				serviceName = s.Name
				break
			}
		}

		for _, i := range ingresses.Items {
			for _, rule := range i.Spec.Rules {
				if rule.HTTP == nil {
					continue
				}
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service != nil && path.Backend.Service.Name == serviceName {
						ingressName = i.Name
						break
					}
				}
				if ingressName != "" {
					break
				}
			}
			if ingressName != "" {
				break
			}
		}

		sb.WriteString("<tr>")
		sb.WriteString(fmt.Sprintf("<td>%s</td>", d.Name))
		sb.WriteString(fmt.Sprintf("<td>%d</td>", len(matchedPods)))
		sb.WriteString(fmt.Sprintf("<td>%s</td>", serviceName))
		sb.WriteString(fmt.Sprintf("<td>%s</td>", ingressName))
		sb.WriteString(fmt.Sprintf("<td>%s</td>", d.Namespace))
		sb.WriteString("</tr>")
	}

	sb.WriteString("</tbody></table>")
	return sb.String(), nil
}

func matchLabels(selector map[string]string, labels map[string]string) bool {
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}

func renderDeploymentJSONTable() (string, error) {
	config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
	if err != nil {
		return "", fmt.Errorf("load kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("clientset: %w", err)
	}

	ctx := context.TODO()
	deployments, err := clientset.AppsV1().Deployments("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("list deployments: %w", err)
	}

	services, _ := clientset.CoreV1().Services("default").List(ctx, metav1.ListOptions{})
	ingresses, _ := clientset.NetworkingV1().Ingresses("default").List(ctx, metav1.ListOptions{})
	pods, _ := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})

	var result []map[string]interface{}

	for _, d := range deployments.Items {
		item := map[string]interface{}{
			"deployment": d,
			"pods":       []corev1.Pod{},
			"service":    nil,
			"ingress":    nil,
		}

		// pods ที่ตรงกับ deployment
		var matchedPods []corev1.Pod
		for _, p := range pods.Items {
			if matchLabels(d.Spec.Selector.MatchLabels, p.Labels) {
				matchedPods = append(matchedPods, p)
			}
		}
		item["pods"] = matchedPods

		// service ที่ตรงกับ deployment
		var serviceName string
		for _, s := range services.Items {
			if matchLabels(d.Spec.Selector.MatchLabels, s.Spec.Selector) {
				item["service"] = s
				serviceName = s.Name
				break
			}
		}

		// ingress ที่ใช้ service นี้
		for _, i := range ingresses.Items {
			for _, rule := range i.Spec.Rules {
				if rule.HTTP == nil {
					continue
				}
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service != nil && path.Backend.Service.Name == serviceName {
						item["ingress"] = i
						break
					}
				}
			}
		}

		result = append(result, item)
	}

	raw, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	html := `<h2>📦 Deployment + Resources (JSON)</h2><pre style="background:#f0f0f0;padding:1em;">` +
		string(raw) + "</pre>"

	return html, nil
}
