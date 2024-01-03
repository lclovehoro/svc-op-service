package util

import (
	"context"
	"flag"
	"path/filepath"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

type clientSet struct {
	*kubernetes.Clientset
}

type deploymentsSet struct {
	*v1.Deployment
}

func NewClientSet() (*kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

func (c *clientSet) ListDeployments(ctx context.Context, namespace string) ([]string, error) {
	var nameLists []string
	deploymentsList, err := c.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	for _, item := range deploymentsList.Items {
		nameLists = append(nameLists, item.GetName())
	}

	return nameLists, err
}

func (c *clientSet) GetDeployments(namespace, name string, ctx context.Context) (*v1.Deployment, error) {
	deploymentsClient := c.AppsV1().Deployments(namespace)

	deployment, err := deploymentsClient.Get(ctx, name, metav1.GetOptions{})
	return deployment, err
}

func (d *deploymentsSet) GetResources() corev1.ResourceRequirements {
	resource := d.Spec.Template.Spec.Containers[0].Resources
	return resource
}
