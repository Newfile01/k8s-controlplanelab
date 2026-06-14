# ☸️ Kubernetes Control Plane Stress Operator

Opérateur Kubernetes développé avec Kubebuilder permettant de générer des scénarios de stress ciblant les composants du control-plane Kubernetes afin d'observer leur comportement, leur résilience et leurs limites sous charge.

L’opérateur peut influer sur :

| Composant | Effet généré |
|---|---|
| API Server | Tempêtes de requêtes et updates fréquentes |
| Scheduler | Forte pression de scheduling |
| ETCD | Grand volume d’écritures Kubernetes |
| Controller Manager | Réconciliations intensives |
| Operator | Requeues et reconciliations agressives |
| Nodes | Saturation Pods / scheduling / ressources |

---

# 🏗️ Architecture

```text
CR (ControlPlaneTest)
        ↓
Controller (Reconciler)
        ↓
Ressources Kubernetes générées
(Deployments, Services, ConfigMaps, Pods...)
```

## 📦 Composants principaux

| Élément | Rôle |
|---|---|
| `api/v1alpha1/controlplanetest_types.go` | Définition de la CRD |
| `internal/controller/controlplanetest_controller.go` | Logique de réconciliation |
| `config/crd/` | Génération CRD Kubernetes |
| `config/manager/` | Déploiement opérateur |
| `config/prometheus/` | Intégration Prometheus |
| `cmd/main.go` | Point d’entrée controller-runtime |

## 🧩 Ressources générées

### CRD

```yaml
kind: CustomResourceDefinition
metadata:
  name: controlplanetests.controlplane.lab.local
```

