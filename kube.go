package main

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploysInfo struct {
	Replicas   int32
	Image      string
	IsActive   bool
	Running    int32
	NotRunning int32
	NsID       string
}

func GetImages() (activeImages map[string]int, notActiveImages map[string]int, err error) {
	deployments, err := svc.kube.ExtensionsV1beta1().Deployments(apiv1.NamespaceAll).List(metav1.ListOptions{})
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

func GetDeployCount() (int, error) {
	deployments, err := svc.kube.ExtensionsV1beta1().Deployments(apiv1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("Can't get deployments list - %v\n", err)
	}

	return len(deployments.Items), nil

}

func GetNsCount() (int, error) {
	ns, err := svc.kube.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("Can't get namespaces list - %v\n", err)
	}

	return len(ns.Items), nil
}

func GetUserDeploys(info []UserInfo) ([]DeploysInfo, error) {

	var all []DeploysInfo

	for _, v := range info {
		deploys, err := svc.kube.ExtensionsV1beta1().Deployments(v.NamespaceId).List(metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("Can't get deployments list - %v\n", err)
		}
		for _, deploy := range deploys.Items {
			all = append(all, DeploysInfo{
				Replicas:   *deploy.Spec.Replicas,
				Image:      deploy.Spec.Template.Spec.Containers[0].Image,
				Running:    deploy.Status.AvailableReplicas,
				NotRunning: deploy.Status.UnavailableReplicas,
				NsID:       v.NamespaceId,
			})
		}
	}

	return all, nil
}
