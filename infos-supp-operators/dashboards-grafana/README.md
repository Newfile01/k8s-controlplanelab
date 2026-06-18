# 🧩 Structure des dashboards Grafana par scénario de stress

L’objectif est d’avoir :

* 1 dashboard dédié par composant stressé
* 2 lignes focalisées sur le composant cible
* 1 ligne macro sur les autres composants du control-plane
* 1 ligne dédiée au comportement de l’opérateur Kubernetes

Chaque dashboard peut ensuite être exporté en JSON Grafana et intégré dans :

```text
templates/dashboards/
```

---

# 📡 Dashboard API Server Stress

## 🧱 Ligne 1 — Charge API Server

| Panel              | Type       | PromQL                                                                                       |
| ------------------ | ---------- | -------------------------------------------------------------------------------------------- |
| Latence API Server | TimeSeries | `histogram_quantile(0.99, sum(rate(apiserver_request_duration_seconds_bucket[5m])) by (le))` |
| QPS API Server     | TimeSeries | `sum(rate(apiserver_request_total[5m]))`                                                     |
| Events API Server  | TimeSeries | `sum(rate(apiserver_audit_event_total[5m]))`                                                 |

---

## 🧱 Ligne 2 — Activité API détaillée

| Panel               | Type       | PromQL                                                               |
| ------------------- | ---------- | -------------------------------------------------------------------- |
| I/O Disk API Server | TimeSeries | `rate(container_fs_writes_bytes_total{pod=~"kube-apiserver.*"}[5m])` |
| Requests inflight   | Gauge      | `sum(apiserver_current_inflight_requests)`                           |
| Requests par verb   | BarGauge   | `sum(rate(apiserver_request_total[5m])) by (verb)`                   |

---

## 🧱 Ligne 3 — Vue macro control-plane

| Composant          | Métriques                                 |
| ------------------ | ----------------------------------------- |
| Controller Manager | CPU / RAM / workqueue / reconcile latency |
| ETCD               | CPU / RAM / commit latency / db size      |
| Scheduler          | CPU / RAM / scheduling latency            |

---

## 🧱 Ligne 4 — Vue opérateur

| Panel            | Type       |
| ---------------- | ---------- |
| Reconcile rate   | TimeSeries |
| Active workers   | Gauge      |
| Reconcile errors | Stat       |
| Operator CPU/RAM | Gauge      |

---

# 🗓️ Dashboard Scheduler Stress

## 🧱 Ligne 1 — Activité scheduler

| Panel               | Type       | PromQL                                                                                              |
| ------------------- | ---------- | --------------------------------------------------------------------------------------------------- |
| Scheduling latency  | TimeSeries | `histogram_quantile(0.99, sum(rate(scheduler_e2e_scheduling_duration_seconds_bucket[5m])) by (le))` |
| Pods pending        | TimeSeries | `sum(scheduler_pending_pods)`                                                                       |
| Scheduling attempts | TimeSeries | `sum(rate(scheduler_schedule_attempts_total[5m]))`                                                  |

---

## 🧱 Ligne 2 — Placement Pods

| Panel             | Type       | PromQL                                                                           |
| ----------------- | ---------- | -------------------------------------------------------------------------------- |
| Failed scheduling | TimeSeries | `sum(rate(scheduler_pod_scheduling_attempts_total{result="unschedulable"}[5m]))` |
| Node saturation   | Gauge      | `sum(kube_node_status_allocatable_cpu_cores)`                                    |
| Pods par node     | BarGauge   | `count(kube_pod_info) by (node)`                                                 |

---

## 🧱 Ligne 3 — Vue macro control-plane

| Composant          | Métriques                     |
| ------------------ | ----------------------------- |
| API Server         | CPU / RAM / inflight requests |
| ETCD               | commit latency / db size      |
| Controller Manager | workqueues / deployments      |

---

## 🧱 Ligne 4 — Vue opérateur

| Panel             | Type       |
| ----------------- | ---------- |
| Reconcile latency | TimeSeries |
| Queue depth       | Gauge      |
| Operator CPU      | Gauge      |
| Events Kubernetes | TimeSeries |

---

# 💾 Dashboard ETCD Stress

## 🧱 Ligne 1 — Écriture ETCD

| Panel                 | Type       | PromQL                                                                                              |
| --------------------- | ---------- | --------------------------------------------------------------------------------------------------- |
| Latence écriture ETCD | TimeSeries | `histogram_quantile(0.99, sum(rate(etcd_disk_backend_commit_duration_seconds_bucket[5m])) by (le))` |
| QPS ETCD              | TimeSeries | `sum(rate(etcd_server_proposals_applied_total[5m]))`                                                |
| Taille DB ETCD        | Gauge      | `etcd_mvcc_db_total_size_in_bytes / 1024 / 1024`                                                    |

