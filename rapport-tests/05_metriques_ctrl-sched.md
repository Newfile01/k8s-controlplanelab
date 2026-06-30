# 📊 Métriques d'observation

## Scheduler

Le Scheduler est responsable de la sélection du nœud sur lequel chaque Pod sera exécuté. Les métriques suivantes permettront d'évaluer sa capacité à absorber une augmentation progressive ou brutale de la charge de planification.

## Métriques

### 📈 Charge

- **Tentatives de scheduling** : `scheduler_schedule_attempts_total`  
Nombre de décisions de placement effectuées
- **Pods en attente** : `scheduler_pending_pods`  
Quantité de Pods en attente de traitement
- **Taille du cache Scheduler** : `scheduler_cache_size`  
Nombre d'objets actuellement maintenus dans le cache

### ⚙️ Activité

- **Durée complète d'un scheduling** : `scheduler_scheduling_attempt_duration_seconds_bucket`  
Temps nécessaire pour planifier un Pod
- **Durée de l'algorithme de décision** : `scheduler_scheduling_algorithm_duration_seconds_bucket`  
Temps réellement consacré au choix du nœud
- **Résultat des tentatives** : `scheduler_schedule_attempts_total{result=...}`  
Répartition des planifications réussies ou échouées

### 🚨 Saturation

- **Pods Unschedulable** : `scheduler_schedule_attempts_total{result="unschedulable"}`  
Détection des Pods impossibles à planifier
- **Backoff Queue** : `scheduler_pending_pods{queue="backoff"}`  
Saturation liée aux nouvelles tentatives
- **Active Queue** : `scheduler_pending_pods{queue="active"}`  
Charge instantanée du Scheduler
- **Gated Queue** : `scheduler_pending_pods{queue="gated"}`  
Pods volontairement retenus avant planification

### Corrélations

| Comparaison | Objectif |
|-------------|----------|
| `scheduler_schedule_attempts_total` vs `scheduler_scheduling_attempt_duration_seconds_bucket` | Influence du nombre de planifications sur la latence |
| `scheduler_schedule_attempts_total` vs CPU Scheduler | Coût CPU de la planification |
| `scheduler_pending_pods` vs `scheduler_scheduling_attempt_duration_seconds_bucket` | Détection d'une saturation progressive |
| `scheduler_cache_size` vs Temps de scheduling | Influence de la taille du cluster |
| Pods créés par l'opérateur vs Tentatives de scheduling | Vérifier que la charge injectée est bien absorbée |
| Pods créés vs Pods Unschedulable | Identifier les limites de planification |

## Métriques prioritaires

- `scheduler_schedule_attempts_total`
- `scheduler_scheduling_attempt_duration_seconds_bucket`
- `scheduler_pending_pods`
- `scheduler_cache_size`

## Type de requêtes PromQL

| Type de métrique | Forme recommandée |
|------------------|-------------------|
| Counter | `sum(rate(...[5m]))` |
| Histogram | `histogram_quantile(0.99, sum by(le)(rate(..._bucket[5m])))` |
| Gauge | Valeur instantanée |

---

# Controller Manager

Le Controller Manager assure la convergence de l'état réel du cluster vers l'état désiré. Son comportement est directement influencé par la fréquence des réconciliations, le nombre de ressources manipulées et la quantité d'évènements produits par le cluster.

## Métriques

### 📈 Charge

- **Débit des workqueues** : `workqueue_adds_total`  
Nombre d'éléments ajoutés dans les files de traitement
- **Débit par contrôleur** : `workqueue_adds_total{name=...}`  
Répartition de la charge entre les différents contrôleurs

### ⚙️ Activité

- **Profondeur des workqueues** : `workqueue_depth`  
Nombre d'éléments restant à traiter
- **Temps passé en file d'attente** : `workqueue_queue_duration_seconds_bucket`  
Temps avant prise en charge d'un élément
- **Temps de traitement** : `workqueue_work_duration_seconds_bucket`  
Temps nécessaire pour traiter une réconciliation
- **Retries** : `workqueue_retries_total`  
Nombre de réconciliations ayant nécessité une nouvelle tentative
- **Retries Deployment** : `workqueue_retries_total{name="deployment"}`  
Difficultés spécifiques aux Deployments
- **Retries ReplicaSet** : `workqueue_retries_total{name="replicaset"}`  
Difficultés spécifiques aux ReplicaSets
- **Retries Garbage Collector** : `workqueue_retries_total{name=~"garbage.*"}`  
Réconciliations liées au Garbage Collector

### 🚨 Saturation

- **Profondeur des files** : `workqueue_depth`  
Détection d'un engorgement
- **Temps d'attente p99** : `workqueue_queue_duration_seconds_bucket`  
Allongement du délai avant traitement
- **Temps de traitement p99** : `workqueue_work_duration_seconds_bucket`  
Dégradation des performances des contrôleurs
- **Taux de retries** : `workqueue_retries_total`  
Perte de convergence du cluster

### Corrélations

| Comparaison | Objectif |
|-------------|----------|
| `workqueue_adds_total` vs `workqueue_depth` | Vérifier si les files grossissent plus vite qu'elles ne sont vidées |
| `workqueue_depth` vs `workqueue_queue_duration_seconds_bucket` | Impact de la saturation sur les délais |
| `workqueue_adds_total` vs CPU Controller Manager | Coût CPU des réconciliations |
| `workqueue_retries_total` vs `workqueue_depth` | Détection d'une non-convergence |
| `workqueue_retries_total` vs Erreurs de l'opérateur | Identifier une éventuelle propagation des erreurs |
| Réconciliations de l'opérateur vs `workqueue_adds_total` | Mesurer l'impact direct de l'opérateur sur le Controller Manager |

## Métriques prioritaires

- `workqueue_adds_total`
- `workqueue_depth`
- `workqueue_queue_duration_seconds_bucket`
- `workqueue_work_duration_seconds_bucket`
- `workqueue_retries_total`

## Type de requêtes PromQL

| Type de métrique | Forme recommandée |
|------------------|-------------------|
| Counter | `sum(rate(...[5m]))` |
| Histogram | `histogram_quantile(0.99, sum by(le)(rate(..._bucket[5m])))` |
| Gauge | Valeur instantanée |