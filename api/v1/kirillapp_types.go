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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KirillAppSpec defines the desired state of KirillApp
type KirillAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

            Replicas *int32 `json:"replicas"`
	    Selector *metav1.LabelSelector `json:"selector"`
	    Template PodTemplateSpec `json:"template"`
    

	// Foo is an example field of KirillApp. Edit kirillapp_types.go to remove/update
}



type PodTemplateSpec struct {
	// Metadata of the pods
	Metadata metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the pods
	Spec PodSpec `json:"spec"`
}

type PodSpec struct {
	// List of containers
	Containers []Container `json:"containers"`
}

type Container struct {
	// Name of the container
	Name string `json:"name"`

	// Docker image to use for the container
	Image string `json:"image"`

	// List of ports to expose
	Ports []ContainerPort `json:"ports,omitempty"`

	// Resource requirements for the container
	Resources ResourceRequirements `json:"resources,omitempty"`
}

type ContainerPort struct {
	// Name of the port
	Name string `json:"name,omitempty"`

	// Port number
	ContainerPort int32 `json:"containerPort"`
}
type ResourceRequirements struct {
	// Limits describes the maximum amount of compute resources allowed
	Limits map[string]string `json:"limits,omitempty"`

	// Requests describes the minimum amount of compute resources required
	Requests map[string]string `json:"requests,omitempty"`
}



// KirillAppStatus defines the observed state of KirillApp
type KirillAppStatus struct {

           AvailableReplicas int32 `json:"availableReplicas"`

	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// KirillApp is the Schema for the kirillapps API
type KirillApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KirillAppSpec   `json:"spec,omitempty"`
	Status KirillAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KirillAppList contains a list of KirillApp
type KirillAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KirillApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KirillApp{}, &KirillAppList{})
}
