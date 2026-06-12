# 📦 README — Dynamic Deployments & ConfigMaps

# 🎯 Objectif

L’objectif de cette évolution est de transformer l’Operator Kubernetes en :

```text
un générateur dynamique de ressources Kubernetes
```

piloté directement depuis la CRD.

L’Operator ne gère désormais plus :
- un seul Deployment
- une seule ConfigMap

mais :
- plusieurs Deployments
- plusieurs ConfigMaps
- dynamiquement
- selon les paramètres du manifest utilisateur

Cela permet :
- de simuler des scénarios de charge
- de stresser le control-plane
- de générer des workloads massifs
- de produire des scénarios Kubernetes réalistes

---

# 🧠 Architecture finale

```text
Custom Resource
        ↓
Reconcile()
        ↓
Lecture paramètres CRD
        ↓
Génération dynamique ressources
        ├── Deployments
        ├── ConfigMaps
        └── Service
        ↓
Convergence Kubernetes
        ↓
Cleanup ressources excédentaires
        ↓
Status global agrégé
```

---

# 🔧 Modifications CRD — controlplanetest_types.go

# 🎯 Objectif

Permettre au manifest CR :
- de piloter le nombre de ressources
- de définir leur taille
- de générer du stress control-plane

---

# 📦 Ajout SchedulerStress

## Code ajouté

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

## Rôle

Permet de :
- générer plusieurs Deployments
- générer plusieurs Pods
- piloter kube-scheduler

---

# 📦 Ajout EtcdStress

## Code ajouté

```go
type EtcdStressSpec struct {
	Enabled bool `json:"enabled,omitempty"`
	ConfigMapCount int32 `json:"configMapCount,omitempty"`
	ConfigMapSizeKB int32 `json:"configMapSizeKB,omitempty"`
	SecretCount int32 `json:"secretCount,omitempty"`
	SecretSizeKB int32 `json:"secretSizeKB,omitempty"`
}
```

## Rôle

Permet :
- de créer massivement des ConfigMaps
- de générer de gros objets etcd
- de stresser :
  - etcd
  - API Server
  - informer cache

---

# 📦 Modification du Status

## Ancien status

```go
DeploymentName string
ConfigMapName string
```

## Nouveau status

```go
DeploymentNames []string
ConfigMapNames []string
```

## Rôle

Permet :
- d’agréger plusieurs ressources
- d’exposer un état global réel
- de suivre tous les objets créés

---

# 🔧 Modifications controller.go

# 🎯 Génération dynamique des Deployments

## Code ajouté

```go
deploymentCount := int32(1)

if controlPlaneTest.Spec.SchedulerStress.Enabled {

	if controlPlaneTest.Spec.SchedulerStress.DeploymentCount > 0 {
		deploymentCount =
			controlPlaneTest.Spec.SchedulerStress.DeploymentCount
	}
}
```

## Rôle

Détermine dynamiquement :
```text
combien de Deployments créer
```

---

# 📦 Boucle de création Deployments

## Code ajouté

```go
for i := int32(0); i < deploymentCount; i++ {
```

## Rôle

Permet :
- création massive
- scaling dynamique
- stress scheduler
- stress API Server

---

# 📦 Nommage dynamique

## Code ajouté

```go
deploymentName := fmt.Sprintf(
	"%s-deployment-%d",
	controlPlaneTest.Name,
	i,
)
```

## Rôle

Génère :
```text
deployment-0
deployment-1
deployment-2
...
```

---

# 📦 Aggregation Status

## Code ajouté

```go
var totalReadyReplicas int32
var totalAvailableReplicas int32
```

Puis :

```go
totalReadyReplicas += existingDeployment.Status.ReadyReplicas
totalAvailableReplicas += existingDeployment.Status.AvailableReplicas
```

## Rôle

Permet :
- un status global
- une vue consolidée du cluster
- des métriques réalistes

---

# 📦 Génération dynamique des ConfigMaps

## Code ajouté

```go
configMapCount := int32(1)

if controlPlaneTest.Spec.EtcdStress.Enabled {

	if controlPlaneTest.Spec.EtcdStress.ConfigMapCount > 0 {
		configMapCount =
			controlPlaneTest.Spec.EtcdStress.ConfigMapCount
	}
}
```

## Rôle

Permet :
- génération massive ConfigMaps
- stress etcd
- stress API Server

---

# 📦 Taille dynamique des ConfigMaps

## Code ajouté

```go
strings.Repeat(
	"A",
	int(controlPlaneTest.Spec.EtcdStress.ConfigMapSizeKB)*1024,
)
```

## Rôle

Permet :
- grosses payloads etcd
- stress stockage Kubernetes
- augmentation mémoire informer cache

---

# 🧹 Cleanup dynamique des ressources

# 🎯 Objectif

Supprimer :
```text
les ressources qui ne sont plus désirées
```

---

# 📦 Cleanup Deployments

## Code ajouté

```go
deploymentList := &appsv1.DeploymentList{}
```

Puis :

```go
r.List(...)
```

et :

```go
r.Delete(...)
```

## Rôle

Permet :
- convergence réelle
- scaling DOWN
- garbage collection applicative

---

# 📦 Cleanup ConfigMaps

## Même logique

```go
configMapList := &corev1.ConfigMapList{}
```

Puis :
- comparaison état réel / désiré
- suppression des ConfigMaps excédentaires

---

# 🎯 Résultat final

L’Operator sait désormais :

✅ Créer dynamiquement des Deployments  
✅ Créer dynamiquement des ConfigMaps  
✅ Agréger le status global  
✅ Supprimer les ressources excédentaires  
✅ Générer du stress control-plane  
✅ Simuler des scénarios Kubernetes massifs

---

# 🚀 Exemple — Stress Scheduler simple

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest

metadata:
  name: scheduler-stress

spec:
  image: nginx:latest

  schedulerStress:
    enabled: true
    deploymentCount: 20
    replicasPerDeployment: 10
```

## Résultat attendu

```text
20 Deployments
200 Pods
forte activité scheduler
forte activité controller-manager
forte activité API Server
```

---

# 🚀 Exemple — Stress etcd massif

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest

metadata:
  name: etcd-stress

spec:
  image: nginx:latest

  etcdStress:
    enabled: true
    configMapCount: 500
    configMapSizeKB: 1024
```

## Résultat attendu

```text
500 ConfigMaps
~500MB données Kubernetes
forte charge etcd
forte charge informer cache
augmentation mémoire API Server
```

---

# 🚀 Exemple — Scaling DOWN massif

## Etat initial

```yaml
deploymentCount: 100
```

Puis :

```yaml
deploymentCount: 5
```

## Résultat attendu

```text
Suppression 95 Deployments
suppression ReplicaSets
suppression Pods
garbage collection Kubernetes massive
tempête d'évènements Kubernetes
```

---

# 🎯 Conclusion

L’Operator n’est plus simplement :
```text
un créateur de ressources Kubernetes
```

mais devient désormais :

```text
un moteur de convergence dynamique
capable de générer des scénarios de stress control-plane
```

avec :
- génération massive
- cleanup dynamique
- scaling UP/DOWN
- stress etcd
- stress scheduler
- stress API Server
- convergence Kubernetes complète