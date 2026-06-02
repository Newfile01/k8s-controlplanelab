# ✅ Validation de l'étape 1 : Création du cluster Kubernetes

Cette étape a pour objectif de valider la création d'un cluster Kubernetes de laboratoire composé de :

- 1 nœud Control Plane
- 3 nœuds Workers
- Une connectivité réseau fonctionnelle
- Les composants système Kubernetes opérationnels

---

# 🔍 Vérification du profil Minikube

## Liste des profils

```bash
minikube profile list
```

```text
PROFILE            DRIVER  RUNTIME  IP              VERSION  STATUS   NODES
control-plane-lab  docker  docker   192.168.67.2    v1.35.1  OK       4
mnd2               docker  docker   192.168.58.2    v1.35.1  Stopped  4
multi-node         kvm2    docker   192.168.39.199  v1.35.1  Stopped  1
```

## Sélection du profil

```bash
minikube profile control-plane-lab
```

```text
✅ minikube profile was successfully set to control-plane-lab
```

---

# 🖥️ Vérification de l'état du cluster

## État Minikube

```bash
minikube status
```

```text
control-plane-lab
├── Control Plane : Running
├── kubelet       : Running
├── apiserver     : Running
└── kubeconfig    : Configured

control-plane-lab-m02
└── Worker : Running

control-plane-lab-m03
└── Worker : Running

control-plane-lab-m04
└── Worker : Running
```

### Validation

✅ Tous les nœuds sont démarrés

---

# ⚙️ Vérification du contexte Kubernetes

```bash
kubectl config current-context
```

```text
control-plane-lab
```

### Validation

✅ Le contexte actif correspond au cluster de laboratoire

---

# 🖧 Vérification des nœuds

```bash
kubectl get nodes -o wide
```

```text
NAME                    STATUS   ROLES           VERSION
control-plane-lab       Ready    control-plane   v1.35.1
control-plane-lab-m02   Ready    worker          v1.35.1
control-plane-lab-m03   Ready    worker          v1.35.1
control-plane-lab-m04   Ready    worker          v1.35.1
```

### Validation

✅ 1 Control Plane présent

✅ 3 Workers présents

✅ Tous les nœuds sont à l'état `Ready`

---

# 🚀 Vérification des composants système

```bash
kubectl get pods -A
```

### Composants observés

| Composant | État |
|------------|-------|
| API Server | ✅ Running |
| Controller Manager | ✅ Running |
| Scheduler | ✅ Running |
| etcd | ✅ Running |
| CoreDNS | ✅ Running |
| kube-proxy | ✅ Running |
| storage-provisioner | ✅ Running |
| kindnet | ✅ Running |

### Validation

Le cluster est sain et tous les composants essentiels du Control Plane sont présents.

---

# 🏗️ Architecture observée

```text
┌─────────────────────────┐
│      Control Plane      │
├─────────────────────────┤
│ kube-apiserver          │
│ kube-controller-manager │
│ kube-scheduler          │
│ etcd                    │
└─────────────────────────┘

┌─────────────────────────┐
│         Workers         │
├─────────────────────────┤
│ kubelet                 │
│ kube-proxy              │
│ Pods applicatifs        │
└─────────────────────────┘
```

---

# 🌐 Réseau

## Plugin CNI utilisé

Observation des pods réseau :

```bash
kubectl get pods -n kube-system | grep kindnet
```

```text
kindnet-xxxx   1/1 Running
kindnet-xxxx   1/1 Running
kindnet-xxxx   1/1 Running
kindnet-xxxx   1/1 Running
```

### Conclusion

Minikube utilise **Kindnet** comme solution CNI (*Container Network Interface*).

Le réseau assure :

- L'attribution des adresses IP aux Pods.
- La communication Pod ↔ Pod.
- La communication inter-nœuds.

```text
Pod
 ↓
Adresse IP Pod
 ↓
Communication réseau
 ↓
Autres Pods / Autres Nœuds
```

---

# 💾 Stockage

Présence du composant :

```text
storage-provisioner
```

Ce composant assure le provisionnement automatique du stockage Kubernetes.

Architecture simplifiée :

```text
PVC
 ↓
PV
 ↓
Volume physique
```

Il sera notamment utilisé lors du déploiement futur de :

- Prometheus
- Grafana

---

# ✅ Résultat de l'étape

L'environnement de laboratoire est opérationnel :

- Cluster Kubernetes fonctionnel
- 1 Control Plane
- 3 Workers
- Réseau CNI opérationnel
- Stockage dynamique disponible
- Composants système Kubernetes en état de fonctionnement

Le cluster est prêt pour l'étape suivante : **déploiement de la pile d'observabilité et étude du Control Plane Kubernetes**.