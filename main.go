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

package main

import (
	"context"
	"flag"
	"fmt"
	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/discovery"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/kcp"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	tutorialkubebuilderiov1alpha1 "github.com/yourrepo/kb-kcp-tutorial/api/v1alpha1"
	"github.com/yourrepo/kb-kcp-tutorial/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(tutorialkubebuilderiov1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var configFile string
	var configFile2 string
	var apiExportName string
	flag.StringVar(&apiExportName, "api-export-name", "", "The name of the APIExport.")
	flag.StringVar(&configFile, "config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")
	flag.StringVar(&configFile2, "config2", "",
		"The mirror controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	setupLog = setupLog.WithValues("api-export-name", apiExportName)
	mgr, err := manager.New(ctrl.GetConfigOrDie(), manager.Options{Scheme: scheme})
	if err != nil {
		panic(err)
	}

	var err2 error
	options2 := ctrl.Options{Scheme: scheme}
	if configFile2 != "" {
		options2, err2 = options2.AndFrom(ctrl.ConfigFile().AtPath(configFile2))
		if err2 != nil {
			setupLog.Error(err, "unable to load the config file 2")
			os.Exit(1)
		}
	}
	options2.MetricsBindAddress = "0"

	var clientCfg clientcmd.ClientConfig

	var kubeConfig, _ = ioutil.ReadFile(configFile2)
	clientCfg, err = clientcmd.NewClientConfigFromBytes(kubeConfig)
	if err != nil {
		log.Fatal(err)
	}

	var restCfg *rest.Config

	setupLog.Info("here1")
	restCfg, err = clientCfg.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	setupLog.Info("here2")
	mirrorCluster, err2 := cluster.New(restCfg)

	if err2 != nil {
		panic(err2)
	}

	setupLog.Info("here3")
	if err := mgr.Add(mirrorCluster); err != nil {
		panic(err)
	}

	setupLog.Info("here4")
	if err := NewMirrorWidgetReconciler(mgr, mirrorCluster); err != nil {
		panic(err)
	}

	//setupLog.Info("here4")
	//if err = (&controllers.WidgetReconciler{
	//	Client: mgr.GetClient(),
	//	Scheme: mgr.GetScheme(),
	//}).SetupWithManager(mgr); err != nil {
	//	setupLog.Error(err, "unable to create controller", "controller", "Widget")
	//	os.Exit(1)
	//}

	setupLog.Info("here5")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		panic(err)
	}
}

func NewMirrorWidgetReconciler(mgr manager.Manager, mirrorCluster cluster.Cluster) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Watch Secrets in the reference cluster
		For(&tutorialkubebuilderiov1alpha1.Widget{}).
		// Watch Secrets in the mirror cluster
		Watches(
			source.NewKindWithCache(&tutorialkubebuilderiov1alpha1.Widget{}, mirrorCluster.GetCache()),
			&handler.EnqueueRequestForObject{},
		).
		Complete(&controllers.WidgetReconciler{
			Client: mirrorCluster.GetClient(),
			Scheme: mgr.GetScheme(),
		})
}

