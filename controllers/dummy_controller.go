/*
Copyright 2023.

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

package controllers

import (
	"context"
	"fmt"
	dummyapi "github.com/piyush1146115/dummy-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kmc "kmodules.xyz/client-go/client"
	core_util "kmodules.xyz/client-go/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// DummyReconciler reconciles a Dummy object
type DummyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=interview.com,resources=dummies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=interview.com,resources=dummies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=interview.com,resources=dummies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Dummy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *DummyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	dum := &dummyapi.Dummy{}
	if err := r.Client.Get(ctx, req.NamespacedName, dum); err != nil {
		logger.Error(err, "failed to get Dummy")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if dum.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dum.Name,
			Namespace: dum.Namespace,
		},
	}

	ownerRef := metav1.NewControllerRef(dum, dummyapi.GroupVersion.WithKind(dum.Kind))
	if err := r.createOrPatchPod(ctx, pod.ObjectMeta, ownerRef); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create/patch Pod: %w", err)
	}

	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(pod), pod); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get Pod: %w", err)
	}

	dum.Status.PodStatus = string(pod.Status.Phase)
	dum.Status.SpecEcho = dum.Spec.Message

	r.updateStatus(ctx, dum)
	return ctrl.Result{}, nil
}

func (r *DummyReconciler) createOrPatchPod(ctx context.Context, meta metav1.ObjectMeta, owner *metav1.OwnerReference) error {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name,
			Namespace: meta.Namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{
					Name:  "nginx",
					Image: "nginx",
					Command: []string{
						"bin/sh",
						"sleep 360000",
					},
				},
			},
		},
	}

	_, _, err := kmc.CreateOrPatch(
		ctx,
		r.Client,
		pod.DeepCopy(),
		func(obj client.Object, createOp bool) client.Object {
			in := obj.(*v1.Pod)
			in.ObjectMeta = pod.ObjectMeta
			in.Spec = pod.Spec

			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

			return in
		},
	)

	return err
}

func (r *DummyReconciler) updateStatus(ctx context.Context, dum *dummyapi.Dummy) error {
	_, _, err := kmc.PatchStatus(
		ctx,
		r.Client,
		dum.DeepCopy(),
		func(obj client.Object) client.Object {
			in := obj.(*dummyapi.Dummy)
			in.Status = dum.Status
			return in
		},
	)

	return client.IgnoreNotFound(err)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DummyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dummyapi.Dummy{}).
		Owns(&v1.Pod{}).
		Complete(r)
}
