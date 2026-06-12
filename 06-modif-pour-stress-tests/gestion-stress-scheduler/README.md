# 📦 README — Kubernetes Scheduler Stress

# 🎯 Objectif

Cette évolution transforme l’Operator en :

```text
générateur de contraintes scheduler Kubernetes
```

Le but est de permettre :
- stress kube-scheduler
- simulation production HA
- contraintes de placement
- déséquilibres cluster
- Pending Pods
- scheduling pressure

directement depuis :
```yaml
schedulerStress:
```

---

# 🧠 Architecture finale

```text
Custom Resource
        ↓
Operator
        ↓
Deployment
        ↓
PodSpec avancé
        ├── NodeSelector
        ├── Affinity
        ├── AntiAffinity
        └── TopologySpreadConstraints
        ↓
kube-scheduler
        ↓
Décisions de placement
```

---

# 🔧 Modifications CRD — controlplanetest_types.go

# 📦 Ajout SchedulerStressSpec

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

---

# 🧠 Rôle des paramètres

# 📦 deploymentCount

## Rôle

Nombre de Deployments générés.

## Effets

Augmente :
- ReplicaSets
- Pods
- évènements scheduler
- activité controller-manager

---

# 📦 replicasPerDeployment

## Rôle

Nombre de Pods par Deployment.

## Effets

Augmente :
- scheduling pressure
- Pending Pods
- calculs scheduler

---

# 📦 nodeSelector

## Rôle

Force les Pods :
```text
à être placés uniquement sur certains nodes
```

## Exemple

```yaml
nodeSelector:
  stress: high
```

## Effets

Le scheduler :
- filtre les nodes compatibles
- réduit les placements possibles
- peut provoquer des Pending Pods

---

# 📦 topologySpread

## Rôle

Force le scheduler :
```text
à répartir les Pods entre plusieurs nodes
```

## Effets

Augmente :
- calculs équilibrage
- coût scheduling
- décisions HA

---

# 📦 affinity

## Rôle

Force les Pods :
```text
à se rapprocher d'autres Pods
```

## Modes

### soft

```text
préférence
```

### hard

```text
obligation
```

## Effets

Le scheduler :
- analyse les Pods existants
- compare les nodes
- augmente fortement la complexité de placement

---

# 📦 antiAffinity

## Rôle

Force les Pods :
```text
à éviter certains autres Pods
```

## Modes

### soft

```text
préférence
```

### hard

```text
obligation stricte
```

## Effets

Très coûteux pour kube-scheduler :
- scans inter-Pods
- comparaisons massives
- Pending Pods fréquents

➡️ excellent pour stress scheduler.

---

# 🔧 Modifications controller.go

# 📦 Injection NodeSelector

## Code ajouté

```go
NodeSelector:
	controlPlaneTest.Spec.SchedulerStress.NodeSelector,
```

## Rôle

Injecte dynamiquement :
```yaml
nodeSelector:
```

dans le PodSpec.

---

# 📦 Injection TopologySpreadConstraints

## Code ajouté

```go
TopologySpreadConstraints:
[]corev1.TopologySpreadConstraint{
```

## Rôle

Force :
- répartition cluster
- équilibrage scheduler
- anti-concentration

---

# 📦 Injection Affinity

## Code ajouté

```go
Affinity: affinity,
```

## Variable dynamique générée

```go
var affinity *corev1.Affinity
```

## Rôle

Construit dynamiquement :
- PodAffinity
- PodAntiAffinity

selon :
```yaml
affinity:
antiAffinity:
```

---

# 📦 Affinity soft

## Code ajouté

```go
PreferredDuringSchedulingIgnoredDuringExecution
```

## Rôle

Le scheduler :
```text
essaie
mais n'est pas obligé
```

---

# 📦 Affinity hard

## Code ajouté

```go
RequiredDuringSchedulingIgnoredDuringExecution
```

## Rôle

Le scheduler :
```text
doit respecter la contrainte
```

sinon :
```text
Pod Pending
```

---

# 🎯 Résultat final

L’Operator sait désormais :

✅ Piloter kube-scheduler  
✅ Forcer placement Pods  
✅ Répartir les Pods cluster-wide  
✅ Générer Pending Pods  
✅ Simuler contraintes production  
✅ Générer scheduling pressure  
✅ Générer contraintes HA Kubernetes

---

# 🚀 Exemple — NodeSelector

## Label node

```bash
kubectl label node worker1 stress=high
```

## CR

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest

metadata:
  name: node-selector-test

spec:
  image: nginx:latest

  schedulerStress:
    enabled: true
    deploymentCount: 3
    replicasPerDeployment: 5

    nodeSelector:
      stress: high
```

## Résultat attendu

```text
Tous les Pods placés sur worker1
```

---

# 🚀 Exemple — TopologySpread

```yaml
schedulerStress:
  enabled: true
  deploymentCount: 4
  replicasPerDeployment: 10
  topologySpread: true
```

## Résultat attendu

```text
Pods répartis sur plusieurs nodes
```

➡️ augmentation :
- équilibrage scheduler
- décisions placement

---

# 🚀 Exemple — AntiAffinity hard

```yaml
schedulerStress:
  enabled: true
  deploymentCount: 10
  replicasPerDeployment: 20
  antiAffinity: hard
```

## Résultat attendu

```text
Scheduler refuse certains placements
Pods Pending
forte pression scheduler
```

---

# 🚀 Exemple — Stress scheduler massif

```yaml
schedulerStress:
  enabled: true

  deploymentCount: 50
  replicasPerDeployment: 20

  topologySpread: true
  antiAffinity: hard
```

## Résultat attendu

```text
1000 Pods
contraintes scheduler massives
Pending Pods
forte charge kube-scheduler
activité API Server importante
```

---

# 🎯 Conclusion

Le scheduler Kubernetes n’est plus simplement :
```text
un composant passif du cluster
```

mais devient désormais :

```text
une cible de stress pilotée dynamiquement
par l’Operator Kubernetes
```

L’Operator peut maintenant :
- influencer les décisions scheduler
- générer des contraintes avancées
- produire des scénarios HA réalistes
- simuler des tempêtes de scheduling Kubernetes
- provoquer Pending Pods massifs
- stresser kube-scheduler comme en production