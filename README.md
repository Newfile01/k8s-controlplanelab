# ☸️ Kubernetes Control Plane Stress Operator

Opérateur Kubernetes développé avec Kubebuilder permettant de générer des scénarios de stress ciblant les composants du control-plane Kubernetes afin d'observer leur comportement, leurs limites et leurs métriques sous charge.

Composants pouvant être sollicités :

* API Server
* Scheduler
* ETCD
* Controller Manager
* Controller Runtime Operator
* Nodes Kubernetes

---

# 🚀 Installation via Helm OCI

## 🔐 Authentification GHCR

Créer un token GitHub avec permissions : `read:packages`

```bash
# Connexion au registry OCI qui stocke l'artefact final
export CR_PAT=<github_token>

echo $CR_PAT | helm registry login ghcr.io \
-u <github_user> \
--password-stdin
```

## 📦 Installation standard

```bash
helm install controlplane-operator \
oci://ghcr.io/newfile01/charts/controlplane-operator \
-n operator-system \
--create-namespace
```

## ⚡ Installation avec scénarios de stress
Ajouter ensuite les options :
* Scheduler stress : `--set stressTests.scheduler.enabled=true`
* API Server stress : `--set stressTests.apiServer.enabled=true`
* ETCD stress : `--set stressTests.etcd.enabled=true`
* Plusieurs scénarios simultanément : les ajouter tous (avec des espaces)

Ex.

```bash
helm install controlplane-operator \
oci://ghcr.io/newfile01/charts/controlplane-operator \
-n operator-system \
--create-namespace \
--set stressTests.scheduler.enabled=true \
--set stressTests.apiServer.enabled=true \
--set stressTests.etcd.enabled=true
```

---

# 🔎 Vérification rapide

```bash
# Vérifier le release Helm
helm list -A
# Vérifier les ressources opérateur
kubectl get all -n operator-system
# Vérifier les CRDs
kubectl get crd | grep controlplane
# Vérifier les scénarios générés
kubectl get controlplanetest -A
# Vérifier les logs opérateur
kubectl logs -n operator-system \
deployment/operator-controller-manager -f
# Vérifier le ServiceMonitor
kubectl get servicemonitor -A
# Vérifier les métriques exposées
kubectl port-forward \
svc/operator-controller-manager-metrics-service \
8443:8443 \
-n operator-system
# Endpoint métriques
# https://localhost:8443/metrics
```

---

# 🧪 Création d'une Custom Resource `ControlPlaneTest`

L’opérateur fonctionne à partir de ressources Kubernetes personnalisées (`Custom Resources`) de type :

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest
```

Chaque ressource permet d’activer un ou plusieurs scénarios de stress ciblant différents composants du control-plane Kubernetes.

---

## 📄 Squelette global d'une CR

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest

metadata:
  name: my-controlplane-test
  namespace: operator-system

spec:
  # options générales
  image: nginx

  schedulerStress:
  # Vos options avec 'enabled' passé à 'true'
    enabled: false

  apiServerStress:
  # Vos options avec 'enabled' passé à 'true'
    enabled: false

  etcdStress:
  # Vos options avec 'enabled' passé à 'true'
    enabled: false

  controllerManagerStress:
  # Vos options avec 'enabled' passé à 'true'
    enabled: false

  operatorStress:
  # Vos options avec 'enabled' passé à 'true'
    enabled: false

  podLifecycleStorm:
  # Vos options avec 'enabled' passé à 'true'
    enabled: false
```

---

## ⚙️ Déploiement de la CR

```bash
# Création du fichier avec Vi ou Nano
vi my-scenario.yaml
# Puis appliquer
kubectl apply -f my-scenario.yaml
```

---

## 🔎 Vérifications

```bash
# Vérifier la CR
kubectl get controlplanetest -A
# Détailler le scénario
kubectl describe controlplanetest my-controlplane-test \
-n operator-system
# Observer les ressources générées
kubectl get deployments,pods,svc,configmaps \
-n operator-system
# Logs opérateur
kubectl logs -n operator-system \
deployment/operator-controller-manager -f
```

Les sections suivantes détaillent l’ensemble des champs disponibles pour chaque scénario de stress.

---

# 🧪 Configuration de la CRD

### Légende

