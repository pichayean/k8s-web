func DeploymentGraphHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		http.Error(w, "Failed to load kubeconfig", http.StatusInternalServerError)
		return
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		http.Error(w, "Failed to connect to cluster", http.StatusInternalServerError)
		return
	}

	ctx := context.TODO()
	d, err := clientset.AppsV1().Deployments("default").Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
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
				if rule.HTTP == nil {
					continue
				}
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service != nil && path.Backend.Service.Name == matchedService.Name {
						matchedIngress = &i
						break
					}
				}
				if matchedIngress != nil {
					break
				}
			}
			if matchedIngress != nil {
				break
			}
		}
	}

	// ⚠️ Mermaid syntax ต้องไม่มี indent + มี \n จริง
	graph := "graph TD\n"
	if matchedIngress != nil {
		graph += fmt.Sprintf("Ingress[\"Ingress: %s\"]\n", matchedIngress.Name)
		graph += "Ingress --> Service\n"
	}
	if matchedService != nil {
		graph += fmt.Sprintf("Service[\"Service: %s\"]\n", matchedService.Name)
		graph += "Service --> Deploy\n"
	}
	graph += fmt.Sprintf("Deploy[\"Deployment: %s\"]\n", d.Name)
	for _, p := range pods.Items {
		graph += fmt.Sprintf("Deploy --> %s[\"Pod: %s\"]\n", p.Name, p.Name)
	}

	err = graphTpl.Execute(w, map[string]string{
		"Graph": graph,
		"Name":  name,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
