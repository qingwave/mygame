/*
Copyright 2021.

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
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	myappv1 "qingwave.github.io/mygame/api/v1"
)

// GameReconciler reconciles a Game object
type GameReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=myapp.qingwave.github.io,resources=games,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=myapp.qingwave.github.io,resources=games/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=myapp.qingwave.github.io,resources=games/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Game object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *GameReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	defer utilruntime.HandleCrash()

	logger := log.FromContext(ctx)
	logger.Info("revice reconcile event", "name", req.String())

	// get game object
	game := &myappv1.Game{}
	if err := r.Get(ctx, req.NamespacedName, game); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if game.DeletionTimestamp != nil {
		logger.Info("game in deleting", "name", req.String())
		return ctrl.Result{}, nil
	}

	if err := r.syncGame(ctx, game); err != nil {
		logger.Error(err, "failed to sync game", "name", req.String())
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

const (
	gameLabelName = "qingwave.github.io/game"
	port          = 80
)

func (r *GameReconciler) syncGame(ctx context.Context, obj *myappv1.Game) error {
	logger := log.FromContext(ctx)

	game := obj.DeepCopy()
	name := types.NamespacedName{
		Namespace: game.Namespace,
		Name:      game.Name,
	}

	owner := []metav1.OwnerReference{
		{
			APIVersion:         game.APIVersion,
			Kind:               game.Kind,
			Name:               game.Name,
			Controller:         pointer.BoolPtr(true),
			BlockOwnerDeletion: pointer.BoolPtr(true),
			UID:                game.UID,
		},
	}

	labels := map[string]string{
		gameLabelName: game.Name,
	}

	meta := metav1.ObjectMeta{
		Name:            game.Name,
		Namespace:       game.Namespace,
		Labels:          labels,
		OwnerReferences: owner,
	}

	deploy := &appsv1.Deployment{}
	if err := r.Get(ctx, name, deploy); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		deploy = &appsv1.Deployment{
			ObjectMeta: meta,
			Spec:       getDeploymentSpec(game, labels),
		}
		if err := r.Create(ctx, deploy); err != nil {
			return err
		}
		logger.Info("create deployment success", "name", name.String())
	} else {
		want := getDeploymentSpec(game, labels)
		get := getSpecFromDeployment(deploy)
		if !reflect.DeepEqual(want, get) {
			new := deploy.DeepCopy()
			new.Spec = want
			if err := r.Update(ctx, new); err != nil {
				return err
			}
			logger.Info("update deployment success", "name", name.String())
		}
	}

	svc := &corev1.Service{}
	if err := r.Get(ctx, name, svc); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		svc = &corev1.Service{
			ObjectMeta: meta,
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Port:       int32(port),
						TargetPort: intstr.FromInt(port),
						Protocol:   corev1.ProtocolTCP,
					},
				},
				Selector: labels,
			},
		}
		if err := r.Create(ctx, svc); err != nil {
			return err
		}
		logger.Info("create service success", "name", name.String())
	}

	ing := &networkingv1.Ingress{}
	if err := r.Get(ctx, name, ing); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		ing = &networkingv1.Ingress{
			ObjectMeta: meta,
			Spec:       getIngressSpec(game, labels),
		}
		if err := r.Create(ctx, ing); err != nil {
			return err
		}
		logger.Info("create ingress success", "name", name.String())
	}

	newStatus := myappv1.GameStatus{
		Replicas:      *game.Spec.Replicas,
		ReadyReplicas: deploy.Status.ReadyReplicas,
	}

	if newStatus.Replicas == newStatus.ReadyReplicas {
		newStatus.Phase = myappv1.Running
	} else {
		newStatus.Phase = myappv1.NotReady
	}

	if !reflect.DeepEqual(game.Status, newStatus) {
		game.Status = newStatus
		logger.Info("update game status", "name", name.String())
		return r.Client.Status().Update(ctx, game)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GameReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 3,
		}).
		For(&myappv1.Game{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)

	return nil
}

func getDeploymentSpec(game *myappv1.Game, labels map[string]string) appsv1.DeploymentSpec {
	return appsv1.DeploymentSpec{
		Replicas: game.Spec.Replicas,
		Selector: metav1.SetAsLabelSelector(labels),
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "main",
						Image: game.Spec.Image,
					},
				},
			},
		},
	}
}

func getSpecFromDeployment(deploy *appsv1.Deployment) appsv1.DeploymentSpec {
	container := deploy.Spec.Template.Spec.Containers[0]
	return appsv1.DeploymentSpec{
		Replicas: deploy.Spec.Replicas,
		Selector: deploy.Spec.Selector,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: deploy.Spec.Template.Labels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  container.Name,
						Image: container.Image,
					},
				},
			},
		},
	}
}

func getIngressSpec(game *myappv1.Game, labels map[string]string) networkingv1.IngressSpec {
	pathType := networkingv1.PathTypePrefix
	return networkingv1.IngressSpec{
		Rules: []networkingv1.IngressRule{
			{
				Host: game.Spec.Host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								PathType: &pathType,
								Path:     "/",
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: game.Name,
										Port: networkingv1.ServiceBackendPort{
											Number: int32(port),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
