package main

import (
	"flag"
	"github.com/resouer/k8s-controller-custom-resource/pkg/signals"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/golang/glog"
	"k8s.io/client-go/informers"
	"time"
	"k8s.io/client-go/kubernetes"
	"woniuxiaoan/p47-network-controller/pkg/client/informers/externalversions"
	"woniuxiaoan/p47-network-controller/pkg/client/clientset/versioned"
)

var (
	masterURL string
	kubeconfig string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "")
	flag.StringVar(&masterURL, "master", "","")
}

func main() {
	flag.Parse()

	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig:  %s", err.Error())
	}

	kubeset, err := kubernetes.NewForConfig(cfg)
	networkclientset, _ := versioned.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error gen kubeset: %s", err.Error())
	}

	informerFactory := informers.NewSharedInformerFactory(kubeset, time.Second * 30)
	networkFactory := externalversions.NewSharedInformerFactory(networkclientset, time.Second * 30)

	controller := NewController(kubeset,
		networkclientset,
		informerFactory.Core().V1().Pods(),
		informerFactory.Core().V1().Nodes(),
		informerFactory.Apps().V1().Deployments(),
		networkFactory.P47().V1().Networks())

	go informerFactory.Start(stopCh)
	go networkFactory.Start(stopCh)

	if err := controller.Run(16,stopCh); err != nil {
		glog.Fatalf("Controller exited")
	}

}