package main

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/client-go/tools/record"
	"github.com/golang/glog"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/api/core/v1"
	informerCorev1 "k8s.io/client-go/informers/core/v1"
	informerAppsv1 "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/util/runtime"
	"fmt"
	"time"
	"k8s.io/apimachinery/pkg/util/wait"
	p47v1 "woniuxiaoan/p47-network-controller/pkg/client/informers/externalversions/p47/v1"
	"woniuxiaoan/p47-network-controller/pkg/client/clientset/versioned"
	"woniuxiaoan/p47-network-controller/pkg/apis/p47/v1"
)

const controllerAgentName = "p47-network-controller"

type Controller struct {
	kubeclientset kubernetes.Interface
	networkclientset versioned.Interface
	workqueue workqueue.RateLimitingInterface
	recorder record.EventRecorder
	nodeInformer informerCorev1.NodeInformer
	nodeSynced cache.InformerSynced
	deploymentInformer informerAppsv1.DeploymentInformer
	deploymentSynced cache.InformerSynced
	networkInformer p47v1.NetworkInformer
	networkSynced cache.InformerSynced
}

func(c *Controller) addNetworkEventHandler(obj interface{}) {
	var key string
	var err error

	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}

	glog.Info("network add event: ", key)
	c.workqueue.AddRateLimited(key)
}

func(c *Controller) deleteNetworkEventHandler(obj interface{}){
	var key string
	var err error
	if key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
	}

	glog.Info("network delete event: ", key)
	c.workqueue.AddRateLimited(key)
}

func(c *Controller) updateNetworkEventHandler(old, new interface{}) {
	oldDm := old.(*v1.Network)
	newDm := new.(*v1.Network)

	if oldDm.ResourceVersion == newDm.ResourceVersion {
		glog.Infof("network update event ignore, because %s's version(%s) is not change", oldDm.Name, oldDm.ResourceVersion)
		return
	}

	var key string
	var err error

	if key, err = cache.MetaNamespaceKeyFunc(new); err != nil {
		runtime.HandleError(err)
		return
	}

	glog.Info("network update event: ", key)
	c.workqueue.AddRateLimited(key)
}


func NewController (kubeclientset kubernetes.Interface,
	networkclientset versioned.Interface,
	podInformer informerCorev1.PodInformer,
	nodeInformer informerCorev1.NodeInformer,
	deploymentInformer informerAppsv1.DeploymentInformer,
	neworkINnformer p47v1.NetworkInformer) *Controller {

	glog.V(4).Info("Creating Event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component:controllerAgentName})

	controller := &Controller{
		kubeclientset:kubeclientset,
		networkclientset:networkclientset,
		recorder: recorder,
		workqueue:workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "network"),
		networkSynced:neworkINnformer.Informer().HasSynced,
	}

	glog.Info("Setting up event handlers")
	neworkINnformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:controller.addNetworkEventHandler,
		UpdateFunc:controller.updateNetworkEventHandler,
		DeleteFunc:controller.deleteNetworkEventHandler,
	})

	return controller
}

func(c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer func() {
		c.workqueue.ShutDown()
	}()

	glog.Info("Starting CustomController control loop")
	glog.Info("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(stopCh, c.networkSynced); !ok {
		return fmt.Errorf("failed to wait for deployment cache to sync")
	}

	glog.Info("Starting workers")

	for i:=0; i< threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Start workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil

}

func(c *Controller) runWorker(){
	for c.processNextWorkItem(){}
}

func(c *Controller) handleObj(obj interface{}) error {
	defer c.workqueue.Done(obj)
	var key string
	var ok bool

	if key, ok = obj.(string); !ok {
		c.workqueue.Forget(obj)
		err := fmt.Errorf("handle network obj error: %v",obj)
		runtime.HandleError(err)
		return err
	}

	glog.Infof("Handle network %s success", key)

	return nil
}

func(c *Controller) processNextWorkItem() bool {
	network, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	if err := c.handleObj(network); err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}