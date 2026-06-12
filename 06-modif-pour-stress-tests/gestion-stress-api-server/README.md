# 🔥 API SERVER STRESS — Kubernetes Operator

# 🎯 Objectif

Cette partie du lab permet de générer volontairement de la charge sur :

* kube-apiserver
* etcd
* informer cache
* controller-runtime
* workqueues reconcile

L’objectif est d’observer :

* les limites du control-plane
* le comportement des reconciliations
* les tempêtes d’évènements Kubernetes
* l’impact des writes API massifs

---

# 🧠 Principe général

Le stress est généré via :

* mises à jour fréquentes du Status
* requeues agressifs
* suppressions/recréations ressources
* multiplication des appels API Kubernetes

Architecture générale :

```text
Operator
    ↓
GET / LIST / PATCH / DELETE / CREATE
    ↓
kube-apiserver
    ↓
etcd
    ↓
watch events
    ↓
informers
    ↓
reconcile storms
```

---

# 🔧 Paramètres ajoutés dans la CRD

## 📍 controlplanetest_types.go

```go
type APIServerStressSpec struct {
	Enabled bool `json:"enabled,omitempty"`
	FrequentStatusUpdates bool `json:"frequentStatusUpdates,omitempty"`
	AggressiveReconcile bool `json:"aggressiveReconcile,omitempty"`
	RecreateResources bool `json:"recreateResources,omitempty"`
	QPS int32 `json:"qps,omitempty"`
	Burst int32 `json:"burst,omitempty"`
}
```

---

# 🧠 Rôle des paramètres

## frequentStatusUpdates

Force :

```text
PATCH /status
```

à chaque reconcile.

Impact :

* writes etcd
* propagation informer
* watch events massifs

---

## aggressiveReconcile

Force :

```text
RequeueAfter
```

fréquent.

Impact :

* explosions reconciliations
* workqueue pressure
* GET/LIST massifs

---

## recreateResources

Supprime puis recrée :

* Deployments
* ReplicaSets
* Pods

Impact :

* DELETE/CREATE storms
* scheduler churn
* ReplicaSet churn
* garbage collection

---

## QPS / Burst

Préparation au pilotage futur :

* client-go rate limiting
* saturation API Server contrôlée

---

# 🔧 Ajout dans reconcile()

## 📍 Frequent Status Updates

```go
if controlPlaneTest.Spec.APIServerStress.Enabled &&
	controlPlaneTest.Spec.APIServerStress.FrequentStatusUpdates {

	controlPlaneTest.Status.ObservedGeneration =
		controlPlaneTest.Generation

	controlPlaneTest.Status.ReadyReplicas =
		newReadyReplicas

	controlPlaneTest.Status.AvailableReplicas =
		newAvailableReplicas

	err = r.Status().Update(
		ctx,
		controlPlaneTest,
	)
}
```

---

## 📍 Aggressive Reconcile

```go
if controlPlaneTest.Spec.APIServerStress.Enabled &&
	controlPlaneTest.Spec.APIServerStress.AggressiveReconcile {

	requeueResult = &ctrl.Result{
		RequeueAfter: 1 * time.Second,
	}
}
```

---

## 📍 Recreate Resources

```go
if controlPlaneTest.Spec.APIServerStress.Enabled &&
	controlPlaneTest.Spec.APIServerStress.RecreateResources {

	for _, deployment := range deploymentList.Items {

		err = r.Delete(
			ctx,
			&deployment,
		)
	}
}
```

---

# 🔄 Fonctionnement réel

## Frequent Status Updates

```text
Reconcile()
    ↓
PATCH /status
    ↓
API Server
    ↓
etcd write
    ↓
watch events
    ↓
informers
```

---

## Aggressive Reconcile

```text
Reconcile()
    ↓
RequeueAfter
    ↓
workqueue
    ↓
nouveau reconcile
```

---

## Recreate Resources

```text
DELETE Deployment
    ↓
DELETE ReplicaSet
    ↓
DELETE Pods
    ↓
Events Kubernetes
    ↓
Reconcile()
    ↓
CREATE Deployment
```

---

# 📊 Métriques ajoutées

## API churn

```text
controlplanetest_status_updates_total
controlplanetest_requeues_forcees_total
controlplanetest_resources_recreated_total
```

---

# 📈 Résultat observé

Le cluster commence à générer :

* reconcile storms
* watch storms
* API saturation
* churn Kubernetes
* ReplicaSet churn
* scheduler churn

---

# 🚀 Exemple de scénario

```yaml
apiServerStress:
  enabled: true
  frequentStatusUpdates: true
  aggressiveReconcile: true
  recreateResources: true
```

---

# 🎯 Résultat attendu

## kube-apiserver

* augmentation requêtes/sec
* saturation progressive

## etcd

* writes massifs
* DELETE/CREATE massifs

## controller-runtime

* reconcile storms
* queue pressure

## scheduler

* scheduling continu

---

# 🧠 Ce que cette partie met en lumière

Le control-plane Kubernetes est :

```text
massivement piloté par les évènements
```

Un fort churn API provoque :

* propagation massive d’évènements
* resynchronisations
* réactions en chaîne

Cette partie permet de visualiser :

```text
la fragilité potentielle
du control-plane Kubernetes
sous forte pression API
```
