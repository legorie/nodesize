package main

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	nodeinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	nodelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type controller struct {
	clientset       kubernetes.Interface
	nodeLister      nodelisters.NodeLister
	nodeCacheSynced cache.InformerSynced
	queue           workqueue.RateLimitingInterface
}

func newController(clientset kubernetes.Interface, nodeInformer nodeinformers.NodeInformer) *controller {
	c := &controller{
		clientset:       clientset,
		nodeLister:      nodeInformer.Lister(),
		nodeCacheSynced: nodeInformer.Informer().HasSynced,
		queue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "nodesize"),
	}

	nodeInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    handleAdd,
			UpdateFunc: handleUpdate,
		},
	)

	return c
}

func (c *controller) run(ch <-chan struct{}) {
	if !cache.WaitForCacheSync(ch, c.nodeCacheSynced) {
		fmt.Println("waiting for cache to be synced")
	}

	go wait.Until(c.worker, 1*time.Second, ch)
	<-ch
}

func (c *controller) worker() {

}

func handleAdd(obj interface{}) {
	fmt.Println("add was called")
}

func handleUpdate(old1, obj interface{}) {
	// fmt.Println("===Update was called")
	fmt.Println("#################################")
	// fmt.Println(obj.(*v1.Node).Status.Capacity)
	// fmt.Println(obj.(*v1.Node).Status.Images)

	new_Size := calcImageSize(obj.(*v1.Node))
	fmt.Println("New Size:", new_Size)
	old_Size := calcImageSize(old1.(*v1.Node))
	fmt.Println("Old Size:", old_Size)
	// lp: Was not able to find the field name of the struct, hence tried some reverse engineering.
	// at the end, kubectl get node -o yaml, did the trick, it list all the fields which can be used :)
	// fmt.Println(obj.(*v1.Node).Status.Capacity["ephemeral-storage"])
	// val := reflect.Indirect(reflect.ValueOf(obj.(*v1.Node).Status.Capacity["ephemeral-storage"]))
	// fmt.Println(val.Field(0), val.Field(0).Type().Name())

	// oldSize := old1.Images
	if old_Size != new_Size {
		fmt.Println("!!! Node size change by ", new_Size-old_Size)
	}
}

func calcImageSize(node *v1.Node) int64 {
	// fmt.Println("in calcImageSize", node.Status.Images)
	var storage int64
	for _, image := range node.Status.Images {
		storage = storage + image.SizeBytes
	}
	return storage
}

// Trying : https://www.youtube.com/watch?v=QIMz4V9WxVc
// https://github.com/alena1108/kubecon2017/blob/master/main.go
// https://github.com/feiskyer/kubernetes-handbook/blob/master/examples/client/informer/informer.go
