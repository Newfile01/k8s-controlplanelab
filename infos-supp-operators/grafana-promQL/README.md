# Construction d'un dashboard Grafana & compréhension des PromQL

## 🎯 Présentation

Cette étape introduit :
- Grafana,
- organisation dashboards,
- visualisations,
- PromQL,
- compréhension des métriques Kubernetes.

But :
- superviser le control-plane Kubernetes,
- visualiser les stress tests,
- comprendre le comportement interne du cluster.

---

## 🏗️ Architecture & emplacement de travail

### Architecture supervision

```text
Prometheus
→ collecte métriques
→ Grafana
→ Dashboards
```

### Structure Grafana

```text
Dashboard
→ Tabs
→ Rows
→ Panels
```

### Onglets créés

```text
OVERVIEW
CONTROL-PLANE
```

---

## ⚙️ Actions effectuées

### Construction panels

Panels créés :
- API CPU,
- API RAM,
- API QPS,
- ETCD latency,
- Node pressure,
- Network,
- Disk IO.

### Compréhension PromQL

Fonctions étudiées :

```promql
rate()
sum()
histogram_quantile()
```

### Exemple CPU API Server

```promql
sum(rate(container_cpu_usage_seconds_total{
pod=~"kube-apiserver.*"
}[5m]))
```

### Exemple mémoire API Server

```promql
sum(container_memory_working_set_bytes{
pod=~"kube-apiserver.*"
}) / 1024 / 1024
```

### Exemple percentile ETCD

```promql
histogram_quantile(
0.99,
sum(rate(etcd_disk_backend_commit_duration_seconds_bucket[5m])) by (le)
)
```

---

## 🔎 Vérifications

### Vérifier dashboards

Dans Grafana :

```text
Dashboards
```

### Vérifier métriques Prometheus

```promql
up
```

### Vérifier targets

```text
Prometheus > Targets
```

---

## ✅ Bilan

Le cluster dispose désormais :
- dashboards Grafana structurés,
- supervision control-plane,
- visualisation stress tests,
- compréhension approfondie PromQL et métriques Kubernetes.