/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"testing"
	"time"

	kcpclienthelper "github.com/kcp-dev/apimachinery/pkg/client"
	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/apis/third_party/conditions/util/conditions"
	"github.com/kcp-dev/logicalcluster/v2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	tutorialkubebuilderiov1alpha1 "github.com/yourrepo/kb-kcp-tutorial/api/v1alpha1" //+kubebuilder:scaffold:imports
)

// The tests in this package expect to be called when:
// - kcp is running
// - a kind cluster is up and running
// - it is hosting a syncer, and the SyncTarget is ready to go
// - the controller-manager from this repo is deployed to kcp
// - that deployment is synced to the kind cluster
// - the deployment is rolled out & ready
//
// We can then check that the controllers defined here are working as expected.

var workspaceName string

func init() {
	rand.Seed(time.Now().Unix())
	flag.StringVar(&workspaceName, "workspace", "", "Workspace in which to run these tests.")
}

func parentWorkspace(t *testing.T) logicalcluster.Name {
	flag.Parse()
	if workspaceName == "" {
		t.Fatal("--workspace cannot be empty")
	}

	return logicalcluster.New(workspaceName)
}

func loadClusterConfig(t *testing.T, clusterName logicalcluster.Name) *rest.Config {
	t.Helper()
	restConfig, err := config.GetConfigWithContext("base")
	if err != nil {
		t.Fatalf("failed to load *rest.Config: %v", err)
	}
	return rest.AddUserAgent(kcpclienthelper.SetCluster(rest.CopyConfig(restConfig), clusterName), t.Name())
}

func loadClient(t *testing.T, clusterName logicalcluster.Name) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add client go to scheme: %v", err)
	}
	if err := tenancyv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add %s to scheme: %v", tenancyv1alpha1.SchemeGroupVersion, err)
	}
	if err := tutorialkubebuilderiov1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add %s to scheme: %v", tutorialkubebuilderiov1alpha1.GroupVersion, err)
	}
	if err := apisv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add %s to scheme: %v", apisv1alpha1.SchemeGroupVersion, err)
	}
	tenancyClient, err := client.New(loadClusterConfig(t, clusterName), client.Options{Scheme: scheme})
	if err != nil {
		t.Fatalf("failed to create a client: %v", err)
	}
	return tenancyClient
}

func createWorkspace(t *testing.T, clusterName logicalcluster.Name) client.Client {
	t.Helper()
	parent, ok := clusterName.Parent()
	if !ok {
		t.Fatalf("cluster %s has no parent", clusterName)
	}
	c := loadClient(t, parent)
	t.Logf("creating workspace %s", clusterName)
	if err := c.Create(context.TODO(), &tenancyv1alpha1.ClusterWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName.Base(),
		},
		Spec: tenancyv1alpha1.ClusterWorkspaceSpec{
			Type: tenancyv1alpha1.ClusterWorkspaceTypeReference{
				Name: "universal",
				Path: "root",
			},
		},
	}); err != nil {
		t.Fatalf("failed to create workspace: %s: %v", clusterName, err)
	}

	t.Logf("waiting for workspace %s to be ready", clusterName)
	var workspace tenancyv1alpha1.ClusterWorkspace
	if err := wait.PollImmediate(100*time.Millisecond, wait.ForeverTestTimeout, func() (done bool, err error) {
		fetchErr := c.Get(context.TODO(), client.ObjectKey{Name: clusterName.Base()}, &workspace)
		if fetchErr != nil {
			t.Logf("failed to get workspace %s: %v", clusterName, err)
			return false, fetchErr
		}
		var reason string
		if actual, expected := workspace.Status.Phase, tenancyv1alpha1.ClusterWorkspacePhaseReady; actual != expected {
			reason = fmt.Sprintf("phase is %s, not %s", actual, expected)
			t.Logf("not done waiting for workspace %s to be ready: %s", clusterName, reason)
		}
		return reason == "", nil
	}); err != nil {
		t.Fatalf("workspace %s never ready: %v", clusterName, err)
	}

	return createAPIBinding(t, clusterName)
}

