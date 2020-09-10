package loadbalancing

import (
	"fmt"
	"github.com/emicklei/go-restful/log"
	"github.com/ettec/otp-common/k8s"
	"github.com/ettec/otp-common/model"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
)

func GetBalancingOrdinal(listing *model.Listing, numStatefulServices int32) int {
	ordinal := int(listing.Id - (listing.Id/numStatefulServices)*numStatefulServices)
	return ordinal
}

type StatefulPodAddress struct {
	targetAddress string
	ordinal int
}

func GetMicToStatefulPodAddresses( serviceType string) (map[string][]StatefulPodAddress, error) {
	micToTargetAddress := map[string][]StatefulPodAddress{}

	clientSet := k8s.GetK8sClientSet(false)

	namespace := "default"
	list, err := clientSet.CoreV1().Pods(namespace).List(v1.ListOptions{
		LabelSelector: "servicetype=" + serviceType,
	})

	if err != nil {
		return nil, err
	}

	log.Printf("found %v stateful pods with service type %v", len(list.Items), serviceType)

	for _, pod := range list.Items {
		const micLabel = "mic"
		if _, ok := pod.Labels[micLabel]; !ok {
			log.Printf("ignoring stateful pod as it does not have a mic label, pod: %v", pod)
			continue
		}

		mic := pod.Labels[micLabel]

		targetAddress, err := getStatefulSetMemberAddress(pod)
		if err != nil {
			return nil, fmt.Errorf("failed to get stateful pod address:%v", err)
		}

		ordinal, err := getStatefulSetPodOrdinal(pod)
		micToTargetAddress[mic] = append(micToTargetAddress[mic], StatefulPodAddress{targetAddress: targetAddress, ordinal: ordinal})
	}
	return micToTargetAddress, nil
}

func getStatefulSetPodOrdinal(pod v12.Pod) (int, error) {
	idx := strings.LastIndex(pod.Name, "-")
	r := []rune(pod.Name)
	podOrd := string(r[idx+1:len(pod.Name)])
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
	//podId := string(r[idx+1:len(p)])

	targetAddress := pod.Name + "." + serviceName + ":" + strconv.Itoa(int(podPort))
	return targetAddress, nil
}