```text
F : Fonction
V : Valeur par défaut
O : Observation / comportement produit
M : Métriques Prometheus à regarder
```

## 📝 Liste des clé/valeurs possible dans une CR

| Champs | Description |
|---|---|
| `SchedulerStressSpec` | F : Génère une forte activité du scheduler Kubernetes via la création massive de Deployments et Pods répartis sur plusieurs noeuds<br>V : {}<br>O : Impact visible sur scheduler, etcd, controller-manager et kubelet<br>M : `scheduler_pending_pods` \| `scheduler_e2e_scheduling_duration_seconds` \| `process_cpu_seconds_total` |
| `SchedulerStressSpec.Enabled` | F : Active ou désactive le scénario de stress scheduler<br>V : false<br>O : Déclenche la création des ressources de stress scheduler<br>M : `scheduler_schedule_attempts_total` |
| `SchedulerStressSpec.NodeCount` | F : Nombre de noeuds ciblés pour la répartition des Pods<br>V : 0<br>O : Influence la dispersion des Pods dans le cluster<br>M : `kube_node_status_capacity` \| `scheduler_pending_pods` |
| `SchedulerStressSpec.DeploymentCount` | F : Nombre de Deployments créés simultanément<br>V : 0<br>O : Augmente fortement les objets surveillés par Kubernetes<br>M : `apiserver_request_total` \| `workqueue_depth` |
| `SchedulerStressSpec.ReplicasPerDeployment` | F : Nombre de Pods créés par Deployment<br>V : 0<br>O : Augmente directement la charge du scheduler et du kubelet<br>M : `scheduler_pod_scheduling_attempts` \| `kube_pod_status_phase` |
| `SchedulerStressSpec.NodeSelector` | F : Contraintes de placement imposées aux Pods via labels Kubernetes<br>V : {}<br>O : Peut provoquer des Pods Pending si aucun noeud compatible<br>M : `scheduler_pending_pods` |
| `SchedulerStressSpec.TopologySpread` | F : Active les topologySpreadConstraints pour équilibrer les Pods<br>V : false<br>O : Complexifie les décisions du scheduler<br>M : `scheduler_framework_extension_point_duration_seconds` |
| `SchedulerStressSpec.Affinity` | F : Définit des règles d'affinité de placement des Pods<br>V : ""<br>O : Augmente les calculs de placement du scheduler<br>M : `scheduler_scheduling_algorithm_duration_seconds` |
| `SchedulerStressSpec.AntiAffinity` | F : Définit des règles anti-affinité empêchant certains placements<br>V : ""<br>O : Peut fortement ralentir le scheduling sur clusters chargés<br>M : `scheduler_pending_pods` \| `scheduler_e2e_scheduling_duration_seconds` |
| `APIServerStressSpec` | F : Génère une forte activité sur l'API Kubernetes via requêtes et reconciliations agressives<br>V : {}<br>O : Impact direct sur API Server et ETCD<br>M : `apiserver_request_total` \| `apiserver_current_inflight_requests` |
| `APIServerStressSpec.Enabled` | F : Active le stress API Server<br>V : false<br>O : Déclenche les scénarios API intensifs<br>M : `apiserver_request_total` |
| `APIServerStressSpec.FrequentStatusUpdates` | F : Effectue des mises à jour fréquentes du champ status des CRs<br>V : false<br>O : Génère beaucoup d'écritures ETCD<br>M : `etcd_request_duration_seconds` |
| `APIServerStressSpec.AggressiveReconcile` | F : Force des reconciliations très fréquentes<br>V : false<br>O : Augmente fortement les appels API Kubernetes<br>M : `controller_runtime_reconcile_total` |
| `APIServerStressSpec.RecreateResources` | F : Supprime et recrée régulièrement les ressources Kubernetes<br>V : false<br>O : Génère beaucoup d'events et d'opérations ETCD<br>M : `apiserver_request_total` \| `etcd_disk_backend_commit_duration_seconds` |
| `APIServerStressSpec.QPS` | F : Limite QPS utilisée par le client Kubernetes<br>V : 0<br>O : Plus élevé = plus de requêtes API par seconde<br>M : `apiserver_request_total` |
| `APIServerStressSpec.Burst` | F : Nombre maximal de requêtes burst avant throttling<br>V : 0<br>O : Génère des pics brutaux de trafic API<br>M : `apiserver_current_inflight_requests` |
| `EtcdStressSpec` | F : Génère de nombreuses écritures ETCD via ConfigMaps et Secrets<br>V : {}<br>O : Charge fortement le backend disque ETCD<br>M : `etcd_disk_backend_commit_duration_seconds` \| `etcd_mvcc_db_total_size_in_bytes` |
| `EtcdStressSpec.Enabled` | F : Active le stress ETCD<br>V : false<br>O : Déclenche les écritures massives dans ETCD<br>M : `etcd_server_proposals_applied_total` |
| `EtcdStressSpec.ConfigMapCount` | F : Nombre de ConfigMaps créées<br>V : 0<br>O : Augmente fortement le nombre d'objets stockés dans ETCD<br>M : `etcd_mvcc_db_total_size_in_bytes` |
| `EtcdStressSpec.ConfigMapSizeKB` | F : Taille des ConfigMaps générées<br>V : 0<br>O : Impacte directement les écritures disque ETCD<br>M : `etcd_disk_backend_commit_duration_seconds` |
| `EtcdStressSpec.SecretCount` | F : Nombre de Secrets générés<br>V : 0<br>O : Charge supplémentaire sur sérialisation et stockage ETCD<br>M : `etcd_server_proposals_applied_total` |
| `EtcdStressSpec.SecretSizeKB` | F : Taille des Secrets générés<br>V : 0<br>O : Augmente taille mémoire et disque de la base ETCD<br>M : `etcd_mvcc_db_total_size_in_bytes` |
| `ControllerManagerStressSpec` | F : Génère une activité importante du controller-manager Kubernetes<br>V : {}<br>O : Charge ReplicaSets, Deployments et garbage collector Kubernetes<br>M : `workqueue_depth` \| `workqueue_adds_total` |
| `ControllerManagerStressSpec.Enabled` | F : Active le scénario controller-manager<br>V : false<br>O : Déclenche les contrôleurs Kubernetes associés<br>M : `workqueue_queue_duration_seconds` |
| `ControllerManagerStressSpec.DeploymentCount` | F : Nombre de Deployments générés<br>V : 0<br>O : Augmente les opérations de convergence Kubernetes<br>M : `deployment_controller_requeues_total` |
| `ControllerManagerStressSpec.ReplicasPerDeployment` | F : Nombre de réplicas par Deployment<br>V : 0<br>O : Augmente Pods et ReplicaSets gérés<br>M : `replicaset_controller_requeues_total` |
| `ControllerManagerStressSpec.RecreateReplicaSets` | F : Supprime et recrée fréquemment les ReplicaSets<br>V : false<br>O : Génère beaucoup d'events Kubernetes<br>M : `workqueue_adds_total` |
| `ControllerManagerStressSpec.AggressiveGarbageCollection` | F : Force un nettoyage fréquent des ressources supprimées<br>V : false<br>O : Impact visible sur controller-manager et ETCD<br>M : `garbagecollector_controller_resources_sync_error_total` |
| `OperatorStressSpec` | F : Simule une forte activité interne de l'opérateur Kubernetes<br>V : {}<br>O : Charge controller-runtime, API Server et ETCD<br>M : `controller_runtime_reconcile_total` \| `controller_runtime_active_workers` |
| `OperatorStressSpec.Enabled` | F : Active le stress de l'opérateur<br>V : false<br>O : Déclenche les comportements agressifs de reconciliation<br>M : `controller_runtime_reconcile_total` |
| `OperatorStressSpec.Profile` | F : Définit un profil de stress prédéfini<br>V : ""<br>O : Permet de charger plusieurs paramètres automatiquement<br>M : `depends_on_profile` |
| `OperatorStressSpec.Informer` | F : Configure les watchers Kubernetes utilisés par l'opérateur<br>V : {}<br>O : Plus les watchers sont nombreux, plus les événements Kubernetes sont consommés<br>M : `controller_runtime_reconcile_total` |
| `OperatorStressSpec.Informer.WatchPods` | F : Active la surveillance des Pods<br>V : false<br>O : Génère beaucoup d'événements sur gros clusters<br>M : `controller_runtime_reconcile_total` |
| `OperatorStressSpec.Informer.WatchConfigMaps` | F : Active la surveillance des ConfigMaps<br>V : false<br>O : Très bavard avec ETCD stress<br>M : `controller_runtime_reconcile_total` |
| `OperatorStressSpec.Informer.WatchDeployments` | F : Active la surveillance des Deployments<br>V : false<br>O : Augmente fortement les reconciliations<br>M : `controller_runtime_reconcile_total` |
| `OperatorStressSpec.Reconcile` | F : Configure le comportement de reconciliation controller-runtime<br>V : {}<br>O : Impact direct sur CPU opérateur et trafic API<br>M : `controller_runtime_reconcile_time_seconds` |
| `OperatorStressSpec.Reconcile.MaxConcurrent` | F : Nombre maximal de reconciliations simultanées<br>V : 0<br>O : Augmente le parallélisme du controller<br>M : `controller_runtime_active_workers` |
| `OperatorStressSpec.Reconcile.QPS` | F : Limite QPS du client Kubernetes<br>V : 0<br>O : Plus élevé = plus de requêtes API simultanées<br>M : `apiserver_request_total` |
| `OperatorStressSpec.Reconcile.Burst` | F : Nombre maximal de requêtes burst<br>V : 0<br>O : Génère des pics de trafic API<br>M : `apiserver_current_inflight_requests` |
| `OperatorStressSpec.Reconcile.BaseDelaySeconds` | F : Délai minimal entre deux reconciliations<br>V : 0<br>O : Plus faible = comportement plus agressif<br>M : `controller_runtime_reconcile_time_seconds` |
| `OperatorStressSpec.Reconcile.MaxDelaySeconds` | F : Délai maximal du backoff de reconciliation<br>V : 0<br>O : Contrôle les retries après erreur<br>M : `controller_runtime_reconcile_errors_total` |
| `PodLifecycleStormSpec` | F : Génère des perturbations constantes du cycle de vie des Pods<br>V : {}<br>O : Charge fortement scheduler, kubelet et controller-manager<br>M : `kube_pod_container_status_restarts_total` \| `scheduler_pending_pods` |
| `PodLifecycleStormSpec.Enabled` | F : Active le scénario Pod lifecycle storm<br>V : false<br>O : Déclenche les perturbations Pod<br>M : `kube_pod_status_phase` |
| `PodLifecycleStormSpec.RestartPodsEverySeconds` | F : Redémarre périodiquement les Pods<br>V : 0<br>O : Plus faible = plus agressif<br>M : `kube_pod_container_status_restarts_total` |
| `PodLifecycleStormSpec.DeletePodsRandomly` | F : Supprime aléatoirement des Pods<br>V : false<br>O : Génère beaucoup de recréations Kubernetes<br>M : `kube_pod_status_phase` |
| `PodLifecycleStormSpec.CrashLoopSimulation` | F : Simule des Pods en CrashLoopBackOff<br>V : false<br>O : Génère de nombreux redémarrages et events kubelet<br>M : `kube_pod_container_status_restarts_total` |

