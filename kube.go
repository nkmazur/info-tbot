package main

import (
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"
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
	NsName     string
}

type imageCount struct {
	Image string
	Count int
}
type imageCountList []imageCount

func GetImages() ([]imageCount, []imageCount, error) {

	deployments, err := svc.kube.ExtensionsV1beta1().Deployments(apiv1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Can't get deployments list")
		return nil, nil, fmt.Errorf("Can't get deployments list - %v\n", err)
	}

	activeImages := make(map[string]int)
	notActiveImages := make(map[string]int)

	for _, v := range deployments.Items {
		if v.Status.AvailableReplicas == *v.Spec.Replicas && *v.Spec.Replicas != 0 {
			activeImages[v.Spec.Template.Spec.Containers[0].Image]++
		} else {
			notActiveImages[v.Spec.Template.Spec.Containers[0].Image]++
		}
	}

	activeList := make(imageCountList, len(activeImages))
	notActiveList := make(imageCountList, len(notActiveImages))
	i := 0
	for k, v := range activeImages {
		activeList[i] = imageCount{k, v}
		i++
	}
	sort.Slice(activeList, func(i, j int) bool { return activeList[i].Count > activeList[j].Count })
	i = 0
	for k, v := range notActiveImages {
		notActiveList[i] = imageCount{k, v}
		i++
	}
	sort.Slice(notActiveList, func(i, j int) bool { return notActiveList[i].Count > notActiveList[j].Count })

	return activeList, notActiveList, nil
}

func GetDeployCount() (int, error) {
	deployments, err := svc.kube.ExtensionsV1beta1().Deployments(apiv1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Can't get deployments list")
		return 0, fmt.Errorf("Can't get deployments list - %v\n", err)
	}

	return len(deployments.Items), nil

}

func GetNsCount() (int, error) {
	ns, err := svc.kube.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Can't get namespaces list")
		return 0, fmt.Errorf("Can't get namespaces list - %v\n", err)
	}

	return len(ns.Items), nil
}

func GetUserDeploys(info []UserInfo) (map[string][]DeploysInfo, error) {

	all := make(map[string][]DeploysInfo)

	for _, v := range info {
		deploys, err := svc.kube.ExtensionsV1beta1().Deployments(v.NamespaceId).List(metav1.ListOptions{})
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Error("Can't get deployments list")
			return nil, fmt.Errorf("Can't get deployments list - %v\n", err)
		}
		for _, deploy := range deploys.Items {
			key := fmt.Sprintf("%s (%s)", v.Label, v.NamespaceId)
			var active bool

			if deploy.Status.AvailableReplicas == *deploy.Spec.Replicas && deploy.Status.UnavailableReplicas == 0 {
				active = true
			}

			all[key] = append(all[key], DeploysInfo{
				Replicas:   *deploy.Spec.Replicas,
				Image:      deploy.Spec.Template.Spec.Containers[0].Image,
				IsActive:   active,
				Running:    deploy.Status.AvailableReplicas,
				NotRunning: deploy.Status.UnavailableReplicas,
				NsID:       v.NamespaceId,
				NsName:     v.Label,
			})
		}
	}

	return all, nil
}

func kubeErrors() (string, error) {
	pods, err := svc.kube.CoreV1().Pods("").List(metav1.ListOptions{})
	log.WithFields(log.Fields{
		"Error": err,
	}).Error("Can't get pods list")

	var text string

	for _, pod := range pods.Items {
		for _, state := range pod.Status.ContainerStatuses {
			if state.State.Waiting != nil {
				text += fmt.Sprintf("Namespace: %v:\nPod:%v Reason: %v, Message: %v\n\n", pod.Namespace, state.Name, state.State.Waiting.Reason, state.State.Waiting.Message)
			}
			if pod.DeletionTimestamp != nil {
				text += fmt.Sprintf("Namespace: %v:\nPod:%v Terminating\n\n", pod.Namespace, state.Name)
			}
		}
	}
	return text, nil
}
