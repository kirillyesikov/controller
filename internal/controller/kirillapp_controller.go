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
        appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

          "sigs.k8s.io/controller-runtime/pkg/client"
        "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
        "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	appv1 "github.com/kirillyesikov/controller/api/v1"
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
	_ = log.FromContext(ctx)
            
	kirillApp := &appv1.KirillApp{}
	err := r.Get(ctx, req.NamespacedName, kirillApp)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
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
					 Image: kirillApp.Spec.Image,		 

			           },
			   },
                         },
		       },
	       },     

              }




	      if err := controllerutil.SetControllerReference(kirillApp, deployment, r.Scheme); err != nil {
               return ctrl.Result{}, err

	   }



	           found := &appsv1.Deployment{}
		   err = r.Get(ctx, client.ObjectKeyFromObject(deployment), found)
		    if err != nil {
			  if client.IgnoreNotFound(err) != nil {
				  err = r.Create(ctx, deployment)
				  if err != nil {
				     return ctrl.Result{}, err
			          }
		         }
		 } else {
			 if *found.Spec.Replicas != *deployment.Spec.Replicas || found.Spec.Template.Spec.Containers[0].Image != deployment.Spec.Template.Spec.Containers[0].Image {
				 found.Spec.Replicas = deployment.Spec.Replicas
				 found.Spec.Template.Spec.Containers[0].Image = deployment.Spec.Template.Spec.Containers[0].Image
				 err = r.Update(ctx, found)
				 if err != nil {
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
					Port:       80, // External port
					TargetPort: intstr.FromInt(3000), // Port in the container
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeLoadBalancer, // Change to NodePort/LoadBalancer if needed
		},
	}

	// Set controller reference for Service
	if err := controllerutil.SetControllerReference(kirillApp, service, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}


foundService := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKeyFromObject(service), foundService)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			err = r.Create(ctx, service)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// Update Service if needed
		if !serviceSpecEqual(service.Spec, foundService.Spec) {
			foundService.Spec = service.Spec
			err = r.Update(ctx, foundService)
			if err != nil {
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
