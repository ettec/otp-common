// Code to interact with the apis of standard otp service interfaces
package api

import (
	"fmt"
	"github.com/ettec/otp-common/api/executionvenue"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"strconv"
	"time"
)

func GetOrderRouter(clientSet *kubernetes.Clientset, maxConnectRetrySecs time.Duration) (executionvenue.ExecutionVenueClient, error) {
	namespace := "default"
	list, err := clientSet.CoreV1().Services(namespace).List(v1.ListOptions{
		LabelSelector: "app=order-router",
	})

	if err != nil {
		panic(err)
	}

	var client executionvenue.ExecutionVenueClient

	for _, service := range list.Items {

		var podPort int32
		for _, port := range service.Spec.Ports {
			if port.Name == "api" {
				podPort = port.Port
			}
		}

		if podPort == 0 {
			log.Printf("ignoring order router service as it does not have a port named executionvenue, service: %v", service)
			continue
		}

		targetAddress := service.Name + ":" + strconv.Itoa(int(podPort))

		log.Printf("connecting to order router service %v at: %v", service.Name, targetAddress)

		conn, err := grpc.Dial(targetAddress, grpc.WithInsecure(), grpc.WithBackoffMaxDelay(maxConnectRetrySecs))

		if err != nil {
			panic(err)
		}

		client = executionvenue.NewExecutionVenueClient(conn)
		break
	}

	if client == nil {
		return nil, fmt.Errorf("failed to find order router")
	}

	return client, nil
}
