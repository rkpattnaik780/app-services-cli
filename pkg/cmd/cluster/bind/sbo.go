package bind

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bf2fc6cc711aee1a0c2a/cli/internal/localizer"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/cluster"
	"github.com/redhat-developer/service-binding-operator/api/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/builder"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func RunSBOBind(serviceName string, ns string, appName string) error {
	serviceRef := v1alpha1.Service{
		NamespacedRef: v1alpha1.NamespacedRef{
			Ref: v1alpha1.Ref{
				Group:    cluster.AKCResource.Group,
				Version:  cluster.AKCResource.Version,
				Resource: cluster.AKCResource.Resource,
				Name:     serviceName,
			},
		},
	}
	appGVR := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	appRef := v1alpha1.Application{
		Ref: v1alpha1.Ref{
			Group:    appGVR.Group,
			Version:  appGVR.Version,
			Resource: appGVR.Resource,
			Name:     appName,
		},
	}
	boolFalse := false
	now := time.Now()
	sb := &v1alpha1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("binding-%v", now.Unix()),
			Namespace: ns,
		},
		Spec: v1alpha1.ServiceBindingSpec{
			BindAsFiles: &boolFalse,
			Services:    []v1alpha1.Service{serviceRef},
			Application: &appRef,
		},
	}
	sb.SetGroupVersionKind(v1alpha1.GroupVersionKind)

	client, config, err := client()
	if err != nil {
		return err
	}
	restMapper, err := apiutil.NewDynamicRESTMapper(config)
	if err != nil {
		return err
	}
	typeLookup := context.ResourceLookup(restMapper)

	p := builder.DefaultBuilder.WithContextProvider(context.Provider(client, typeLookup)).Build()

	retry, err := p.Process(sb)

	if retry {
		_, err = p.Process(sb)
	}

	if err != nil {
		return err
	}
	return nil
}

func client() (dynamic.Interface, *rest.Config, error) {
	kubeconfig := os.Getenv("KUBECONFIG")

	if kubeconfig == "" {
		home, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	_, err := os.Stat(kubeconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("%v: %w", localizer.MustLocalizeFromID("cluster.kubernetes.error.configNotFoundError"), err)
	}

	kubeClientConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("%v: %w", localizer.MustLocalizeFromID("cluster.kubernetes.error.loadConfigError"), err)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("%v: %w", localizer.MustLocalizeFromID("cluster.kubernetes.error.loadConfigError"), err)
	}

	dynamicClient, err := dynamic.NewForConfig(kubeClientConfig)

	if err != nil {
		return nil, nil, fmt.Errorf("%v: %w", localizer.MustLocalizeFromID("cluster.kubernetes.error.loadConfigError"), err)
	}
	return dynamicClient, kubeClientConfig, nil
}
