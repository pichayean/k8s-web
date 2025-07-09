package main

import (
    "context"
    "log"

    appsv1 "k8s.io/api/apps/v1"
    netv1 "k8s.io/api/networking/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
)

func getClientSet() *kubernetes.Clientset {
    config, err := rest.InClusterConfig()
    if err != nil {
        kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
            clientcmd.NewDefaultClientConfigLoadingRules(),
            &clientcmd.ConfigOverrides{},
        )
        config, err = kubeconfig.ClientConfig()
        if err != nil {
            log.Fatalf("Failed to load kubeconfig: %v", err)
        }
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatalf("Failed to create clientset: %v", err)
    }

    return clientset
}

func listDeployments() ([]appsv1.Deployment, error) {
    client := getClientSet()
    result, err := client.AppsV1().Deployments("default").List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    return result.Items, nil
}

func listIngresses() ([]netv1.Ingress, error) {
    client := getClientSet()
    result, err := client.NetworkingV1().Ingresses("default").List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    return result.Items, nil
}
