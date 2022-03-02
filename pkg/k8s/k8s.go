package k8s

import (
	"fmt"

	"github.com/ergoapi/zlog"
	"github.com/ysicing/ingress_exporter/internal/kube"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/informers"
	networkinginformers "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var K8SClient kubernetes.Interface

func init() {
	var err error
	kubecfg := &kube.ClientConfig{}
	K8SClient, err = kube.New(kubecfg)
	if err != nil {
		panic(err)
	}
}

func KV() string {
	kv, err := K8SClient.Discovery().ServerVersion()
	if err != nil {
		return "unknow"
	}
	return kv.String()
}

func NewNamespaceControlller(i informers.SharedInformerFactory) *MonitorControlller {
	ingresssInformer := i.Networking().V1().Ingresses()
	c := &MonitorControlller{
		informerFactory:  i,
		ingresssInformer: ingresssInformer,
	}
	ingresssInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.ingressadd,
			DeleteFunc: c.ingressdel,
			// UpdateFunc: c.nsupdate,
		},
	)
	return c
}

type MonitorControlller struct {
	informerFactory  informers.SharedInformerFactory
	ingresssInformer networkinginformers.IngressInformer
}

func (c *MonitorControlller) ingressadd(obj interface{}) {
	ings := obj.(*networkingv1.Ingress)
	for _, ing := range ings.Spec.Rules {
		if ing.Host == "" {
			continue
		}
		zlog.Info("ingress add: %s", ing.Host)
	}
}

func (c *MonitorControlller) ingressdel(obj interface{}) {
	ings := obj.(*networkingv1.Ingress)
	for _, ing := range ings.Spec.Rules {
		if ing.Host == "" {
			continue
		}
		zlog.Info("ingress delete: %s", ing.Host)
	}
}

func (c *MonitorControlller) Run(stopCh chan struct{}) error {
	// Starts all the shared informers that have been created by the factory so
	// far.
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.ingresssInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync")
	}
	return nil
}
