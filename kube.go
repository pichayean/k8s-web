package main

import (
    "context"
    "log"

    appsv1 "k8s.io/api/apps/v1"
    netv1 "k8s.io/api/networking/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/rest"
)

func getClientSet() *kubernetes.Clientset {
    // ใช้ใน-cluster config ถ้าอยู่ใน Pod
    config, err := rest.InClusterConfig()
    if err != nil {
        // ถ้าไม่อยู่ใน cluster ใช้ local config (~/.kube/config)
        kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
            clientcmd.NewDefaultClientConfigLoadingRules(),
            &clientcmd.ConfigOverrides{},
        )
        config, err = kubeconfig.ClientConfig()
        if err != nil {
            log.Fatalf("Error loading kubeconfig: %v", err)
        }
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatalf("Error creating clientset: %v", err)
    }
    return clientset
}

func listDeployments() ([]appsv1.Deployment, error) {
    client := getClientSet()
    deployments, err := client.AppsV1().Deployments("default").List(context.TODO(), metav1ListOptions())
    if err != nil {
        return nil, err
    }
    return deployments.Items, nil
}

func listIngresses() ([]netv1.Ingress, error) {
    client := getClientSet()
    ingresses, err := client.NetworkingV1().Ingresses("default").List(context.TODO(), metav1ListOptions())
    if err != nil {
        return nil, err
    }
    return ingresses.Items, nil
}

func metav1ListOptions() interface{ /* empty */ } {
    // ใช้เพื่อให้แยก Logic ได้ง่ายขึ้น (แต่จริงๆ return metav1.ListOptions{} ก็ได้เลย)
    return nil
}
