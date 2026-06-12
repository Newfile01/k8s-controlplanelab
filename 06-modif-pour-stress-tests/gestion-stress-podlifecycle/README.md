# 🌪️ POD LIFECYCLE STORM — Kubernetes Operator

# 🎯 Objectif

Cette partie du lab permet de générer volontairement :

* des suppressions Pods
* des recréations ReplicaSets
* des rescheduling continus
* des tempêtes d’évènements Kubernetes

L’objectif est de stresser :

* kube-scheduler
* kube-controller-manager
* kubelet
* kube-apiserver
* etcd
* controller-runtime

---

# 🧠 Principe général

Le PodLifecycleStorm provoque :

```text
DELETE Pod
    ↓
ReplicaSet détecte manque replicas
    ↓
CREATE nouveau Pod
    ↓
Scheduler
    ↓
Node binding
    ↓
Kubelet startup
    ↓
Events Kubernetes
    ↓
Reconcile()
```

Cette boucle génère :

```text
du churn Kubernetes massif
```

---

# 🔧 Paramètres ajoutés dans la CRD

## 📍 controlplanetest_types.go

```go
type PodLifecycleStormSpec struct {
	Enabled bool `json:"enabled,omitempty"`
	RestartPodsEverySeconds int32 `json:"restartPodsEverySeconds,omitempty"`
	DeletePodsRandomly bool `json:"deletePodsRandomly,omitempty"`
	CrashLoopSimulation bool `json:"crashLoopSimulation,omitempty"`
}
```

---

# 🧠 Rôle des paramètres

## enabled

Active :

```text
le Pod Lifecycle Storm
```

---

## restartPodsEverySeconds

Force :

```text
une tempête périodique
de suppressions/recréations Pods
```

---

## deletePodsRandomly

Supprime :

```text
un Pod aléatoire
```

à chaque reconcile.

---

## crashLoopSimulation

Préparation future :

```text
CrashLoopBackOff storms
```

---

# 🔧 Ajout dans reconcile()

## 📍 Récupération Pods

```go
podList := &corev1.PodList{}

err = r.List(
	ctx,
	podList,
	client.InNamespace(req.Namespace),
	client.MatchingLabels{
		"app": controlPlaneTest.Name,
	},
)
```

---

## 📍 Suppression Pod aléatoire

```go
randomIndex := rand.Intn(len(podList.Items))

randomPod := podList.Items[randomIndex]

err = r.Delete(
	ctx,
	&randomPod,
)
```

---

## 📍 Requeue périodique

```go
requeueResult = &ctrl.Result{
	RequeueAfter: restartDelay,
}
```

---

# 🔄 Fonctionnement réel

```text
DELETE Pod
    ↓
Event Kubernetes
    ↓
ReplicaSet reconcile
    ↓
CREATE Pod
    ↓
Scheduler
    ↓
Kubelet
    ↓
Container startup
    ↓
Readiness probes
    ↓
Events Kubernetes
```

---

# 📊 Métriques ajoutées

```text
controlplanetest_pods_supprimes_total
controlplanetest_pods_pending
```

---

# 📈 Résultat observé

Le cluster génère :

* Pod churn
* scheduler churn
* kubelet churn
* watch storms
* reconcile storms

---

# 🚀 Exemple de scénario

```yaml
podLifecycleStorm:
  enabled: true
  deletePodsRandomly: true
  restartPodsEverySeconds: 1
```

---

# 🎯 Résultat attendu

## kube-scheduler

* scheduling continu

## kubelet

* démarrage/arrêt constant containers

## controller-manager

* ReplicaSet reconcile permanent

## kube-apiserver

* DELETE/CREATE massifs

## etcd

* writes permanents

---

# 🧠 Ce que cette partie met en lumière

Dans Kubernetes :

```text
les Pods sont éphémères
```

Le control-plane est conçu pour :

* recréer
* rescheduler
* reconverger

en permanence.

Cette partie permet de visualiser :

```text
la capacité de résilience
du control-plane Kubernetes
face aux tempêtes de Pods
```