### CR manipulée

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest
```

---

# 🚀 Déploiement

## 🔨 Build et déploiement

```bash
make docker-build docker-push IMG=<registry>/operator-k8s:vX
make deploy IMG=<registry>/operator-k8s:vX
```

## 📋 Ressources générées

```bash
kubectl get all -A | grep operator-system
```

Exemple :

```text
operator-system    pod/operator-controller-manager-xxxxx
operator-system    service/operator-controller-manager-metrics-service
operator-system    deployment.apps/operator-controller-manager
operator-system    replicaset.apps/operator-controller-manager-xxxxx
```

## ⚙️ Fonctionnement

L’opérateur génère :
- le controller manager,
- les RBAC nécessaires,
- la CRD,
- le service metrics Prometheus,
- les watches Kubernetes.

Il surveille ensuite :
- les CR `ControlPlaneTest`,
- les Deployments générés,
- les Services,
- les ConfigMaps,
- les Pods associés.

---

# 🧪 Configuration de la CRD
| Champ | | | Description |
|---|---|---|---|
| Niveau 1 | Niveau 2 | Niveau 3 | |
| SchedulerStressSpec (struct) | | | F : Génère une forte activité du scheduler Kubernetes via la création massive de Deployments et Pods répartis sur plusieurs noeuds |
| | | | V : {} |
| | | | O : Impact visible sur scheduler, etcd, controller-manager et kubelet |
| | | | M : scheduler_pending_pods \| scheduler_e2e_scheduling_duration_seconds \| process_cpu_seconds_total |
| | .Enabled (bool) | | F : Active ou désactive le scénario de stress scheduler |
| | | | V : false |
| | | | O : Déclenche la création des ressources de stress scheduler |
| | | | M : scheduler_schedule_attempts_total |
| | .NodeCount (int32) | | F : Nombre de noeuds ciblés pour la répartition des Pods |
| | | | V : 0 |
| | | | O : Influence la dispersion des Pods dans le cluster |
| | | | M : kube_node_status_capacity \| scheduler_pending_pods |
| | .DeploymentCount (int32) | | F : Nombre de Deployments créés simultanément |
| | | | V : 0 |
| | | | O : Augmente fortement les objets surveillés par Kubernetes |
| | | | M : apiserver_request_total \| workqueue_depth |
| | .ReplicasPerDeployment (int32) | | F : Nombre de Pods créés par Deployment |
| | | | V : 0 |
| | | | O : Augmente directement la charge du scheduler et du kubelet |
| | | | M : scheduler_pod_scheduling_attempts \| kube_pod_status_phase |
| | .NodeSelector (map[string]string) | | F : Contraintes de placement imposées aux Pods via labels Kubernetes |
| | | | V : {} |
| | | | O : Peut provoquer des Pods Pending si aucun noeud compatible |
| | | | M : scheduler_pending_pods |
| | .TopologySpread (bool) | | F : Active les topologySpreadConstraints pour équilibrer les Pods |
| | | | V : false |
| | | | O : Complexifie les décisions du scheduler |
| | | | M : scheduler_framework_extension_point_duration_seconds |
| | .Affinity (string) | | F : Définit des règles d'affinité de placement des Pods |
| | | | V : "" |
| | | | O : Augmente les calculs de placement du scheduler |
| | | | M : scheduler_scheduling_algorithm_duration_seconds |
| | .AntiAffinity (string) | | F : Définit des règles anti-affinité empêchant certains placements |
| | | | V : "" |
| | | | O : Peut fortement ralentir le scheduling sur clusters chargés |
| | | | M : scheduler_pending_pods \| scheduler_e2e_scheduling_duration_seconds |
| APIServerStressSpec (struct) | | | F : Génère une forte activité sur l'API Kubernetes via requêtes et reconciliations agressives |
| | | | V : {} |
| | | | O : Impact direct sur API Server et ETCD |
| | | | M : apiserver_request_total \| apiserver_current_inflight_requests |
| | .Enabled (bool) | | F : Active le stress API Server |
| | | | V : false |
| | | | O : Déclenche les scénarios API intensifs |
| | | | M : apiserver_request_total |
| | .FrequentStatusUpdates (bool) | | F : Effectue des mises à jour fréquentes du champ status des CRs |
| | | | V : false |
| | | | O : Génère beaucoup d'écritures ETCD |
| | | | M : etcd_request_duration_seconds |
| | .AggressiveReconcile (bool) | | F : Force des reconciliations très fréquentes |
| | | | V : false |
| | | | O : Augmente fortement les appels API Kubernetes |
| | | | M : controller_runtime_reconcile_total |
| | .RecreateResources (bool) | | F : Supprime et recrée régulièrement les ressources Kubernetes |
| | | | V : false |
| | | | O : Génère beaucoup d'events et d'opérations ETCD |
| | | | M : apiserver_request_total \| etcd_disk_backend_commit_duration_seconds |
| | .QPS (int32) | | F : Limite QPS utilisée par le client Kubernetes |
| | | | V : 0 |
| | | | O : Plus élevé = plus de requêtes API par seconde |
| | | | M : apiserver_request_total |
| | .Burst (int32) | | F : Nombre maximal de requêtes burst avant throttling |
| | | | V : 0 |
| | | | O : Génère des pics brutaux de trafic API |
| | | | M : apiserver_current_inflight_requests |
| EtcdStressSpec (struct) | | | F : Génère de nombreuses écritures ETCD via ConfigMaps et Secrets |
| | | | V : {} |
| | | | O : Charge fortement le backend disque ETCD |
| | | | M : etcd_disk_backend_commit_duration_seconds \| etcd_mvcc_db_total_size_in_bytes |
| | .Enabled (bool) | | F : Active le stress ETCD |
| | | | V : false |
| | | | O : Déclenche les écritures massives dans ETCD |
| | | | M : etcd_server_proposals_applied_total |
| | .ConfigMapCount (int32) | | F : Nombre de ConfigMaps créées |
| | | | V : 0 |
| | | | O : Augmente fortement le nombre d'objets stockés dans ETCD |
| | | | M : etcd_mvcc_db_total_size_in_bytes |
| | .ConfigMapSizeKB (int32) | | F : Taille des ConfigMaps générées |
| | | | V : 0 |
| | | | O : Impacte directement les écritures disque ETCD |
| | | | M : etcd_disk_backend_commit_duration_seconds |
| | .SecretCount (int32) | | F : Nombre de Secrets générés |
| | | | V : 0 |
| | | | O : Charge supplémentaire sur sérialisation et stockage ETCD |
| | | | M : etcd_server_proposals_applied_total |
| | .SecretSizeKB (int32) | | F : Taille des Secrets générés |
| | | | V : 0 |
| | | | O : Augmente taille mémoire et disque de la base ETCD |
| | | | M : etcd_mvcc_db_total_size_in_bytes |
| ControllerManagerStressSpec (struct) | | | F : Génère une activité importante du controller-manager Kubernetes |
| | | | V : {} |
| | | | O : Charge ReplicaSets, Deployments et garbage collector Kubernetes |
| | | | M : workqueue_depth \| workqueue_adds_total |
| | .Enabled (bool) | | F : Active le scénario controller-manager |
| | | | V : false |
| | | | O : Déclenche les contrôleurs Kubernetes associés |
| | | | M : workqueue_queue_duration_seconds |
| | .DeploymentCount (int32) | | F : Nombre de Deployments générés |
| | | | V : 0 |
| | | | O : Augmente les opérations de convergence Kubernetes |
| | | | M : deployment_controller_requeues_total |
| | .ReplicasPerDeployment (int32) | | F : Nombre de réplicas par Deployment |
| | | | V : 0 |
| | | | O : Augmente Pods et ReplicaSets gérés |
| | | | M : replicaset_controller_requeues_total |
| | .RecreateReplicaSets (bool) | | F : Supprime et recrée fréquemment les ReplicaSets |
| | | | V : false |
| | | | O : Génère beaucoup d'events Kubernetes |
| | | | M : workqueue_adds_total |
| | .AggressiveGarbageCollection (bool) | | F : Force un nettoyage fréquent des ressources supprimées |
| | | | V : false |
| | | | O : Impact visible sur controller-manager et ETCD |
| | | | M : garbagecollector_controller_resources_sync_error_total |
| OperatorStressSpec (struct) | | | F : Simule une forte activité interne de l'opérateur Kubernetes |
| | | | V : {} |
| | | | O : Charge controller-runtime, API Server et ETCD |
| | | | M : controller_runtime_reconcile_total \| controller_runtime_active_workers |
| | .Enabled (bool) | | F : Active le stress de l'opérateur |
| | | | V : false |
| | | | O : Déclenche les comportements agressifs de reconciliation |
| | | | M : controller_runtime_reconcile_total |
| | .Profile (string) | | F : Définit un profil de stress prédéfini |
| | | | V : "" |
| | | | O : Permet de charger plusieurs paramètres automatiquement |
| | | | M : dépend du profil utilisé |
| | .Informer (struct) | | F : Configure les watchers Kubernetes utilisés par l'opérateur |
| | | | V : {} |
| | | | O : Plus les watchers sont nombreux, plus les événements Kubernetes sont consommés |
| | | | M : controller_runtime_reconcile_total |
| | | .WatchPods (bool) | F : Active la surveillance des Pods |
| | | | V : false |
| | | | O : Génère beaucoup d'événements sur gros clusters |
| | | | M : controller_runtime_reconcile_total |
| | | .WatchConfigMaps (bool) | F : Active la surveillance des ConfigMaps |
| | | | V : false |
| | | | O : Très bavard avec ETCD stress |
| | | | M : controller_runtime_reconcile_total |
| | | .WatchDeployments (bool) | F : Active la surveillance des Deployments |
| | | | V : false |
| | | | O : Augmente fortement les reconciliations |
| | | | M : controller_runtime_reconcile_total |
| | .Reconcile (struct) | | F : Configure le comportement de reconciliation controller-runtime |
| | | | V : {} |
| | | | O : Impact direct sur CPU opérateur et trafic API |
| | | | M : controller_runtime_reconcile_time_seconds |
| | | .MaxConcurrent (int32) | F : Nombre maximal de reconciliations simultanées |
| | | | V : 0 |
| | | | O : Augmente le parallélisme du controller |
| | | | M : controller_runtime_active_workers |
| | | .QPS (int32) | F : Limite QPS du client Kubernetes |
| | | | V : 0 |
| | | | O : Plus élevé = plus de requêtes API simultanées |
| | | | M : apiserver_request_total |
| | | .Burst (int32) | F : Nombre maximal de requêtes burst |
| | | | V : 0 |
| | | | O : Génère des pics de trafic API |
| | | | M : apiserver_current_inflight_requests |
| | | .BaseDelaySeconds (int32) | F : Délai minimal entre deux reconciliations |
| | | | V : 0 |
| | | | O : Plus faible = comportement plus agressif |
| | | | M : controller_runtime_reconcile_time_seconds |
| | | .MaxDelaySeconds (int32) | F : Délai maximal du backoff de reconciliation |
| | | | V : 0 |
| | | | O : Contrôle les retries après erreur |
| | | | M : controller_runtime_reconcile_errors_total |
| PodLifecycleStormSpec (struct) | | | F : Génère des perturbations constantes du cycle de vie des Pods |
| | | | V : {} |
| | | | O : Charge fortement scheduler, kubelet et controller-manager |
| | | | M : kube_pod_container_status_restarts_total \| scheduler_pending_pods |
| | .Enabled (bool) | | F : Active le scénario Pod lifecycle storm |
| | | | V : false |
| | | | O : Déclenche les perturbations Pod |
| | | | M : kube_pod_status_phase |
| | .RestartPodsEverySeconds (int32) | | F : Redémarre périodiquement les Pods |
| | | | V : 0 |
| | | | O : Plus faible = plus agressif |
| | | | M : kube_pod_container_status_restarts_total |
| | .DeletePodsRandomly (bool) | | F : Supprime aléatoirement des Pods |
| | | | V : false |
| | | | O : Génère beaucoup de recréations Kubernetes |
| | | | M : kube_pod_status_phase |
| | .CrashLoopSimulation (bool) | | F : Simule des Pods en CrashLoopBackOff |
| | | | V : false |
| | | | O : Génère de nombreux redémarrages et events kubelet |
| | | | M : kube_pod_container_status_restarts_total |

### Légende

```text
F : Fonction
V : Valeur par défaut
O : Observation / comportement produit
M : Métriques Prometheus concernées


