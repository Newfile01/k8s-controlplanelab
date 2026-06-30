# 📊 Métriques d'observation

## Opérateur Kubernetes

L'opérateur constitue la source de charge du benchmark. Les métriques suivantes permettront de caractériser précisément son activité et de la corréler avec le comportement du Control Plane.

## Métriques

## 📈 Charge

- **Nombre de réconciliations** : `controlplanetest_reconciliation_total`  
Activité globale de l'opérateur
- **Nombre de Deployments générés** : `controlplanetest_deployments_generes`  
Charge injectée dans le Scheduler et le Controller Manager
- **Nombre de Pods désirés** : `controlplanetest_pods_desires`  
Taille cible du scénario
- **Nombre de Pods supprimés** : `controlplanetest_pods_supprimes_total`  
Intensité des scénarios de suppression
- **Nombre de recréations forcées** : `controlplanetest_resources_recreated_total`  
Intensité des scénarios Delete/Recreate
- **Nombre de Status Update** : `controlplanetest_status_updates_total`  
Charge d'écriture générée sur l'API Server
- **Nombre de Requeues forcées** : `controlplanetest_requeues_forcees_total`  
Charge volontaire de réconciliation

## ⚙️ Activité

- **Pods actuellement gérés** : `controlplanetest_pods_geres`  
Taille réelle de la charge gérée
- **Réplicas disponibles** : `controlplanetest_replicas_disponibles`  
État réel du déploiement
- **Configuration active** : `controlplanetest_configuration_info`  
Paramètres du scénario exécuté
- **Durée des réconciliations** : `controlplanetest_duree_reconciliation_secondes_bucket`  
Temps de traitement d'une réconciliation
- **Workers actifs** : `controller_runtime_active_workers`  
Niveau d'utilisation des workers du contrôleur
- **Workers maximum** : `controller_runtime_max_concurrent_reconciles`  
Capacité maximale du contrôleur
- **Réconciliations Controller Runtime** : `controller_runtime_reconcile_total`  
Activité interne du framework controller-runtime

## 🚨 Saturation

- **Erreurs de réconciliation** : `controlplanetest_erreurs_reconciliation_total`  
Difficultés rencontrées par l'opérateur
- **Erreurs Controller Runtime** : `controller_runtime_reconcile_errors_total`  
Erreurs internes du framework
- **Timeout de réconciliation** : `controller_runtime_reconcile_timeouts_total`  
Réconciliations trop longues
- **Panic** : `controller_runtime_reconcile_panics_total`  
Détection d'erreurs critiques
- **Durée p99 des réconciliations** : `controller_runtime_reconcile_time_seconds_bucket`  
Dégradation des performances




### Corrélations


| Comparaison                                                                         | Objectif                                                |
| ----------------------------------------------------------------------------------- | ------------------------------------------------------- |
| `controlplanetest_reconciliation_total` vs `controller_runtime_reconcile_total`     | Vérifier la cohérence entre logique métier et framework |
| `controlplanetest_reconciliation_total` vs `apiserver_request_total`                | Impact des réconciliations sur l'API Server             |
| `controlplanetest_status_updates_total` vs `etcd_requests_total`                    | Impact des Status Update sur ETCD                       |
| `controlplanetest_resources_recreated_total` vs `scheduler_schedule_attempts_total` | Effet des recréations sur le Scheduler                  |
| `controlplanetest_resources_recreated_total` vs `workqueue_adds_total`              | Effet des recréations sur le Controller Manager         |
| `controller_runtime_active_workers` vs Durée de réconciliation                      | Saturation du contrôleur                                |
| `controlplanetest_pods_desires` vs `controlplanetest_pods_geres`                    | Vérifier que la cible est effectivement atteinte        |
| `controlplanetest_requeues_forcees_total` vs `workqueue_depth`                      | Effet des requeues sur les files de traitement          |




## Métriques prioritaires

- `controlplanetest_reconciliation_total`
- `controlplanetest_duree_reconciliation_secondes_bucket`
- `controlplanetest_status_updates_total`
- `controlplanetest_resources_recreated_total`
- `controller_runtime_reconcile_total`
- `controller_runtime_reconcile_time_seconds_bucket`
- `controller_runtime_active_workers`



## Type de requêtes PromQL


| Type de métrique | Forme recommandée                                            |
| ---------------- | ------------------------------------------------------------ |
| Counter          | `sum(rate(...[5m]))`                                         |
| Histogram        | `histogram_quantile(0.99, sum by(le)(rate(..._bucket[5m])))` |
| Gauge            | valeur instantanée                                           |


---



# Cluster Kubernetes

Ces métriques permettent d'observer l'état global du cluster et de relier les performances du Control Plane à la charge réellement présente sur les nœuds.

## Métriques


