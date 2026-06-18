# Régulation des scénarios de stress Kubernetes

## 1. Solutions immédiatement exploitables (sans modifier l'opérateur)

---

## 1.1 NodeSelector (déjà implémenté)

### CRD

Déjà présent dans :

```go
type SchedulerStressSpec struct {
    ...
    NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}
```

### Exemple

```yaml
schedulerStress:
  enabled: true

  nodeSelector:
    kubernetes.io/hostname: worker-1
```

---

## 1.2 Affinity / AntiAffinity (déjà implémenté)

### Pod Affinity

Concentre les Pods sur les mêmes nœuds.

```yaml
schedulerStress:
  affinity: hard
```

ou

```yaml
schedulerStress:
  affinity: soft
```

---

### Pod AntiAffinity

Force la dispersion.

```yaml
schedulerStress:
  antiAffinity: hard
```

ou

```yaml
schedulerStress:
  antiAffinity: soft
```

---

## 1.3 TopologySpread (déjà implémenté)

Répartit les Pods sur les nœuds.

```yaml
schedulerStress:
  topologySpread: true
```

---

## 1.4 LimitRange (fortement recommandé)

Permet de définir automatiquement les ressources CPU/RAM de tous les Pods du namespace même si l'opérateur ne les définit pas.

```yaml
apiVersion: v1
kind: LimitRange

metadata:
  name: operator-default-limits
  namespace: operator-system

spec:
  limits:

  - type: Container

    defaultRequest:
      cpu: 100m
      memory: 128Mi

    default:
      cpu: 500m
      memory: 512Mi
```

Application :

```bash
kubectl apply -f limitrange.yaml
```

---

## 1.5 ResourceQuota (fortement recommandé)

Empêche le namespace de dépasser certaines limites.

```yaml
apiVersion: v1
kind: ResourceQuota

metadata:
  name: operator-quota
  namespace: operator-system

spec:
  hard:
    pods: "200"

    requests.cpu: "8"
    requests.memory: 16Gi

    limits.cpu: "16"
    limits.memory: 32Gi
```

Application :

```bash
kubectl apply -f quota.yaml
```

---

# 2. Isolation complète du Control Plane

Objectif :

* kube-apiserver
* etcd
* scheduler
* controller-manager

restent isolés des Pods de stress.

---

## Étape 1 : Vérifier les Pods système

```bash
kubectl get pods -A -o wide
```

Vérifier que tous les Pods kube-system sont déjà présents.

---

## Étape 2 : Ajouter un label

```bash
kubectl label node control-plane \
dedicated=controlplane
```

Vérification :

```bash
kubectl get node --show-labels
```

---

## Étape 3 : Ajouter un taint

⚠️ À faire seulement après vérification des Pods système.

```bash
kubectl taint node control-plane \
dedicated=controlplane:NoSchedule
```

---

## Étape 4 : Vérifier les tolérations existantes

```bash
kubectl get pod -A -o yaml \
| grep -A5 tolerations:
```

Sur la plupart des clusters :

* kube-apiserver
* etcd
* scheduler
* controller-manager

tolèrent déjà les taints du control-plane.

---

## Étape 5 : Si nécessaire patch manuel

Exemple :

```bash
kubectl edit pod kube-apiserver-control-plane \
-n kube-system
```

ou plus durable :

```bash
kubectl edit staticpod \
(ou manifest sous /etc/kubernetes/manifests)
```

Selon le type de cluster.

---

## Étape 6 : Utiliser NodeSelector dans la CR

Exemple :

```yaml
schedulerStress:

  enabled: true

  nodeSelector:
    dedicated: workers
```

Les Pods de stress ne pourront jamais aller sur le control-plane.

---

# 3. Évolution n°1 (très simple)

# Ajouter CPU/RAM dans la CRD

---

## Fichier

```text
api/v1alpha1/controlplanetest_types.go
```

---

## Ajouter avant

```go
type SchedulerStressSpec struct {
```

Ajouter :

```go
type ResourcesSpec struct {

    CPURequest string `json:"cpuRequest,omitempty"`
    CPULimit string `json:"cpuLimit,omitempty"`

    MemoryRequest string `json:"memoryRequest,omitempty"`
    MemoryLimit string `json:"memoryLimit,omitempty"`
}
```

---

## Puis dans SchedulerStressSpec

Avant :