# Exemple

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest

metadata:
  name: controlplane-full-stress
  namespace: operator-system

spec:

  # ============================================================
  # LEGACY SIMPLE MODE
  # ============================================================

  image: nginx
  replicas: 20

  # ============================================================
  # SCHEDULER STRESS
  # ============================================================

  schedulerStress:
    enabled: true
    nodeCount: 4
    deploymentCount: 20
    replicasPerDeployment: 10

    nodeSelector:
      kubernetes.io/os: linux

    topologySpread: true

    affinity: "soft"
    antiAffinity: "soft"

  # ============================================================
  # API SERVER STRESS
  # ============================================================

  apiServerStress:
    enabled: true
    frequentStatusUpdates: true
    aggressiveReconcile: true
    recreateResources: true

    qps: 100
    burst: 200

  # ============================================================
  # ETCD STRESS
  # ============================================================

  etcdStress:
    enabled: true

    configMapCount: 100
    configMapSizeKB: 64

    secretCount: 50
    secretSizeKB: 32

  # ============================================================
  # CONTROLLER MANAGER STRESS
  # ============================================================

  controllerManagerStress:
    enabled: true

    deploymentCount: 50
    replicasPerDeployment: 5

    recreateReplicaSets: true
    aggressiveGarbageCollection: true

  # ============================================================
  # OPERATOR STRESS
  # ============================================================

  operatorStress:
    enabled: true

    profile: "aggressive"

    informer:
      watchPods: true
      watchConfigMaps: true
      watchDeployments: true

    reconcile:
      maxConcurrent: 20

      qps: 100
      burst: 200

      baseDelaySeconds: 1
      maxDelaySeconds: 10

  # ============================================================
  # POD LIFECYCLE STORM
  # ============================================================

  podLifecycleStorm:
    enabled: true

    restartPodsEverySeconds: 30

    deletePodsRandomly: true
    crashLoopSimulation: true
    ```

---

# 🔥 Exemples de scénarios

## 📡 API Server Stress

Objectif : générer un grand nombre d’updates API, d’écritures ETCD et de reconciliations.

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest
metadata:
  name: api-flood
spec:
  deploymentCount: 5
  replicas: 10
  apiServerStressEnabled: true
  frequentStatusUpdates: true
  aggressiveReconcile: true
  image: nginx:latest
```

