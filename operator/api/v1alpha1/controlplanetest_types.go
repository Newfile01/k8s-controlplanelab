/*
Copyright 2026.

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

// ============================================================
// SCHEDULER STRESS
// ============================================================

type SchedulerStressSpec struct {
	Enabled               bool              `json:"enabled,omitempty"`
	NodeCount             int32             `json:"nodeCount,omitempty"`
	DeploymentCount       int32             `json:"deploymentCount,omitempty"`
	ReplicasPerDeployment int32             `json:"replicasPerDeployment,omitempty"`
	NodeSelector          map[string]string `json:"nodeSelector,omitempty"`
	TopologySpread        bool              `json:"topologySpread,omitempty"`
	Affinity              string            `json:"affinity,omitempty"`
	AntiAffinity          string            `json:"antiAffinity,omitempty"`
}

// ============================================================
// API SERVER STRESS
// ============================================================

type APIServerStressSpec struct {
	Enabled               bool  `json:"enabled,omitempty"`
	FrequentStatusUpdates bool  `json:"frequentStatusUpdates,omitempty"`
	AggressiveReconcile   bool  `json:"aggressiveReconcile,omitempty"`
	RecreateResources     bool  `json:"recreateResources,omitempty"`
	QPS                   int32 `json:"qps,omitempty"`
	Burst                 int32 `json:"burst,omitempty"`
}

// ============================================================
// ETCD STRESS
// ============================================================

type EtcdStressSpec struct {
	Enabled         bool  `json:"enabled,omitempty"`
	ConfigMapCount  int32 `json:"configMapCount,omitempty"`
	ConfigMapSizeKB int32 `json:"configMapSizeKB,omitempty"`
	SecretCount     int32 `json:"secretCount,omitempty"`
	SecretSizeKB    int32 `json:"secretSizeKB,omitempty"`
}

// ============================================================
// CONTROLLER MANAGER STRESS
// ============================================================

type ControllerManagerStressSpec struct {
	Enabled                     bool  `json:"enabled,omitempty"`
	DeploymentCount             int32 `json:"deploymentCount,omitempty"`
	ReplicasPerDeployment       int32 `json:"replicasPerDeployment,omitempty"`
	RecreateReplicaSets         bool  `json:"recreateReplicaSets,omitempty"`
	AggressiveGarbageCollection bool  `json:"aggressiveGarbageCollection,omitempty"`
}

// ============================================================
// OPERATOR STRESS
// ============================================================

type OperatorReconcileSpec struct {
	MaxConcurrent    int32 `json:"maxConcurrent,omitempty"`
	QPS              int32 `json:"qps,omitempty"`
	Burst            int32 `json:"burst,omitempty"`
	BaseDelaySeconds int32 `json:"baseDelaySeconds,omitempty"`
	MaxDelaySeconds  int32 `json:"maxDelaySeconds,omitempty"`
}

type OperatorInformerSpec struct {
	WatchPods        bool `json:"watchPods,omitempty"`
	WatchConfigMaps  bool `json:"watchConfigMaps,omitempty"`
	WatchDeployments bool `json:"watchDeployments,omitempty"`
}

type OperatorStressSpec struct {
	Enabled   bool                  `json:"enabled,omitempty"`
	Profile   string                `json:"profile,omitempty"`
	Reconcile OperatorReconcileSpec `json:"reconcile,omitempty"`
	Informer  OperatorInformerSpec  `json:"informer,omitempty"`
}

// ============================================================
// POD LIFECYCLE STORM
// ============================================================

type PodLifecycleStormSpec struct {
	Enabled                 bool  `json:"enabled,omitempty"`
	RestartPodsEverySeconds int32 `json:"restartPodsEverySeconds,omitempty"`
	DeletePodsRandomly      bool  `json:"deletePodsRandomly,omitempty"`
	CrashLoopSimulation     bool  `json:"crashLoopSimulation,omitempty"`
}

// ============================================================
// CONTROL PLANE TEST SPEC
// ============================================================

// ControlPlaneTestSpec defines the desired state of ControlPlaneTest
// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
// Important: Run "make" to regenerate code after modifying this file
// The following markers will use OpenAPI v3 schema to validate the value
// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

type ControlPlaneTestSpec struct {
	// ============================================================
	// LEGACY SIMPLE MODE
	// ============================================================

	// +kubebuilder:validation:MinLength=1
	Image    string `json:"image,omitempty"`
	Replicas int32  `json:"replicas,omitempty"`

	// ============================================================
	// STRESS SCENARIOS
	// ============================================================

	SchedulerStress         SchedulerStressSpec         `json:"schedulerStress,omitempty"`
	APIServerStress         APIServerStressSpec         `json:"apiServerStress,omitempty"`
	EtcdStress              EtcdStressSpec              `json:"etcdStress,omitempty"`
	ControllerManagerStress ControllerManagerStressSpec `json:"controllerManagerStress,omitempty"`
	OperatorStress          OperatorStressSpec          `json:"operatorStress,omitempty"`
	PodLifecycleStorm       PodLifecycleStormSpec       `json:"podLifecycleStorm,omitempty"`
}

// ============================================================
// STATUS
// ============================================================

// ControlPlaneTestStatus defines the observed state of ControlPlaneTest.
// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
// Important: Run "make" to regenerate code after modifying this file

// For Kubernetes API conventions, see:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

// conditions represent the current state of the ControlPlaneTest resource.
// Each condition has a unique type and reflects the status of a specific aspect of the resource.
//
// Standard condition types include:
// - "Available": the resource is fully functional
// - "Progressing": the resource is being created or updated
// - "Degraded": the resource failed to reach or maintain its desired state
//
// The status of each condition is one of True, False, or Unknown.
type ControlPlaneTestStatus struct {
	DeploymentNames    []string           `json:"deploymentNames,omitempty"`
	ServiceName        string             `json:"serviceName,omitempty"`
	ConfigMapNames     []string           `json:"configMapNames,omitempty"`
	ReadyReplicas      int32              `json:"readyReplicas,omitempty"`
	AvailableReplicas  int32              `json:"availableReplicas,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ControlPlaneTest is the Schema for the controlplanetests API
type ControlPlaneTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitzero"`
	// spec defines the desired state of ControlPlaneTest
	// +required
	Spec ControlPlaneTestSpec `json:"spec"`
	// status defines the observed state of ControlPlaneTest
	Status ControlPlaneTestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ControlPlaneTestList contains a list of ControlPlaneTest
type ControlPlaneTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ControlPlaneTest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ControlPlaneTest{}, &ControlPlaneTestList{})
}
