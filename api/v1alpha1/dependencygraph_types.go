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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type InvocationEdge struct {
	// FunctionName is the name of the invoked function, used as a pod/service selector. It should match the function name in another node in the graph.
	FunctionName string `json:"functionName"`
	// FunctionNamespace is the namespace of the invoked function
	FunctionNamespace string `json:"functionNamespace"`
	// Id of the invocation. Edges with the same id are invoked concurrently, different ids imply the invocations happen sequentially.
	EdgeId int32 `json:"edgeId"`
	// Multiplier describes how many invocations to this function are performed by the caller function.
	EdgeMultiplier int32 `json:"edgeMultiplier"`
}

type FunctionNode struct {
	// FunctionName represents what function this node is assigned to and it is used as a selector for the pods running said function.
	FunctionName string `json:"functionName"`
	// FunctionNamespace is the namespace of the invoked function
	FunctionNamespace string `json:"functionNamespace"`
	// Invocations is the list of out-edges from the node to invoked functions.
	Invocations []InvocationEdge `json:"invocations"`
	// Nominal Response Time is the response time recorded in the profiling phase
	NominalResponseTime resource.Quantity `json:"nominalResponseTime"`
}

// DependencyGraphSpec defines the desired state of DependencyGraph.
type DependencyGraphSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Nodes represents the collection of nodes in the graph
	Nodes []FunctionNode `json:"nodes"`
}

type NodeStatus struct {
	// FunctionName is the name of the function
	FunctionName string `json:"functionName"`
	// FunctionNamespace is the namespace of the function
	FunctionNamespace string `json:"functionNamespace"`
	// ExternalResponseTime is the aggregated metric describing the average response time of the function's dependencies
	ExternalResponseTime int64 `json:"externalResponseTime"`
}

// DependencyGraphStatus defines the observed state of DependencyGraph.
type DependencyGraphStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Nodes []NodeStatus `json:"nodes"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DependencyGraph is the Schema for the dependencygraphs API.
type DependencyGraph struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DependencyGraphSpec   `json:"spec,omitempty"`
	Status DependencyGraphStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DependencyGraphList contains a list of DependencyGraph.
type DependencyGraphList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DependencyGraph `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DependencyGraph{}, &DependencyGraphList{})
}
