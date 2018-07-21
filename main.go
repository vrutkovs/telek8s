package main

import (
	"fmt"
	"reflect"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

var tgBot TGBot

func podAdded(obj interface{}) {
	pod := obj.(*v1.Pod)
	go tgBot.sendMessage(
		"Pod created: *" + pod.Name + "* in namespace " + pod.Namespace)
}

func podDeleted(obj interface{}) {
	pod := obj.(*v1.Pod)
	go tgBot.sendMessage(
		"Pod deleted: *" + pod.Name + "* in namespace " + pod.Namespace)
}

func podUpdated(oldObj, newObj interface{}) {
	// Get pod objects
	oldPod := oldObj.(*v1.Pod)
	newPod := newObj.(*v1.Pod)

	// Make a diff
	diff := jsonDiff(oldPod, newPod)

	if diff != "" {
		// Show JSON diff
		go tgBot.sendMessage(
			"Pod changed status: *" + oldPod.Name + "* in namespace " + oldPod.Namespace + " :\n" + diff)
	}
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func jsonDiff(oldPod *v1.Pod, newPod *v1.Pod) string {
	oldPodStatus := oldPod.Status
	newPodStatus := newPod.Status

	oldReflect := reflect.ValueOf(&oldPodStatus).Elem()
	newReflect := reflect.ValueOf(&newPodStatus).Elem()

	oldValues := make([]interface{}, oldReflect.NumField())
	newValues := make([]interface{}, newReflect.NumField())

	for i := 0; i < oldReflect.NumField(); i++ {
		oldValues[i] = oldReflect.Field(i).Interface()
		newValues[i] = newReflect.Field(i).Interface()
	}

	return makeDiff(0, oldReflect, oldValues, newValues)
}

func makeDiff(indent int, oldReflect reflect.Value, oldValues []interface{}, newValues []interface{}) string {
	result := ""
	for i := 0; i < len(oldValues); i++ {
		result += "\n------"
		oldValue := oldValues[i]
		newValue := newValues[i]
		fieldName := oldReflect.Type().Field(i).Name
		if _, ok := oldValue.(string); ok {
			if oldValue != newValue {
				result += fmt.Sprintf("%s:\n  -%s\n  +%s\n", fieldName, oldValue, newValue)
			}
		}
	}
	return result
}

func watchPods(clientset *kubernetes.Clientset) {
	var client = clientset.Core().RESTClient()
	listWatch := cache.NewListWatchFromClient(
		client, "pods", "",
		fields.Everything())
	_, controller := cache.NewInformer(
		listWatch, &v1.Pod{},
		time.Second*0, cache.ResourceEventHandlerFuncs{
			AddFunc:    podAdded,
			DeleteFunc: podDeleted,
			UpdateFunc: podUpdated,
		},
	)
	go controller.Run(wait.NeverStop)
}

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	tgBot.init()

	watchPods(clientset)

	// Loop forever:
	for {
		time.Sleep(time.Second)
	}
}
