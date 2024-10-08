/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kfpcv1alpha1 "github.com/gregsheremeta/kfpipeline-controller/api/v1alpha1"
)

// KFPipelineReconciler reconciles a KFPipeline object
type KFPipelineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kfpipeline-controller.opendatahub.org,namespace=kfpipeline-controller,resources=kfpipelines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kfpipeline-controller.opendatahub.org,namespace=kfpipeline-controller,resources=kfpipelines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kfpipeline-controller.opendatahub.org,namespace=kfpipeline-controller,resources=kfpipelines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KFPipeline object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *KFPipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	fmt.Printf("Reconciling KFPipeline %s in namespace %s\n", req.Name, req.Namespace)

	logger := log.FromContext(ctx)

	kfPipeline := &kfpcv1alpha1.KFPipeline{}
	err := r.Get(ctx, req.NamespacedName, kfPipeline)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "failed to get KFPipeline")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	fmt.Println("found it. Syncing it.")
	err = SyncPipeline(req.Name, req.Namespace, kfPipeline)
	if err != nil {
		logger.Info(fmt.Sprintf("failed to sync KFPipeline: %v", err))
		logger.Info("TODO set a status on the KFPipeline CR to indicate the error")
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	fmt.Println("successfully synced.")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KFPipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kfpcv1alpha1.KFPipeline{}).
		Complete(r)
}
