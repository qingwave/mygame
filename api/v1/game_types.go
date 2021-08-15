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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GameSpec defines the desired state of Game
type GameSpec struct {
	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	//+kubebuilder:default:=1
	//+kubebuilder:validation:Minimum:=1
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// Docker image name
	// +optional
	Image string `json:"image,omitempty"`

	// Ingress Host name
	Host string `json:"host,omitempty"`
}

const (
	Init     = "Init"
	Running  = "Running"
	NotReady = "NotReady"
	Failed   = "Failed"
)

// GameStatus defines the observed state of Game
type GameStatus struct {
	// Phase is the phase of guestbook
	Phase string `json:"phase,omitempty"`

	// replicas is the number of Pods created by the StatefulSet controller.
	Replicas int32 `json:"replicas"`

	// readyReplicas is the number of Pods created by the StatefulSet controller that have a Ready Condition.
	ReadyReplicas int32 `json:"readyReplicas"`

	// LabelSelector is label selectors for query over pods that should match the replica count used by HPA.
	LabelSelector string `json:"labelSelector,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.labelSelector
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="The phase of game."
//+kubebuilder:printcolumn:name="Host",type="string",JSONPath=".spec.host",description="The host address."
//+kubebuilder:printcolumn:name="DESIRED",type="integer",JSONPath=".spec.replicas",description="The desired number of pods."
//+kubebuilder:printcolumn:name="CURRENT",type="integer",JSONPath=".status.replicas",description="The number of currently all pods."
//+kubebuilder:printcolumn:name="READY",type="integer",JSONPath=".status.readyReplicas",description="The number of pods ready."
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."

// Game is the Schema for the games API
type Game struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GameSpec   `json:"spec,omitempty"`
	Status GameStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GameList contains a list of Game
type GameList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Game `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Game{}, &GameList{})
}
