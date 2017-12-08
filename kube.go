package main

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetImages(clientset *kubernetes.Clientset) (activeImages map[string]int, notActiveImages map[string]int, err error) {
	deployments, err := clientset.ExtensionsV1beta1().Deployments(apiv1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("Can't get deployments list - %v\n", err)

	}

	activeImages, notActiveImages = make(map[string]int), make(map[string]int)

	for _, v := range deployments.Items {
		unavailableReplicas := v.Status.UnavailableReplicas
		availableReplicas := v.Status.AvailableReplicas
		replicas := *v.Spec.Replicas
		image := v.Spec.Template.Spec.Containers[0].Image

		if availableReplicas == replicas && unavailableReplicas == 0 {
			activeImages[image]++
		} else {
			notActiveImages[image]++
		}
	}

	return activeImages, notActiveImages, nil
}

func GetDeployCount(clientset *kubernetes.Clientset) (int, error) {
	deployments, err := clientset.ExtensionsV1beta1().Deployments(apiv1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("Can't get deployments list - %v\n", err)

	}

	return len(deployments.Items), nil

}

func GetNsCount(clientset *kubernetes.Clientset) (int, error) {
	ns, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("Can't get namespaces list - %v\n", err)
	}

	return len(ns.Items), nil
}
