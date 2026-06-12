# 📊 METRICS & OBSERVABILITE — Kubernetes Operator

# 🎯 Objectif

Cette partie du lab permet :

* d’exposer des métriques Prometheus
* d’observer les scénarios de stress
* de mesurer le comportement du control-plane
* de construire des dashboards Grafana

L’objectif est de transformer l’opérateur en :

```text
plateforme observable
de simulation Kubernetes
```

---

# 🧠 Architecture générale

```text
Operator
    ↓
Prometheus metrics
    ↓
/metrics
    ↓
Prometheus
    ↓
Grafana
    ↓
Dashboards control-plane
```

---

# 🔧 Métriques initiales

## Reconcile

```text
controlplanetest_reconciliation_total
controlplanetest_erreurs_reconciliation_total
controlplanetest_reconciliation_duree_secondes
```

---

## Ressources

```text
controlplanetest_pods_geres
controlplanetest_replicas_disponibles
```

---

# 🔧 Métriques Scheduler Stress

```text
controlplanetest_deployments_generes
controlplanetest_pods_desires
controlplanetest_pods_pending
```

---

# 🔧 Métriques API Server Stress

```text
controlplanetest_status_updates_total
controlplanetest_requeues_forcees_total
controlplanetest_resources_recreated_total
```

---

# 🔧 Métriques PodLifecycleStorm

```text
controlplanetest_pods_supprimes_total
```

---

# 🧠 Pourquoi utiliser des labels Prometheus ?

Chaque métrique est maintenant :

```text
labelisée par CR
```

Exemple :

```text
controlplanetest_pods_pending{
	cr="scheduler-hard",
	namespace="default",
	profile="hard"
}
```

Cela permet :

* filtrage Grafana
* dashboards dynamiques
* corrélation scénarios ↔ métriques

---

# 🔧 Passage vers GaugeVec / CounterVec

## Avant

```go
prometheus.NewGauge(
```

---

## Après

```go
prometheus.NewGaugeVec(
```

---

# 🧠 Labels utilisés

## Identité scénario

```text
cr
namespace
profile
```

---

## Configuration scénario

```text
affinity
topologyspread
podstorm
aggressive_reconcile
```

---

# 🔧 Configuration Snapshot Metrics

## 📍 Objectif

Exposer :

```text
la configuration active
de chaque scénario
```

---

## 📍 Exemple

```text
controlplanetest_configuration_info{
	cr="scheduler-hard",
	profile="hard",
	topologyspread="enabled",
	aggressive_reconcile="enabled",
	podstorm_enabled="enabled"
} 1
```

---

# 🧠 Pourquoi cette métrique est importante ?

Elle permet :

* d’identifier les scénarios
* de construire des dashboards intelligents
* de corréler comportement ↔ configuration

---

# 📊 Ce que Grafana pourra afficher

## Scheduler

* Pods Pending
* saturation scheduling
* répartition Pods

---

## API Server

* writes/sec
* PATCH storms
* reconcile storms

---

## etcd

* churn objets
* writes massifs

---

## Operator

* reconcile duration
* requeues
* erreurs
* workqueue pressure

---

# 🚀 Exemple de dashboard

## Variables Grafana

```promql
label_values(controlplanetest_configuration_info, cr)
```

---

## Panneaux possibles

```text
Pods Pending
Reconcile Duration
API Churn
Resources Recreated
Pod Churn
Scheduler Pressure
```

---

# 🧠 Ce que cette partie met en lumière

Dans Kubernetes :

```text
l'observabilité
est aussi importante
que le workload lui-même
```

Les métriques permettent :

* d’identifier les limites cluster
* de corréler les évènements
* de comprendre le comportement du control-plane

Cette partie transforme l’opérateur en :

```text
véritable plateforme d’analyse
du control-plane Kubernetes
```