## 📈 Charge

- **Pods Running** : `kubelet_working_pods`  
Nombre de Pods réellement en exécution
- **Pods désirés** : `kubelet_desired_pods`  
Charge attendue par les kubelets
- **Conteneurs Running** : `kubelet_running_containers`  
Charge réellement exécutée
- **Nombre total d'objets Kubernetes** : `apiserver_resource_objects`  
Taille globale du cluster

## ⚙️ Activité

- **CPU API Server** : `process_cpu_seconds_total`  
Consommation CPU de l'API Server
- **CPU ETCD** : `process_cpu_seconds_total`  
Consommation CPU d'ETCD
- **CPU Scheduler** : `process_cpu_seconds_total`  
Consommation CPU du Scheduler
- **CPU Controller Manager** : `process_cpu_seconds_total`  
Consommation CPU du Controller Manager
- **Mémoire des composants** : `process_resident_memory_bytes`  
Consommation mémoire des composants

## 📊 État

- **Nombre de nœuds Ready** : `kube_node_status_condition`  
Disponibilité du cluster
- **Nombre de Pods** : `apiserver_resource_objects{resource="pods"}`  
Taille actuelle du cluster en nombre de Pods
- **Nombre de Deployments** : `apiserver_resource_objects{resource="deployments"}`  
Taille actuelle du cluster en nombre de Deployments

## 🚨 Saturation

- **CPU des nœuds** : `node_cpu_seconds_total`  
Détection d'une saturation CPU des nœuds du cluster
- **Mémoire des nœuds** : `node_memory_MemAvailable_bytes`  
Détection d'une saturation mémoire des nœuds
- **Pression disque** : `node_filesystem_avail_bytes`  
Détection d'une saturation de l'espace de stockage
- **Charge moyenne** : `node_load1`  
Évaluation de la charge système globale des nœuds




### Corrélations


| Comparaison                               | Objectif                            |
| ----------------------------------------- | ----------------------------------- |
| Pods Running vs CPU API Server            | Influence de la taille du cluster   |
| Pods Running vs CPU ETCD                  | Charge induite sur ETCD             |
| Pods Running vs CPU Scheduler             | Charge de planification             |
| Pods Running vs CPU Controller Manager    | Charge de réconciliation            |
| Nombre d'objets vs Taille ETCD            | Croissance du stockage              |
| CPU API Server vs Latence API p99         | Apparition d'une saturation         |
| CPU ETCD vs Backend Commit p99            | Dégradation des performances disque |
| CPU Scheduler vs Temps de scheduling p99  | Saturation de la planification      |
| CPU Controller Manager vs Workqueue Depth | Saturation des contrôleurs          |




## Métriques prioritaires

- `kubelet_working_pods`
- `apiserver_resource_objects`
- `process_cpu_seconds_total`
- `process_resident_memory_bytes`
- `node_cpu_seconds_total`
- `node_memory_MemAvailable_bytes`



## Type de requêtes PromQL


| Type de métrique | Forme recommandée                                            |
| ---------------- | ------------------------------------------------------------ |
| Counter          | `sum(rate(...[5m]))`                                         |
| Histogram        | `histogram_quantile(0.99, sum by(le)(rate(..._bucket[5m])))` |
| Gauge            | valeur instantanée                                           |


---



# Synthèse des formes de requêtes PromQL


| Famille de métriques | Forme de requête recommandée                                    | Utilisation                                         |
| -------------------- | --------------------------------------------------------------- | --------------------------------------------------- |
| Counter              | `sum(rate(metric[5m]))`                                         | Débit (requêtes/s, réconciliations/s, événements/s) |
| Histogram (_bucket)  | `histogram_quantile(0.50, sum by(le)(rate(metric_bucket[5m])))` | Latence médiane (p50)                               |
| Histogram (_bucket)  | `histogram_quantile(0.95, sum by(le)(rate(metric_bucket[5m])))` | Latence p95                                         |
| Histogram (_bucket)  | `histogram_quantile(0.99, sum by(le)(rate(metric_bucket[5m])))` | Latence p99                                         |
| Histogram (_count)   | `sum(rate(metric_count[5m]))`                                   | Nombre d'opérations                                 |
| Histogram (_sum)     | `sum(rate(metric_sum[5m]))`                                     | Temps cumulé                                        |
| Gauge                | Valeur instantanée                                              | État courant                                        |
| Counter cumulatif    | `increase(metric[1h])`                                          | Nombre d'évènements sur une période                 |


> [!IMPORTANT]
> Pour chaque scénario de test, l'ensemble des métriques présentées dans ce document sera relevé. Bien qu'un scénario cible principalement un composant du Control Plane, tous les composants seront observés afin d'identifier les corrélations et les effets indirects induits par la charge générée.

