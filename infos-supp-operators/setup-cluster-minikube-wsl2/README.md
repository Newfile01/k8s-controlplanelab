# Réglages cluster & prise en compte des limites Minikube + WSL2

## 🎯 Présentation

Cette étape vise à :
- ajuster les ressources Minikube,
- comprendre le comportement du driver Docker sous WSL2,
- adapter le cluster à la machine physique,
- éviter la surallocation CPU/RAM.

---

## 🏗️ Architecture & emplacement de travail

### Environnement utilisé

```text
Windows
→ WSL2 Ubuntu
→ Docker Desktop
→ Minikube Driver Docker
→ Kubernetes
```

### Cluster

```text
4 Nodes
1 Control Plane
3 Workers
```

---

## ⚙️ Actions effectuées

### Création du cluster

```bash
minikube start \
--profile=control-plane-lab \
--driver=docker \
--container-runtime=containerd \
--kubernetes-version=v1.35.1 \
--nodes=4
```

### Ajustement CPU/RAM

Paramètres étudiés :

```bash
--cpus
--memory
```

Constat :
- valeurs appliquées par node,
- non globales.

### Vérification ressources nodes

```bash
kubectl get nodes \
-o custom-columns=NAME:.metadata.name,CPU:.status.capacity.cpu,MEM:.status.capacity.memory
```

### Ajustement réaliste ressources

Configuration retenue :

```text
2 vCPU / node
4GB RAM / node
```

---

## 🔎 Vérifications

### Vérifier statut cluster

```bash
minikube status -p control-plane-lab
```

### Vérifier consommation

```bash
kubectl top nodes
```

### Vérifier capacité cluster

```bash
kubectl describe node
```

---

## ✅ Bilan

Le cluster Minikube est désormais :
- cohérent avec la machine physique,
- calibré pour les stress tests,
- adapté à WSL2 + Docker Driver,
- suffisamment contraint pour rendre les métriques control-plane visibles.