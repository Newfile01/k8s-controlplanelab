# 🚀 Étape #1 - Création du cluster Kubernetes

## Objectif

Mettre en place un cluster Kubernetes local servant de laboratoire d'expérimentation pour l'étude du plan de contrôle Kubernetes et le développement d'un opérateur pédagogique basé sur Kubebuilder.

---

## Configuration retenue

### Machine hôte

| Ressource           | Valeur                |
| ------------------- | --------------------- |
| CPU                 | Intel Core i9-14900HX |
| Mémoire             | 32 Go                 |
| Stockage disponible | ~50 Go                |

### Cluster Minikube

```bash
minikube start \
  --profile control-plane-lab \
  --driver=docker \
  --nodes 4 \
  --cpus 2 \
  --memory 4096
```

Configuration obtenue :

* 1 nœud Control Plane
* 3 nœuds Workers
* Runtime Docker
* Kubernetes v1.35.1

---

## Notion de profil Minikube

Un profil représente un cluster Minikube indépendant.

Consultation :

```bash
minikube profile list
```

Activation :

```bash
minikube profile control-plane-lab
```

Vérification :

```bash
minikube status
```

---

## kubeconfig

Le cluster est enregistré dans :

```text
~/.kube/config
```

Contexte actif :

```bash
kubectl config current-context
```

Résultat :

```text
control-plane-lab
```

Le kubeconfig contient :

* les informations du cluster ;
* les certificats ;
* les utilisateurs ;
* les contextes Kubernetes.

---

## Architecture obtenue

```text
┌─────────────────────┐
│   Control Plane     │
├─────────────────────┤
│ API Server          │
│ Scheduler           │
│ Controller Manager  │
│ etcd                │
└─────────────────────┘

        │

 ┌──────┼──────┐
 │      │      │

 ▼      ▼      ▼

Worker Worker Worker
```

---

## Vérifications réalisées

Liste des nœuds :

```bash
kubectl get nodes -o wide
```

Liste des composants système :

```bash
kubectl get pods -A
```

Composants observés :

* kube-apiserver
* kube-controller-manager
* kube-scheduler
* etcd
* coredns
* kube-proxy
* kindnet
* storage-provisioner

---

## Résultat

Le cluster est opérationnel et constitue la base du laboratoire destiné à l'observation du fonctionnement du plan de contrôle Kubernetes.
