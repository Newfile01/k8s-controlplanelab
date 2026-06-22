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

// Paramètres influençant le comportement du Scheduler Kubernetes.
//
// Ces paramètres ne modifient PAS la forme de la charge.
// Ils influencent uniquement le placement des Pods
// sur les nœuds du cluster.
type LabelSelectorSpec struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type SchedulerStressSpec struct {
	// Active le scénario Scheduler Stress.
	Enabled bool `json:"enabled,omitempty"`
	// Nombre de nœuds ciblés par le scénario, utilisés par certaines stratégies de placement.
	NodeCount int32 `json:"nodeCount,omitempty"`
	// Sélection stricte des nœuds via labels Kubernetes.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Répartition homogène des Pods = équilibrage
	TopologySpread bool `json:"topologySpread,omitempty"`
	// Valeurs possibles pour AffinityMode et AntiAffinityMode :
	// "soft" = PreferredDuringSchedulingIgnoredDuringExecution
	// "hard" = RequiredDuringSchedulingIgnoredDuringExecution
	//
	// Utilisé pour forcer le regroupement avec autres pods (par label)
	AffinityMode     string            `json:"affinityMode,omitempty"`
	AffinitySelector LabelSelectorSpec `json:"affinitySelector,omitempty"`
	// PodAntiAffinity, forcera la séparation des pods selon un critère donné
	AntiAffinityMode     string            `json:"antiAffinityMode,omitempty"`
	AntiAffinitySelector LabelSelectorSpec `json:"antiAffinitySelector,omitempty"`
}

// ============================================================
// API SERVER & ETCD STRESS
// ============================================================

type APIServerETCDStressSpec struct {
	Enabled               bool  `json:"enabled,omitempty"`
	FrequentStatusUpdates bool  `json:"frequentStatusUpdates,omitempty"`
	AggressiveReconcile   bool  `json:"aggressiveReconcile,omitempty"`
	RecreateResources     bool  `json:"recreateResources,omitempty"`
	QPS                   int32 `json:"qps,omitempty"`
	Burst                 int32 `json:"burst,omitempty"`
}

// ============================================================
// CONTROLLER MANAGER STRESS
// ============================================================

type ControllerManagerStressSpec struct {
	Enabled                     bool `json:"enabled,omitempty"`
	RecreateReplicaSets         bool `json:"recreateReplicaSets,omitempty"`
	AggressiveGarbageCollection bool `json:"aggressiveGarbageCollection,omitempty"`
}

// ============================================================
// OPERATOR STRESS
// ============================================================

type OperatorReconcileSpec struct {
	// Nbre réconciliations simultanées possibles
	MaxConcurrent int32 `json:"maxConcurrent,omitempty"`
	// QPS autorisées pour le pod operator_controller
	QPS int32 `json:"qps,omitempty"`
	// Explosion/Pic de requêtes recevable
	Burst int32 `json:"burst,omitempty"`
	// Délai (avant réconciliation ?)
	BaseDelaySeconds int32 `json:"baseDelaySeconds,omitempty"`
	// Délai maximum avant nouvelle réconciliation
	MaxDelaySeconds int32 `json:"maxDelaySeconds,omitempty"`
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
// RESSOURCES PAR PODS
// ============================================================

type ResourceSpec struct {
	// Ressources demandées
	CPURequest    string `json:"cpuRequest,omitempty"`
	MemoryRequest string `json:"memoryRequest,omitempty"`
	// Max autorisé à Kubernetes
	CPULimit    string `json:"cpuLimit,omitempty"`
	MemoryLimit string `json:"memoryLimit,omitempty"`
}

// ============================================================
// CONTROL PLANE TEST SPEC
// ============================================================

// ControlPlaneTestSpec defines the desired state of ControlPlaneTest
// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
// Important: Run "make" to regenerate code after modifying this file
// The following markers will use OpenAPI v3 schema to validate the value
// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html
type CustomLabelSpec struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type ControlPlaneTestSpec struct {
	// ============================================================
	// WORKLOAD DEFINITION
	// ============================================================
	ContainerName   string          `json:"containerName"`
	ResourcesPerPod ResourceSpec    `json:"resourcesPerPod,omitempty"`
	CustomLabel     CustomLabelSpec `json:"customLabel,omitempty"`
	// Image utilisée pour les Pods générés.
	Image string `json:"image"`
	// Nombre de Deployments à générer.
	DeploymentCount int32 `json:"deploymentCount"`
	// Nombre de Pods par Deployment.
	ReplicasPerDeployment int32 `json:"replicasPerDeployment,omitempty"`

	// ConfigMaps & Secrets
	ConfigMapCount  int32 `json:"configMapCount,omitempty"`
	ConfigMapSizeKB int32 `json:"configMapSizeKB,omitempty"`
	SecretCount     int32 `json:"secretCount,omitempty"`
	SecretSizeKB    int32 `json:"secretSizeKB,omitempty"`

	// ============================================================
	// STRESS SCENARIOS
	// ============================================================

	SchedulerStress         SchedulerStressSpec         `json:"schedulerStress,omitempty"`
	APIServerETCDStress     APIServerETCDStressSpec     `json:"apiServerEtcdStress,omitempty"`
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
