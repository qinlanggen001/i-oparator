/*
Copyright 2025.

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
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	pkgerror "github.com/pkg/errors"
	v1 "github.com/qinlanggen001/i-operator.git/api/v1"
	"github.com/songzhibin97/gkit/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// 事件记录器
	Recorder record.EventRecorder
}

const (
	AppFinalizer = "genlang.cn/application"
)

// +kubebuilder:rbac:groups=core.crd.genlang.cn,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.crd.genlang.cn,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.crd.genlang.cn,resources=applications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//_ = logf.FromContext(ctx)

	// TODO(user): your logic here
	logger := logf.FromContext(ctx)
	log := logger.WithValues("application", req.NamespacedName)

	log.Info("start reconcile")
	// query app
	var app v1.Application
	err := r.Get(ctx, req.NamespacedName, &app)
	if err != nil {
		log.Error(err, "unable to feath application")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if app.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&app, AppFinalizer) {
			controllerutil.AddFinalizer(&app, AppFinalizer)
			if err := r.Update(ctx, &app); err != nil {
				log.Error(err, "unable to add finalizer to application")
				return ctrl.Result{}, nil
			}
			r.Recorder.Event(&app, corev1.EventTypeNormal, "AddFinalizer", fmt.Sprintf("add finalizer %s", AppFinalizer))
		}
	} else {
		if controllerutil.ContainsFinalizer(&app, AppFinalizer) {
			if err = r.DeleteExternalResources(&app); err != nil {
				log.Error(err, "unable to cleanup application")
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(&app, AppFinalizer)
			if err = r.Update(ctx, &app); err != nil {
				return ctrl.Result{}, err
			}
			r.Recorder.Event(&app, corev1.EventTypeNormal, "RemoveFinalizer", fmt.Sprintf("remove finalizer %s", AppFinalizer))
		}
		return ctrl.Result{}, nil
	}

	log.Info("run reconcile logic")
	if err = r.syncApp(ctx, app); err != nil {
		log.Error(err, "unable to sync application")
		return ctrl.Result{}, nil
	}

	// sync status
	var deploy appsv1.Deployment
	objectKey := client.ObjectKey{Namespace: app.Namespace, Name: app.Name}
	err = r.Get(ctx, objectKey, &deploy)
	if err != nil {
		log.Error(err, "unable to fetch deployment", "deploy", objectKey.String())
	}
	copyApp := app.DeepCopy()
	copyApp.Status.Ready = deploy.Status.ReadyReplicas >= 1
	if !reflect.DeepEqual(app, copyApp) {
		if err = r.Client.Status().Update(ctx, copyApp); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

func (r *ApplicationReconciler) syncApp(ctx context.Context, app v1.Application) error {
	if app.Spec.Enabled {
		return r.syncAppEnabled(ctx, app)
	} else {
		return r.syncAppDisabled(ctx, app)
	}
}

func (r *ApplicationReconciler) syncAppEnabled(ctx context.Context, app v1.Application) error {
	var deploy appsv1.Deployment
	objectKey := client.ObjectKey{Namespace: app.Namespace, Name: app.Name}
	err := r.Get(ctx, objectKey, &deploy)
	if err != nil {
		if errors.IsNotFound(err) {
			deploy = r.generateDeployment(app)
			if err := r.Create(ctx, &deploy); err != nil {
				return pkgerror.WithMessagef(err, "create deployment [%s] faild!", app.Name)
			}
		}
	}
	if !equal(app, deploy) {
		deploy.Spec.Template.Spec.Containers[0].Image = app.Spec.Image
		if err := r.Update(ctx, &app); err != nil {
			return pkgerror.WithMessagef(err, "update deployment [%s] faild", app.Name)
		}
	}
	return nil
}

func equal(app v1.Application, deploy appsv1.Deployment) bool {
	return app.Spec.Image == deploy.Spec.Template.Spec.Containers[0].Image
}
func (r *ApplicationReconciler) generateDeployment(app v1.Application) appsv1.Deployment {
	deploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			Labels: map[string]string{
				"app": app.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": app.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": app.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  app.Name,
							Image: app.Spec.Image,
						},
					},
				},
			},
		},
	}
	_ = controllerutil.SetControllerReference(&app, &deploy, r.Scheme)
	return deploy
}
func (r *ApplicationReconciler) syncAppDisabled(ctx context.Context, app v1.Application) error {
	var deploy appsv1.Deployment
	objectKey := client.ObjectKey{Name: app.Name, Namespace: app.Namespace}
	err := r.Get(ctx, objectKey, &deploy)
	if err != nil {
		return pkgerror.WithMessage(err, "unable fetch deploy")
	}
	if err = r.Delete(ctx, &deploy); err != nil {
		return pkgerror.WithMessage(err, "delete deploy faild")
	}
	return nil
}

func (r *ApplicationReconciler) DeleteExternalResources(app *v1.Application) error {
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Application{}).
		Named("application").
		Complete(r)
}
