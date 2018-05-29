package common

import (
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func CopyObjToNode(obj interface{}) (*v1.Node, error) {
	node := obj.(*v1.Node).DeepCopy()
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
	return node, nil
}
