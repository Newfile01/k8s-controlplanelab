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

| Champ | Description |
|---|---|
| `replicas` | Nombre de Pods par Deployment |
| `deploymentCount` | Nombre de Deployments générés |
| `image` | Image utilisée par les Pods |
| `schedulerStressEnabled` | Active le stress scheduler |
| `topologySpreadEnabled` | Active les contraintes topology spread |
| `affinityEnabled` | Active les règles d’affinité |
| `antiAffinityEnabled` | Active les règles d’anti-affinité |
| `apiServerStressEnabled` | Active le stress API Server |
| `frequentStatusUpdates` | Force des updates Status fréquents |
| `aggressiveReconcile` | Force des reconciliations agressives |
| `recreateResources` | Recrée périodiquement les ressources |
| `podStormEnabled` | Active les suppressions/recréations de Pods |
| `deletePodsRandomly` | Supprime aléatoirement des Pods |
| `configMapSpamEnabled` | Génère des updates ConfigMap fréquentes |
| `serviceSpamEnabled` | Génère des updates Service fréquentes |
| `stressDurationSeconds` | Durée du scénario |
| `requeueIntervalSeconds` | Intervalle de requeue forcé |

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