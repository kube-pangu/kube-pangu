package integrationtests

import (
	"fmt"
	"github.com/kube-pangu/kube-pangu/core/kubepartitioner"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"testing"
)

func TestQueryResourcesReturnSuccessfullyWithLocalKubeConfig(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", fmt.Sprintf("%s/%s", homeDir, ".kube/config"))
	if err != nil {
		fmt.Printf("error getting Kubernetes config: %v\n", err)
		os.Exit(1)
	}

	dynamicClient, err := dynamic.NewForConfig(kubeConfig)

	client := kubepartitioner.NewKubeResourceClient(dynamicClient, []schema.GroupVersionKind{
		schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "Pod",
		},
		schema.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		},
	}, kubeConfig)

	resources, err := client.QueryResources()
	shouldBeNil(t, err)
	shouldBeTrue(t, len(resources) > 0)
}

func shouldBeTrue(t *testing.T, cond bool) {
	if !cond {
		t.Error(fmt.Sprintf("\"%v\" should be true", cond))
	}
}

func shouldBeNil(t *testing.T, obj interface{}) {
	if obj != nil {
		t.Error(fmt.Sprintf("\"%v\" should be Nil", obj))
	}
}
