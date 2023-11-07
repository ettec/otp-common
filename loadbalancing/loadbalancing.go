// Package loadbalancing contains utility functions used in load balancing.
package loadbalancing

import (
	"fmt"
	v12 "k8s.io/api/core/v1"
	"strconv"
	"strings"
)

func GetBalancingOrdinal(listingId int32, numStatefulServices int32) int {
	ordinal := int(listingId - (listingId/numStatefulServices)*numStatefulServices)
	return ordinal
}

type BalancingStatefulPod struct {
	TargetAddress string
	Ordinal       int
	Name          string
	Mic           string
}

func GetBalancingStatefulPod(pod v12.Pod) (*BalancingStatefulPod, error) {
	const micLabel = "mic"
	if _, ok := pod.Labels[micLabel]; !ok {
		return nil, fmt.Errorf("ignoring stateful pod as it does not have a mic label, pod: %v", pod)
	}

	mic := pod.Labels[micLabel]

	targetAddress, err := getStatefulSetMemberAddress(pod)
	if err != nil {
		return nil, fmt.Errorf("failed to get stateful pod address:%v", err)
	}

	ordinal, err := getStatefulSetPodOrdinal(pod)

	bsp := &BalancingStatefulPod{TargetAddress: targetAddress,
		Ordinal: ordinal, Name: pod.Name, Mic: mic}
	return bsp, nil
}

func getStatefulSetPodOrdinal(pod v12.Pod) (int, error) {
	return GetStatefulSetPodOrdinalFromName(pod.Name)
}

func GetStatefulSetPodOrdinalFromName(podName string) (int, error) {
	idx := strings.LastIndex(podName, "-")
	r := []rune(podName)
	podOrd := string(r[idx+1 : len(podName)])
	return strconv.Atoi(podOrd)
}

func getStatefulSetMemberAddress(pod v12.Pod) (string, error) {

	var podPort int32
	for _, port := range pod.Spec.Containers[0].Ports {
		if port.Name == "api" {
			podPort = port.ContainerPort
		}
	}

	if podPort == 0 {
		return "", fmt.Errorf("stateful set pod has no api port defined, pod: %v", pod)
	}

	idx := strings.LastIndex(pod.Name, "-")
	r := []rune(pod.Name)
	serviceName := string(r[0:idx])

	targetAddress := pod.Name + "." + serviceName + ":" + strconv.Itoa(int(podPort))
	return targetAddress, nil
}