func main2() {
	var configFile string
	var configFile2 string
	var apiExportName string
	flag.StringVar(&apiExportName, "api-export-name", "", "The name of the APIExport.")
	flag.StringVar(&configFile, "config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")
	flag.StringVar(&configFile2, "config2", "",
		"The mirror controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	setupLog = setupLog.WithValues("api-export-name", apiExportName)

	ctx := ctrl.SetupSignalHandler()

	restConfig := ctrl.GetConfigOrDie()

	var mgr ctrl.Manager
	var mgr2 ctrl.Manager
	var err error
	var err2 error

	if kcpAPIsGroupPresent(restConfig) {
		setupLog.Info("Looking up virtual workspace URL")
		cfg, err := restConfigForAPIExport(ctx, restConfig, apiExportName)
		if err != nil {
			setupLog.Error(err, "error looking up virtual workspace URL")
		}

		setupLog.Info("Using virtual workspace URL", "url", cfg.Host)

		options := ctrl.Options{Scheme: scheme}
		options.LeaderElectionConfig = restConfig

		if configFile != "" {
			options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile))
			if err != nil {
				setupLog.Error(err, "unable to load the config file")
				os.Exit(1)
			}
		}

		mgr, err = kcp.NewClusterAwareManager(cfg, options)
		if err != nil {
			setupLog.Error(err, "unable to start cluster aware manager")
			os.Exit(1)
		}
	} else {
		setupLog.Info("The apis.kcp.dev group is not present - creating standard manager")

		options := ctrl.Options{Scheme: scheme}
		options2 := ctrl.Options{Scheme: scheme}

		if configFile != "" {
			options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile))
			if err != nil {
				setupLog.Error(err, "unable to load the config file")
				os.Exit(1)
			}
		}

		mgr, err = ctrl.NewManager(ctrl.GetConfigOrDie(), options)
		if err != nil {
			setupLog.Error(err, "unable to start manager")
			os.Exit(1)
		}

		if configFile2 != "" {
			options2, err2 = options.AndFrom(ctrl.ConfigFile().AtPath(configFile2))
			if err2 != nil {
				setupLog.Error(err, "unable to load the config file 2")
				os.Exit(1)
			}
		}
		options2.MetricsBindAddress = "0"

		mgr2, err2 = ctrl.NewManager(ctrl.GetConfigOrDie(), options2)
		if err2 != nil {
			setupLog.Error(err2, "unable to start manager 2")
			os.Exit(1)
		}

	}

	if err = (&controllers.WidgetReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Widget")
		os.Exit(1)
	}

	if err2 = (&controllers.WidgetReconciler{
		Client: mgr2.GetClient(),
		Scheme: mgr2.GetScheme(),
	}).SetupWithManager(mgr2); err2 != nil {
		setupLog.Error(err2, "unable to create controller", "controller", "Widget")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
	setupLog.Info("starting manager 2")
	if err2 := mgr2.Start(ctx); err2 != nil {
		setupLog.Error(err2, "problem running manager 2")
		os.Exit(1)
	}
}

// +kubebuilder:rbac:groups="apis.kcp.dev",resources=apiexports,verbs=get;list;watch

// restConfigForAPIExport returns a *rest.Config properly configured to communicate with the endpoint for the
// APIExport's virtual workspace.
func restConfigForAPIExport(ctx context.Context, cfg *rest.Config, apiExportName string) (*rest.Config, error) {
	scheme := runtime.NewScheme()
	if err := apisv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("error adding apis.kcp.dev/v1alpha1 to scheme: %w", err)
	}

	apiExportClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("error creating APIExport client: %w", err)
	}

	var apiExport apisv1alpha1.APIExport

	if apiExportName != "" {
		if err := apiExportClient.Get(ctx, types.NamespacedName{Name: apiExportName}, &apiExport); err != nil {
			return nil, fmt.Errorf("error getting APIExport %q: %w", apiExportName, err)
		}
	} else {
		setupLog.Info("api-export-name is empty - listing")
		exports := &apisv1alpha1.APIExportList{}
		if err := apiExportClient.List(ctx, exports); err != nil {
			return nil, fmt.Errorf("error listing APIExports: %w", err)
		}
		if len(exports.Items) == 0 {
			return nil, fmt.Errorf("no APIExport found")
		}
		if len(exports.Items) > 1 {
			return nil, fmt.Errorf("more than one APIExport found")
		}
		apiExport = exports.Items[0]
	}

	if len(apiExport.Status.VirtualWorkspaces) < 1 {
		return nil, fmt.Errorf("APIExport %q status.virtualWorkspaces is empty", apiExportName)
	}

	cfg = rest.CopyConfig(cfg)
	// TODO(ncdc): sharding support
	cfg.Host = apiExport.Status.VirtualWorkspaces[0].URL

	return cfg, nil
}

func kcpAPIsGroupPresent(restConfig *rest.Config) bool {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		setupLog.Error(err, "failed to create discovery client")
		os.Exit(1)
	}
	apiGroupList, err := discoveryClient.ServerGroups()
	if err != nil {
		setupLog.Error(err, "failed to get server groups")
		os.Exit(1)
	}

	for _, group := range apiGroupList.Groups {
		if group.Name == apisv1alpha1.SchemeGroupVersion.Group {
			for _, version := range group.Versions {
				if version.Version == apisv1alpha1.SchemeGroupVersion.Version {
					return true
				}
			}
		}
	}
	return false
}
