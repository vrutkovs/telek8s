package main

import (
	"fmt"
	"reflect"
	"strings"
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

	oldReflect := reflect.ValueOf(oldPodStatus)
	newReflect := reflect.ValueOf(newPodStatus)

	return makeDiff("", 0, oldReflect, newReflect)
}

func makeDiff(prefix string, indent int, oldReflect reflect.Value, newReflect reflect.Value) string {
	indentString := strings.Repeat(" ", indent)
	result := ""

	// TODO: Check for new fields?
	// Get field values from this struct
	for i := 0; i < oldReflect.NumField(); i++ {
		oldValue := oldReflect.Field(i)
		newValue := newReflect.Field(i)
		fieldName := oldReflect.Type().Field(i).Name

		// Check field type and make a diff from subelements if available
		switch oldValue.Kind() {
		case reflect.Struct:
			result += makeDiff("", indent+2, oldValue, newValue)
		case reflect.Array, reflect.Slice:
			for j := 0; j < oldValue.Len(); j++ {
				prefixFieldName := oldValue.Index(j).Type().Name() + "."
				result += makeDiff(prefixFieldName, indent+2, oldValue.Index(j), newValue.Index(j))
			}
		default:
			if oldValue.String() != newValue.String() {
				result += fmt.Sprintf("%s_%s%s_:\n  -%s\n  +%s\n", indentString, prefix, fieldName, oldValue, newValue)
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