```go
type SchedulerStressSpec struct {

    Enabled bool `json:"enabled,omitempty"`

    NodeCount int32 `json:"nodeCount,omitempty"`

    DeploymentCount int32 `json:"deploymentCount,omitempty"`

    ReplicasPerDeployment int32 `json:"replicasPerDeployment,omitempty"`

    NodeSelector map[string]string `json:"nodeSelector,omitempty"`

    TopologySpread bool `json:"topologySpread,omitempty"`

    Affinity string `json:"affinity,omitempty"`

    AntiAffinity string `json:"antiAffinity,omitempty"`
}
```

Après :

```go
type SchedulerStressSpec struct {

    Enabled bool `json:"enabled,omitempty"`

    NodeCount int32 `json:"nodeCount,omitempty"`

    DeploymentCount int32 `json:"deploymentCount,omitempty"`

    ReplicasPerDeployment int32 `json:"replicasPerDeployment,omitempty"`

    NodeSelector map[string]string `json:"nodeSelector,omitempty"`

    Resources ResourcesSpec `json:"resources,omitempty"`

    TopologySpread bool `json:"topologySpread,omitempty"`

    Affinity string `json:"affinity,omitempty"`

    AntiAffinity string `json:"antiAffinity,omitempty"`
}
```

---

## Dans le controller

Fichier :

```text
internal/controller/controlplanetest_controller.go
```

Ajouter dans les imports :

```go
"k8s.io/apimachinery/pkg/api/resource"
```

---

Rechercher :

```go
Containers: []corev1.Container{
{
    Name: "nginx",
    Image: controlPlaneTest.Spec.Image,
},
},
```

Remplacer par :

```go
Containers: []corev1.Container{
{
    Name: "nginx",
    Image: controlPlaneTest.Spec.Image,

    Resources: corev1.ResourceRequirements{

        Requests: corev1.ResourceList{
            corev1.ResourceCPU:
                resource.MustParse(
                    controlPlaneTest.Spec.SchedulerStress.Resources.CPURequest,
                ),

            corev1.ResourceMemory:
                resource.MustParse(
                    controlPlaneTest.Spec.SchedulerStress.Resources.MemoryRequest,
                ),
        },

        Limits: corev1.ResourceList{
            corev1.ResourceCPU:
                resource.MustParse(
                    controlPlaneTest.Spec.SchedulerStress.Resources.CPULimit,
                ),

            corev1.ResourceMemory:
                resource.MustParse(
                    controlPlaneTest.Spec.SchedulerStress.Resources.MemoryLimit,
                ),
        },
    },
},
},
```

---

## Utilisation

```yaml
schedulerStress:

  resources:

    cpuRequest: "100m"
    cpuLimit: "500m"

    memoryRequest: "128Mi"
    memoryLimit: "512Mi"
```

---

# 4. Évolution n°2 (simple)

# Ajouter les Tolerations

---

## CRD

Ajouter avant SchedulerStressSpec :

```go
type TolerationSpec struct {

    Key string `json:"key,omitempty"`

    Operator string `json:"operator,omitempty"`

    Value string `json:"value,omitempty"`

    Effect string `json:"effect,omitempty"`
}
```

---

Dans SchedulerStressSpec :

```go
Tolerations []TolerationSpec `json:"tolerations,omitempty"`
```

---

## Controller

Rechercher :

```go
Spec: corev1.PodSpec{
```

Ajouter juste après :

```go
Tolerations: []corev1.Toleration{
{
    Key: "dedicated",

    Operator: corev1.TolerationOpEqual,

    Value: "controlplane",

    Effect: corev1.TaintEffectNoSchedule,
},
},
```

Version statique simple.

---

## Utilisation

```yaml
schedulerStress:

  tolerations:

  - key: dedicated
    operator: Equal
    value: controlplane
    effect: NoSchedule
```

---

# 5. Évolution n°3 (plus avancée)

# Limitation globale de l'opérateur

Déjà partiellement implémentée.

Fichier :

```text
internal/controller/controlplanetest_controller.go
```

Fin de fichier :

```go
WithOptions(controller.Options{

    MaxConcurrentReconciles: 2,

    RateLimiter:
```

Tu peux diminuer :

```go
MaxConcurrentReconciles: 1
```

et :

```go
rate.Limit(5),
50,
```

au lieu de :

```go
rate.Limit(10),
100,
```

pour réduire fortement la pression exercée sur l'API Server.

---

# Recommandation finale

Ordre conseillé :

1. LimitRange
2. ResourceQuota
3. NodeSelector
4. Isolation ControlPlane via taints
5. Resources CPU/RAM dans la CRD
6. Tolerations dans la CRD
7. Réduction des limites internes du controller-runtime

Cela permet déjà de lancer des scénarios très agressifs sans bloquer complètement le cluster.