---

# 🔥 Exemples de scénarios

## 🖼️ Exemple CR complète

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


## 📡 API Server Stress

Objectif : générer un grand nombre de requêtes API Kubernetes, de mises à jour ETCD et de reconciliations agressives.

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest

metadata:
  name: api-flood
  namespace: operator-system

spec:
  image: nginx:latest
  replicas: 20

  apiServerStress:
    enabled: true
    frequentStatusUpdates: true
    aggressiveReconcile: true
    recreateResources: true
    qps: 100
    burst: 250

  schedulerStress:
    enabled: true
    deploymentCount: 10
    replicasPerDeployment: 10
```

Effets observables :

* hausse du trafic API Server
* augmentation des écritures ETCD
* augmentation des reconciliations operator/controller-runtime
* saturation progressive des workers API

Métriques intéressantes :

* `apiserver_request_total`
* `apiserver_current_inflight_requests`
* `controller_runtime_reconcile_total`
* `etcd_disk_backend_commit_duration_seconds`

---

## 🗓️ Scheduler Stress

Objectif : créer une forte pression sur le scheduler Kubernetes via de nombreux Pods et contraintes de placement complexes.

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest

metadata:
  name: scheduler-hard
  namespace: operator-system

spec:
  image: nginx:latest
  replicas: 50

  schedulerStress:
    enabled: true
    deploymentCount: 25
    replicasPerDeployment: 20
    topologySpread: true
    affinity: "zone-spread"
    antiAffinity: "strict"

  podLifecycleStorm:
    enabled: true
    restartPodsEverySeconds: 15
```