---

## 🧱 Ligne 2 — Activité ETCD détaillée

| Panel             | Type       | PromQL                                                                                         |
| ----------------- | ---------- | ---------------------------------------------------------------------------------------------- |
| Read latency      | TimeSeries | `histogram_quantile(0.99, sum(rate(etcd_request_duration_seconds_bucket[5m])) by (le))`        |
| WAL fsync latency | TimeSeries | `histogram_quantile(0.99, sum(rate(etcd_disk_wal_fsync_duration_seconds_bucket[5m])) by (le))` |
| Requests par verb | BarGauge   | `sum(rate(apiserver_request_total[5m])) by (verb)`                                             |

---

## 🧱 Ligne 3 — Vue macro control-plane

| Composant          | Métriques                |
| ------------------ | ------------------------ |
| API Server         | latency / inflight / CPU |
| Scheduler          | pending pods / latency   |
| Controller Manager | workqueue depth          |

---

## 🧱 Ligne 4 — Vue opérateur

| Panel                 | Type       |
| --------------------- | ---------- |
| Reconcile rate        | TimeSeries |
| API requests operator | TimeSeries |
| Operator RAM          | Gauge      |
| Reconcile errors      | Stat       |

---

# 🔁 Dashboard Controller Manager Stress

## 🧱 Ligne 1 — Activité controller-manager

| Panel                      | Type       | PromQL                                                |
| -------------------------- | ---------- | ----------------------------------------------------- |
| Workqueue depth            | TimeSeries | `sum(workqueue_depth)`                                |
| Requeues total             | TimeSeries | `sum(rate(workqueue_retries_total[5m]))`              |
| Deployment reconciliations | TimeSeries | `sum(rate(deployment_controller_requeues_total[5m]))` |

---

## 🧱 Ligne 2 — Garbage collection & ReplicaSets

| Panel                  | Type       | PromQL                                                                  |
| ---------------------- | ---------- | ----------------------------------------------------------------------- |
| ReplicaSet recreations | TimeSeries | `sum(rate(replicaset_controller_requeues_total[5m]))`                   |
| Garbage collector      | TimeSeries | `sum(rate(garbagecollector_controller_resources_sync_error_total[5m]))` |
| Events Kubernetes      | TimeSeries | `sum(rate(apiserver_audit_event_total[5m]))`                            |

---

## 🧱 Ligne 3 — Vue macro control-plane

| Composant  | Métriques          |
| ---------- | ------------------ |
| API Server | latency / inflight |
| ETCD       | commits / db size  |
| Scheduler  | scheduling latency |

---

## 🧱 Ligne 4 — Vue opérateur

| Panel              | Type       |
| ------------------ | ---------- |
| Reconcile duration | TimeSeries |
| Queue depth        | Gauge      |
| Active workers     | Gauge      |
| Operator CPU/RAM   | Gauge      |

---

# 🤖 Dashboard Operator Stress

## 🧱 Ligne 1 — Activité operator/controller-runtime

| Panel            | Type       | PromQL                                                     |
| ---------------- | ---------- | ---------------------------------------------------------- |
| Reconcile total  | TimeSeries | `sum(rate(controller_runtime_reconcile_total[5m]))`        |
| Reconcile errors | TimeSeries | `sum(rate(controller_runtime_reconcile_errors_total[5m]))` |
| Active workers   | Gauge      | `controller_runtime_active_workers`                        |

---

## 🧱 Ligne 2 — Performance operator

| Panel             | Type       | PromQL                                                                                              |
| ----------------- | ---------- | --------------------------------------------------------------------------------------------------- |
| Reconcile latency | TimeSeries | `histogram_quantile(0.99, sum(rate(controller_runtime_reconcile_time_seconds_bucket[5m])) by (le))` |
| Operator CPU      | Gauge      | `sum(rate(container_cpu_usage_seconds_total{pod=~"operator-controller-manager.*"}[5m]))`            |
| Operator RAM      | Gauge      | `sum(container_memory_working_set_bytes{pod=~"operator-controller-manager.*"}) / 1024 / 1024`       |

---

## 🧱 Ligne 3 — Vue macro control-plane

| Composant  | Métriques                         |
| ---------- | --------------------------------- |
| API Server | request rate / latency            |
| ETCD       | write latency / db size           |
| Scheduler  | pending pods / scheduling latency |

---

## 🧱 Ligne 4 — Vue opérateur détaillée

| Panel                   | Type       |
| ----------------------- | ---------- |
| Watch events            | TimeSeries |
| Kubernetes API requests | TimeSeries |
| Requeue spikes          | TimeSeries |
| Runtime goroutines      | Gauge      |