func createAPIBinding(t *testing.T, workspaceCluster logicalcluster.Name) client.Client {
	c := loadClient(t, workspaceCluster)
	apiName := "test-sdk-test-sdk.tutorial.kubebuilder.io"
	t.Logf("creating APIBinding %s|%s", workspaceCluster, apiName)
	if err := c.Create(context.TODO(), &apisv1alpha1.APIBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: apiName,
		},
		Spec: apisv1alpha1.APIBindingSpec{
			Reference: apisv1alpha1.ExportReference{
				Workspace: &apisv1alpha1.WorkspaceExportReference{
					Path:       parentWorkspace(t).String(),
					ExportName: apiName,
				},
			},
			// TODO(user): PermissionClaims need to be configured for the desired resources
			// Example:
			// PermissionClaims: []apisv1alpha1.AcceptablePermissionClaim{
			//      {
			//              PermissionClaim: apisv1alpha1.PermissionClaim{
			//                      GroupResource: apisv1alpha1.GroupResource{Resource: "configmaps"},
			//              },
			//              State: apisv1alpha1.ClaimAccepted,
			//      },
			// },
		},
	}); err != nil {
		t.Fatalf("could not create APIBinding %s|%s: %v", workspaceCluster, apiName, err)
	}

	t.Logf("waiting for APIBinding %s|%s to be bound", workspaceCluster, apiName)
	var apiBinding apisv1alpha1.APIBinding
	if err := wait.PollImmediate(100*time.Millisecond, wait.ForeverTestTimeout, func() (done bool, err error) {
		fetchErr := c.Get(context.TODO(), client.ObjectKey{Name: apiName}, &apiBinding)
		if fetchErr != nil {
			t.Logf("failed to get APIBinding %s|%s: %v", workspaceCluster, apiName, err)
			return false, fetchErr
		}
		var reason string
		if !conditions.IsTrue(&apiBinding, apisv1alpha1.InitialBindingCompleted) {
			condition := conditions.Get(&apiBinding, apisv1alpha1.InitialBindingCompleted)
			if condition != nil {
				reason = fmt.Sprintf("%s: %s", condition.Reason, condition.Message)
			} else {
				reason = "no condition present"
			}
			t.Logf("not done waiting for APIBinding %s|%s to be bound: %s", workspaceCluster, apiName, reason)
		}
		return conditions.IsTrue(&apiBinding, apisv1alpha1.InitialBindingCompleted), nil
	}); err != nil {
		t.Fatalf("APIBinding %s|%s never bound: %v", workspaceCluster, apiName, err)
	}

	return c
}

const characters = "abcdefghijklmnopqrstuvwxyz"

func randomName() string {
	b := make([]byte, 10)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}
	return string(b)
}

// TestController verifies that the controller behavior works.
func TestController(t *testing.T) {
	t.Parallel()
	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("attempt-%d", i), func(t *testing.T) {
			t.Parallel()
			namespaceName := randomName()
			workspaceCluster := parentWorkspace(t).Join(namespaceName)
			c := createWorkspace(t, workspaceCluster)
			t.Logf("workspace client %!v(MISSING)", c)

			// TODO(user): Create resources and check that the desired reconciliation took place.
			// Example:
			t.Logf("creating namespace %s|%s", workspaceCluster, namespaceName)
			if err := c.Create(context.TODO(), &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: namespaceName}}); err != nil {
				t.Fatalf("failed to create a namespace: %v", err)
			}
			if err := c.Create(context.TODO(), &tutorialkubebuilderiov1alpha1.Widget{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespaceName, Name: fmt.Sprintf("resource-%d", i)},
				Spec:       tutorialkubebuilderiov1alpha1.WidgetSpec{Foo: "scott"},
			}); err != nil {
				t.Fatalf("failed to create Widget: %v", err)
			}
		})
	}
}