Effets observables :

* forte augmentation des Pods Pending
* hausse CPU scheduler
* scheduling plus lent
* multiplication des décisions de placement

Métriques intéressantes :

* `scheduler_pending_pods`
* `scheduler_e2e_scheduling_duration_seconds`
* `scheduler_framework_extension_point_duration_seconds`
* `kube_pod_status_phase`

---

## 💾 ETCD Stress

Objectif : produire une forte activité disque et mémoire ETCD via créations massives de ConfigMaps et Secrets.

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest

metadata:
  name: etcd-heavy
  namespace: operator-system

spec:
  image: nginx:latest
  replicas: 10

  etcdStress:
    enabled: true
    configMapCount: 500
    configMapSizeKB: 256
    secretCount: 300
    secretSizeKB: 128

  apiServerStress:
    enabled: true
    frequentStatusUpdates: true
    recreateResources: true
    qps: 80
    burst: 150
```

Effets observables :

* augmentation importante de la taille ETCD
* hausse des commits disque
* latences ETCD visibles
* ralentissement API Kubernetes

Métriques intéressantes :

* `etcd_mvcc_db_total_size_in_bytes`
* `etcd_disk_backend_commit_duration_seconds`
* `etcd_server_proposals_applied_total`
* `apiserver_request_duration_seconds`

---

## 🔁 Controller Manager Stress

Objectif : multiplier les événements Kubernetes et les opérations de convergence des contrôleurs internes.

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest

metadata:
  name: controller-heavy
  namespace: operator-system

spec:
  image: nginx:latest
  replicas: 30

  controllerManagerStress:
    enabled: true
    deploymentCount: 20
    replicasPerDeployment: 10
    recreateReplicaSets: true
    aggressiveGarbageCollection: true

  podLifecycleStorm:
    enabled: true
    deletePodsRandomly: true
    restartPodsEverySeconds: 10
```

