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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KFPipelineSpec defines the Kubeflow Pipeline.
type KFPipelineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Description string `json:"description,omitempty"`

	// The Kubeflow Pipelines PipelineSpec yaml (also sometimes referred to as Pipeline IR).
	// +kubebuilder:validation:Required
	PipelineSpec string `json:"pipelineSpec"`
}

// KFPipelineStatus defines the observed state of KFPipeline
type KFPipelineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// KFPipeline is the Schema for the kfpipelines API
type KFPipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KFPipelineSpec   `json:"spec,omitempty"`
	Status KFPipelineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KFPipelineList contains a list of KFPipeline
type KFPipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KFPipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KFPipeline{}, &KFPipelineList{})
}
