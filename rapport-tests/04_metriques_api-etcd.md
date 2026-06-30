# 📊 Métriques d'observation

## Objectif

Les métriques suivantes seront collectées lors de l'ensemble des campagnes de tests.

Bien que chaque scénario cible principalement un composant particulier du Control Plane, l'ensemble des composants sera systématiquement observé afin d'identifier les interactions entre eux et d'étudier les effets indirects produits par chaque stratégie de charge.

Nous avons regroupées les métriques par composant que nous cherchons à étudier : ceux du control plane et notre opérateur.

---

# API Server

L'API Server constitue le point d'entrée de toutes les opérations réalisées sur le cluster Kubernetes. Les métriques suivantes permettront de mesurer la charge reçue, son activité interne, son état ainsi que l'apparition éventuelle de phénomènes de saturation.

## Métriques

### 📈 Charge

- **Requêtes API/s** : `apiserver_request_total`  
Charge globale appliquée à l'API Server

- **Requêtes API/s par verbe** : `apiserver_request_total{verb=...}`  
Répartition des opérations GET, LIST, WATCH, POST, PATCH et DELETE

- **Requêtes API/s par ressource** : `apiserver_request_total{resource=...}`  
Identification des ressources Kubernetes les plus sollicitées

### ⚙️ Activité

- **Requêtes simultanées** : `apiserver_current_inflight_requests`  
Nombre de requêtes traitées simultanément

- **Objets demandés au Watch Cache** : `apiserver_cache_list_fetched_objects_total`  
Pression de lecture exercée sur le Watch Cache

- **Objets retournés par le Watch Cache** : `apiserver_cache_list_returned_objects_total`  
Efficacité du Watch Cache

- **Évènements WATCH** : `apiserver_watch_events_total`  
Activité de diffusion des changements aux différents clients Kubernetes

### 📊 État

- **Nombre total d'objets Kubernetes** : `apiserver_resource_objects`  
Taille logique du cluster

- **Nombre d'objets par ressource** : `apiserver_resource_objects{resource=...}`  
Répartition des différentes ressources Kubernetes

- **Taille estimée des ressources** : `apiserver_resource_size_estimate_bytes`  
Impact des manifests volumineux

- **Taille physique de la base ETCD** : `apiserver_storage_size_bytes`  
Croissance réelle du stockage utilisé

### 🚨 Saturation

- **Latence API p50** : `apiserver_request_duration_seconds_bucket`  
Temps de réponse nominal

- **Latence API p95** : `apiserver_request_duration_seconds_bucket`  
Apparition des premiers ralentissements

- **Latence API p99** : `apiserver_request_duration_seconds_bucket`  
Détection d'une saturation du composant

- **Erreurs HTTP 4xx** : `apiserver_request_total{code=~"4.."}`  
Erreurs côté client

- **Erreurs HTTP 5xx** : `apiserver_request_total{code=~"5.."}`  
Erreurs internes de l'API Server

- **Erreurs HTTP 429** : `apiserver_request_total{code="429"}`  
Limitation du nombre de requêtes (Rate Limiting)

### Corrélations

| Comparaison | Objectif |
|-------------|----------|
| `apiserver_request_total` vs `apiserver_request_duration_seconds_bucket` | Impact de la charge sur la latence |
| `apiserver_request_total` vs CPU API Server | Coût CPU des traitements |
| `apiserver_resource_objects` vs `apiserver_request_duration_seconds_bucket` | Influence de la taille du cluster |
| `apiserver_resource_size_estimate_bytes` vs `apiserver_storage_size_bytes` | Impact des gros manifests |

## Métriques prioritaires

- `apiserver_request_total`
- `apiserver_request_duration_seconds_bucket`
- `apiserver_current_inflight_requests`
- `apiserver_resource_objects`
- `apiserver_storage_size_bytes`

## Type de requêtes PromQL

| Type de métrique | Forme recommandée |
|------------------|-------------------|
| Counter | `sum(rate(...[5m]))` |
| Histogram | `histogram_quantile(0.99, sum by(le)(rate(..._bucket[5m])))` |
| Gauge | Valeur instantanée |

---

# ETCD

ETCD constitue la base de données (potentiellement distribuée) du cluster Kubernetes. Les métriques suivantes permettront d'évaluer son activité, l'évolution du stockage ainsi que l'impact des opérations de lecture, d'écriture et de surveillance (WATCH).

## Métriques

### 📈 Charge

- **Requêtes ETCD/s** : `etcd_requests_total`  
Activité globale de la base de données

- **Requêtes clientes** : `etcd_server_client_requests_total`  
Charge envoyée à ETCD par les composants du Control Plane

### ⚙️ Activité

- **Activité MVCC** : `etcd_debugging_mvcc_events_total`  
Nombre de révisions et modifications enregistrées

- **Lectures internes** : `etcd_debugging_store_reads_total`  
Charge de lecture exercée sur la base

- **Watchers actifs** : `etcd_debugging_mvcc_watcher_total`  
Nombre de clients actuellement abonnés aux changements

- **Flux WATCH** : `etcd_debugging_mvcc_watch_stream_total`  
Activité de diffusion des évènements aux watchers

### 📊 État

- **Taille de la base** : `apiserver_storage_size_bytes`  
Croissance physique de la base ETCD

- **Nombre de clés MVCC** : `etcd_debugging_mvcc_keys_total`  
Évolution du stockage interne

- **Taille totale des opérations PUT** : `etcd_debugging_mvcc_total_put_size_in_bytes`  
Volume logique des écritures

- **Octets écrits dans la WAL** : `etcd_disk_wal_write_bytes_total`  
Débit réel d'écriture sur le journal WAL

### 🚨 Saturation

- **Latence ETCD p99** : `etcd_request_duration_seconds_bucket`  
Temps de réponse global des requêtes

- **Backend Commit p99** : `etcd_disk_backend_commit_duration_seconds_bucket`  
Temps d'écriture dans la base BBoltDB

- **WAL fsync p99** : `etcd_disk_wal_fsync_duration_seconds_bucket`  
Temps de synchronisation des écritures disque

- **Changements de leader** : `etcd_server_leader_changes_seen_total`  
Détection d'une instabilité éventuelle du cluster ETCD

### Corrélations

| Comparaison | Objectif |
|-------------|----------|
| `apiserver_storage_size_bytes` vs `etcd_disk_backend_commit_duration_seconds_bucket` | Impact de la taille de la base sur les écritures |
| `apiserver_storage_size_bytes` vs `etcd_disk_wal_fsync_duration_seconds_bucket` | Impact de la croissance du stockage sur les performances disque |
| `etcd_requests_total` vs `etcd_request_duration_seconds_bucket` | Mise en évidence d'une saturation |
| `apiserver_resource_objects` vs `apiserver_storage_size_bytes` | Évolution du stockage en fonction du nombre d'objets Kubernetes |
| `etcd_debugging_mvcc_total_put_size_in_bytes` vs `etcd_disk_wal_write_bytes_total` | Impact des objets volumineux sur les écritures disque |

## Métriques prioritaires

- `etcd_requests_total`
- `etcd_request_duration_seconds_bucket`
- `etcd_disk_backend_commit_duration_seconds_bucket`
- `etcd_disk_wal_fsync_duration_seconds_bucket`
- `apiserver_storage_size_bytes`

## Type de requêtes PromQL

| Type de métrique | Forme recommandée |
|------------------|-------------------|
| Counter | `sum(rate(...[5m]))` |
| Histogram | `histogram_quantile(0.99, sum by(le)(rate(..._bucket[5m])))` |
| Gauge | Valeur instantanée |