## 🗓️ Scheduler Stress

Objectif : créer une forte pression sur le scheduler Kubernetes via de nombreux Pods et contraintes topology spread.

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest
metadata:
  name: scheduler-hard
spec:
  deploymentCount: 10
  replicas: 5
  schedulerStressEnabled: true
  topologySpreadEnabled: true
  image: nginx:latest
```

## 💾 ETCD Stress

Objectif : produire une forte activité ETCD via créations, suppressions et updates fréquentes.

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest
metadata:
  name: etcd-heavy
spec:
  deploymentCount: 15
  replicas: 8
  apiServerStressEnabled: true
  recreateResources: true
  frequentStatusUpdates: true
  image: nginx:latest
```

## 🔁 Controller Manager Stress

Objectif : multiplier les événements Kubernetes et boucles de réconciliation internes.

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest
metadata:
  name: controller-heavy
spec:
  deploymentCount: 12
  replicas: 4
  recreateResources: true
  podStormEnabled: true
  image: nginx:latest
```

## 🤖 Operator Stress

Objectif : mettre sous pression l’opérateur lui-même via de nombreuses reconciliations et requeues.

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest
metadata:
  name: operator-stress
spec:
  deploymentCount: 20
  replicas: 3
  aggressiveReconcile: true
  frequentStatusUpdates: true
  requeueIntervalSeconds: 1
  image: nginx:latest
```