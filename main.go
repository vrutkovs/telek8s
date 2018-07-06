package main

import (
	"encoding/json"
	"fmt"
	"time"

	jd "github.com/josephburnett/jd/lib"

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

	// Convert objects to JSON
	oldPodJSON, err := json.Marshal(oldPod)
	if err != nil {
		fmt.Println(err)
		return
	}
	newPodJSON, err := json.Marshal(newPod)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Show JSON diff
	a, _ := jd.ReadJsonString(string(oldPodJSON))
	b, _ := jd.ReadJsonString(string(newPodJSON))
	go tgBot.sendMessage(
		"Pod updated: *" + oldPod.Name + "* in namespace " + oldPod.Namespace + " :\n" + a.Diff(b).Render())
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
