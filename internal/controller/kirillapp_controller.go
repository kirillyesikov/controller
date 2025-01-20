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
	appv1 "github.com/kirillyesikov/controller/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// KirillAppReconciler reconciles a KirillApp object
type KirillAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.kubernetesoperator.atwebpages.com,resources=kirillapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.kubernetesoperator.atwebpages.com,resources=kirillapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.kubernetesoperator.atwebpages.com,resources=kirillapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KirillApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *KirillAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
        log.Info("Reconciling KirillApp", "Name", req.NamespacedName)

	kirillApp := &appv1.KirillApp{}
	if err := r.Get(ctx, req.NamespacedName, kirillApp); err != nil {
		if errors.IsNotFound(err) {
			log.Info("KirillApp resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get KirillApp")
		return ctrl.Result{}, err
	}

	// Define Deployment
	replicas := int32(1) // Default replicas to 1 if unspecified
	if kirillApp.Spec.Replicas != nil {
		replicas = *kirillApp.Spec.Replicas
	}

	
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kirillApp.Name + "-deployment",
			Namespace: kirillApp.Namespace,
		},

		Spec: appsv1.DeploymentSpec{
			Replicas: kirillApp.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": kirillApp.Name},
			},

			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": kirillApp.Name},
				},

				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  kirillApp.Name,
							Image: "kyesikov/radio:latest",
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(kirillApp, deployment, r.Scheme); err != nil {
		log.Error(err, "Failed to set controller reference for Deployment")
		return ctrl.Result{}, err

	}

	foundDeployment := &appsv1.Deployment{}
      if err := r.Get(ctx, client.ObjectKeyFromObject(deployment), foundDeployment); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating Deployment", "DeploymentName", deployment.Name)
			if err := r.Create(ctx, deployment); err != nil {
				log.Error(err, "Failed to create Deployment")
				return ctrl.Result{}, err
			}
		} else {
			log.Error(err, "Failed to get Deployment")
			return ctrl.Result{}, err
		}
	} else {
		if *foundDeployment.Spec.Replicas != *deployment.Spec.Replicas ||
			foundDeployment.Spec.Template.Spec.Containers[0].Image != deployment.Spec.Template.Spec.Containers[0].Image {
			log.Info("Updating Deployment", "DeploymentName", foundDeployment.Name)
			foundDeployment.Spec.Replicas = deployment.Spec.Replicas
			foundDeployment.Spec.Template.Spec.Containers[0].Image = deployment.Spec.Template.Spec.Containers[0].Image
			if err := r.Update(ctx, foundDeployment); err != nil {
				log.Error(err, "Failed to update Deployment")
				return ctrl.Result{}, err
			}
		}
	}


	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kirillApp.Name + "-service",
			Namespace: kirillApp.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": kirillApp.Name},
			Ports: []corev1.ServicePort{
				{
					Port:       80,                   // External port
					TargetPort: intstr.FromInt(3000), // Port in the container
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeLoadBalancer, // Change to NodePort/LoadBalancer if needed
		},
	}

	// Set controller reference for Service
	if err := controllerutil.SetControllerReference(kirillApp, service, r.Scheme); err != nil {
		log.Error(err, "Failed to set controller reference for Service")
		return ctrl.Result{}, err
	}

	foundService := &corev1.Service{}
    if err := r.Get(ctx, client.ObjectKeyFromObject(service), foundService); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating Service", "ServiceName", service.Name)
			if err := r.Create(ctx, service); err != nil {
				log.Error(err, "Failed to create Service")
				return ctrl.Result{}, err
			}
		} else {
			log.Error(err, "Failed to get Service")
			return ctrl.Result{}, err
		}
	} else {
		if !serviceSpecEqual(service.Spec, foundService.Spec) {
			log.Info("Updating Service", "ServiceName", foundService.Name)
			foundService.Spec = service.Spec
			if err := r.Update(ctx, foundService); err != nil {
				log.Error(err, "Failed to update Service")
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func serviceSpecEqual(existing, desired corev1.ServiceSpec) bool {
	if len(existing.Ports) != len(desired.Ports) {
		return false
	}
	for i := range existing.Ports {
		if existing.Ports[i] != desired.Ports[i] {
			return false
		}
	}
	return existing.Selector != nil && desired.Selector != nil &&
		reflect.DeepEqual(existing.Selector, desired.Selector) &&
		existing.Type == desired.Type
}

// TODO(user): your logic here

// SetupWithManager sets up the controller with the Manager.
func (r *KirillAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.KirillApp{}).
		Complete(r)
}
