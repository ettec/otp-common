package loadbalancing

import (
	"fmt"
	"github.com/emicklei/go-restful/log"
	"github.com/ettec/otp-common/executionvenue"
	"github.com/ettec/otp-common/k8s"
	"github.com/ettec/otp-common/model"
	"github.com/ettec/otp-common/orderstore"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
)

func NewLoadBalancedOrderCache( store orderstore.OrderStore, execVenueMic string, id string) (*executionvenue.OrderCache, error) {


	podOrdinal, err := GetStatefulSetPodOrdinalFromName(id)
	if err != nil {
		return nil,fmt.Errorf("failed to get pod ordnial from pod name: %v", err)
	}

	micToBalancingPods, err := GetMicToStatefulPodAddresses("execution-venue")

	if err != nil {
		return nil,fmt.Errorf("failed to get execution venue balancing pods: %v", err)
	}

	var numVenuesForMic int32
	if pods, ok := micToBalancingPods[execVenueMic]; ok {
		numVenuesForMic = int32(len(pods))
	} else {
		return nil,fmt.Errorf("no execution venue balancing pods found for Mic:%v", execVenueMic)
	}

	loadOrder := func(order *model.Order) bool {

		if execVenueMic == order.GetOwnerId() {
			ordinal := GetBalancingOrdinal(order.ListingId, numVenuesForMic)
			if ordinal == podOrdinal {
				return true
			}
		}
		return false
	}

	orderCache, err := executionvenue.NewOrderCache(store, loadOrder)
	if err != nil {
		return nil,fmt.Errorf("failed to create order cache:%v", err)
	}
	return orderCache, nil
}



func GetBalancingOrdinal(listingId int32, numStatefulServices int32) int {
	ordinal := int(listingId - (listingId/numStatefulServices)*numStatefulServices)
	return ordinal
}

type BalancingStatefulPod struct {
	TargetAddress string
	Ordinal       int
}

func GetMicToStatefulPodAddresses(serviceType string) (map[string][]BalancingStatefulPod, error) {
	micToTargetAddress := map[string][]BalancingStatefulPod{}

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
		micToTargetAddress[mic] = append(micToTargetAddress[mic], BalancingStatefulPod{TargetAddress: targetAddress, Ordinal: ordinal})
	}
	return micToTargetAddress, nil
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
	//podId := string(r[idx+1:len(p)])

	targetAddress := pod.Name + "." + serviceName + ":" + strconv.Itoa(int(podPort))
	return targetAddress, nil
}
