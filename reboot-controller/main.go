package main

import (
	"flag"
	"time"

	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/util/wait"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/aaronlevy/kube-controller-demo/common"
)

// TODO(aaron): make configurable and add MinAvailable
const maxUnavailable = 1

func main() {
	// When running as a pod in-cluster, a kubeconfig is not needed. Instead this will make use of the service account injected into the pod.
	// However, allow the use of a local kubeconfig as this can make local development & testing easier.
	kubeconfig := flag.String("kubeconfig", "", "Path to a kubeconfig file")

	// We log to stderr because glog will default to logging to a file.
	// By setting this debugging is easier via `kubectl logs`
	flag.Set("logtostderr", "true")
	flag.Parse()

	// Build the client config - optionally using a provided kubeconfig file.
	config, err := common.GetClientConfig(*kubeconfig)
	if err != nil {
		glog.Fatalf("Failed to load client config: %v", err)
	}

	// Construct the Kubernetes client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Failed to create kubernetes client: %v", err)
	}

	glog.Infof("Starting reboot controller")
	newRebootController(client).controller.Run(wait.NeverStop)
}

type rebootController struct {
	client     kubernetes.Interface
	nodeLister storeToNodeLister
	controller cache.ControllerInterface
}

func newRebootController(client kubernetes.Interface) *rebootController {
	rc := &rebootController{
		client: client,
	}

	store, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(alo api.ListOptions) (runtime.Object, error) {
				var lo v1.ListOptions
				v1.Convert_api_ListOptions_To_v1_ListOptions(&alo, &lo, nil)

				// We do not add any selectors because we want to watch all nodes.
				// This is so we can determine the total count of "unavailable" nodes.
				// However, this could also be implemented using multiple informers (or better, shared-informers)
				return client.Core().Nodes().List(lo)
			},
			WatchFunc: func(alo api.ListOptions) (watch.Interface, error) {
				var lo v1.ListOptions
				v1.Convert_api_ListOptions_To_v1_ListOptions(&alo, &lo, nil)
				return client.Core().Nodes().Watch(lo)
			},
		},
		// The types of objects this informer will return
		&v1.Node{},
		// The resync period of this object. This will force a re-queue of all cached objects at this interval.
		// Every object will trigger the `Updatefunc` even if there have been no actual updates triggered.
		// In some cases you can set this to a very high interval - as you can assume you will see periodic updates in normal operation.
		// The interval is set low here for demo purposes.
		10*time.Second,
		// Callback Functions to trigger on add/update/delete
		cache.ResourceEventHandlerFuncs{
			AddFunc:    rc.handler,
			UpdateFunc: func(old, new interface{}) { rc.handler(new) },
			DeleteFunc: rc.handler,
		},
	)

	rc.controller = controller
	// Convert the cache.Store to a nodeLister to avoid some boilerplate (e.g. convert runtime.Objects to *v1.Nodes)
	// TODO(aaron): use upstream cache.StoreToNodeLister once v3.0.0 client-go available
	rc.nodeLister = storeToNodeLister{store}

	return rc
}

func (c *rebootController) handler(obj interface{}) {
	// TODO(aaron): This would be better handled using a workqueue. This will be added to client-go during v1.6.x release.
	//   As we process objects, add to queue for processing, rather than potentially rebooting whichver node checked in last.
	//   A good example of this pattern is shown in: https://github.com/kubernetes/community/blob/master/contributors/devel/controllers.md
	//   We could also protect against operating against a partial cache by not starting processing until cached synced.

	node := obj.(*v1.Node)
	glog.V(4).Infof("Received update of node: %s", node.Name)
	if node.Annotations == nil {
		return // If node has no annotations, then it doesn't need a reboot
	}

	if _, ok := node.Annotations[common.RebootNeededAnnotation]; !ok {
		return // Node does not need reboot
	}

	// Determine if we should reboot based on maximum number of unavailable nodes
	unavailable, err := c.unavailableNodeCount()
	if err != nil {
		glog.Errorf("Failed to determine number of unavailable nodes: %v", err)
		return
	}

	if unavailable >= maxUnavailable {
		glog.Infof("Too many nodes unvailable (%d/%d). Skipping reboot of %s", unavailable, maxUnavailable, node.Name)
		return
	}

	// We should not modify the cache object directly, so we make a copy first
	nodeCopy, err := common.CopyObjToNode(node)
	if err != nil {
		glog.Errorf("Failed to make copy of node: %v", err)
		return
	}

	glog.Infof("Marking node %s for reboot", node.Name)
	nodeCopy.Annotations[common.RebootAnnotation] = ""
	if _, err := c.client.Core().Nodes().Update(nodeCopy); err != nil {
		glog.Errorf("Failed to set %s annotation: %v", common.RebootAnnotation, err)
	}
}

func (c *rebootController) unavailableNodeCount() (int, error) {
	nodes, err := c.nodeLister.List()
	if err != nil {
		return 0, err
	}
	var unavailable int
	for _, n := range nodes.Items {
		if nodeIsRebooting(&n) {
			unavailable++
			continue
		}
		for _, c := range n.Status.Conditions {
			if c.Type == v1.NodeReady && c.Status == v1.ConditionFalse {
				unavailable++
			}
		}
	}
	return unavailable, nil
}

func nodeIsRebooting(n *v1.Node) bool {
	// Check if node is marked for reeboot-in-progress
	if n.Annotations == nil {
		return false // No annotations - not marked as needing reboot
	}
	if _, ok := n.Annotations[common.RebootInProgressAnnotation]; ok {
		return true
	}
	// Check if node is already marked for immediate reboot
	_, ok := n.Annotations[common.RebootAnnotation]
	return ok
}

// The current client-go StoreToNodeLister expects api.Node - but client returns v1.Node. Add this shim until next release
type storeToNodeLister struct {
	cache.Store
}

func (s *storeToNodeLister) List() (machines v1.NodeList, err error) {
	for _, m := range s.Store.List() {
		machines.Items = append(machines.Items, *(m.(*v1.Node)))
	}
	return machines, nil
}
