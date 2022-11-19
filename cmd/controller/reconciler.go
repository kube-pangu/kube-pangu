package main

import (
	"context"
	"fmt"
	"github.com/kube-pangu/kube-pangu/core/kubepartitioner"
	"github.com/kube-pangu/kube-pangu/core/partitioner"
	"github.com/spaolacci/murmur3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type Reconciler struct {
	config *rest.Config
}

func (r *Reconciler) Run() {
	dynamicClient, err := dynamic.NewForConfig(r.config)
	if err != nil {
		fmt.Println("Could not initialize dynamic client: ", err)
		return
	}

	configsGVR, err := getGVRFromGVK(&schema.GroupVersionKind{
		Group:   "kubepartitions.io",
		Version: "v1alpha1",
		Kind:    "Configuration",
	}, r.config)

	if err != nil {
		fmt.Println("Error getting configs GVR", err)
		return
	}

	nodesGVR, err := getGVRFromGVK(&schema.GroupVersionKind{
		Group:   "kubepartitions.io",
		Version: "v1alpha1",
		Kind:    "Node",
	}, r.config)

	if err != nil {
		fmt.Println("Error getting nodes GVR", err)
		return
	}

	unstructuredConfigs, err := dynamicClient.Resource(*configsGVR).List(context.Background(), v1.ListOptions{})

	if err != nil {
		fmt.Println("Error querying config objects", err)
		return
	}

	unstructuredNodes, err := dynamicClient.Resource(*nodesGVR).List(context.Background(), v1.ListOptions{})

	if err != nil {
		fmt.Println("Error querying node objects", err)
		return
	}

	unstructuredConfigsByName := map[string]unstructured.Unstructured{}
	for _, u := range unstructuredConfigs.Items {
		unstructuredConfigsByName[u.GetName()] = u
	}

	unstructuredNodesByConfigName := map[string][]unstructured.Unstructured{}
	for _, u := range unstructuredNodes.Items {
		configName := u.Object["spec"].(map[string]interface{})["configurationName"].(string)
		if _, ok := unstructuredConfigsByName[configName]; !ok {
			fmt.Println("config named ", configName, "not found. Skipping the node for reconciliation")
		}

		unstructuredNodesByConfigName[configName] = append(unstructuredNodesByConfigName[configName], u)
	}

	for _, config := range unstructuredConfigsByName {
		gvks := []schema.GroupVersionKind{}
		gvkObjects := config.Object["spec"].(map[string]interface{})["targetGroupVersionKinds"].([]interface{})
		for _, gvkObject := range gvkObjects {
			gvkMap := gvkObject.(map[string]interface{})
			gvks = append(gvks, schema.GroupVersionKind{
				Group:   gvkMap["group"].(string),
				Version: gvkMap["version"].(string),
				Kind:    gvkMap["kind"].(string),
			})
		}

		resourceClient := kubepartitioner.NewKubeResourceClient(dynamicClient, gvks, r.config)
		partitioner := partitioner.NewConsistentHashPartitioner(config.GetName(), resourceClient, func(key string) uint64 {
			return murmur3.Sum64([]byte(key))
		})

		for _, unstructuredNode := range unstructuredNodesByConfigName[config.GetName()] {
			partitioner.AddPartition(partitionNode{unstructuredNode: unstructuredNode})
		}

		fmt.Println("Doing partition")
		err = partitioner.DoPartition()
		if err != nil {
			fmt.Println("Reconciliation for configuration", config.GetName(), "failed. Error:", err)
		}
	}

}

func getGVRFromGVK(gvk *schema.GroupVersionKind, config *rest.Config) (*schema.GroupVersionResource, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	return &mapping.Resource, nil
}

type partitionNode struct {
	unstructuredNode unstructured.Unstructured
}

func (n partitionNode) GetId() string {
	return n.unstructuredNode.GetName()
}

func (n partitionNode) GetSeeds() []uint64 {
	interfacedSeeds := n.unstructuredNode.Object["spec"].(map[string]interface{})["seeds"].([]interface{})
	seeds := []uint64{}

	for _, interfacedSeed := range interfacedSeeds {
		float64Seed := interfacedSeed.(float64)
		var maxVal uint64
		maxVal = 0xFFFFFFFFFFFFFFFF
		seeds = append(seeds, uint64(float64(maxVal)*float64Seed))
	}

	return seeds
}