Effets observables :

* forte activité ReplicaSet controller
* nombreuses recréations de Pods
* hausse workqueues Kubernetes
* activité garbage collector élevée

Métriques intéressantes :

* `workqueue_depth`
* `workqueue_adds_total`
* `replicaset_controller_requeues_total`
* `deployment_controller_requeues_total`

---

## 🤖 Operator Stress

Objectif : mettre sous pression l’opérateur Kubernetes lui-même via requeues, watchers et reconciliations massives.

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest

metadata:
  name: operator-stress
  namespace: operator-system

spec:
  image: nginx:latest
  replicas: 40

  operatorStress:
    enabled: true
    profile: aggressive

    informer:
      watchPods: true
      watchConfigMaps: true
      watchDeployments: true

    reconcile:
      maxConcurrent: 20
      qps: 150
      burst: 300
      baseDelaySeconds: 1
      maxDelaySeconds: 3

  apiServerStress:
    enabled: true
    aggressiveReconcile: true
    frequentStatusUpdates: true
```

Effets observables :

* saturation controller-runtime
* augmentation du nombre de reconciliations
* hausse CPU opérateur
* très forte activité API Kubernetes

Métriques intéressantes :

* `controller_runtime_reconcile_total`
* `controller_runtime_active_workers`
* `controller_runtime_reconcile_time_seconds`
* `controller_runtime_reconcile_errors_total`

---

## 💥 Pod Lifecycle Storm

Objectif : provoquer des perturbations constantes du cycle de vie des Pods Kubernetes.

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest

metadata:
  name: pod-chaos
  namespace: operator-system

spec:
  image: nginx:latest
  replicas: 60

  podLifecycleStorm:
    enabled: true
    restartPodsEverySeconds: 5
    deletePodsRandomly: true
    crashLoopSimulation: true

  schedulerStress:
    enabled: true
    deploymentCount: 15
    replicasPerDeployment: 15
```

Effets observables :

* nombreux redémarrages Pods
* forte activité kubelet
* recréations permanentes
* augmentation events Kubernetes

Métriques intéressantes :

* `kube_pod_container_status_restarts_total`
* `kube_pod_status_phase`
* `scheduler_pending_pods`
* `kubelet_runtime_operations_total`
