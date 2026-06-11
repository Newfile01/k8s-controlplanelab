# REPERTOIRE POUR UN CONTROLLER DE POD (UNIQUEMENT)

Ce répertoire marque le premier controller Kubernetes de ce repository réellement fonctionnel et contient :

* Définition de la CRD : `controlplanetest_types.go`
* Définition de la logique de réconciliation : `controlplanetest_controller.go`

## ✔️ Fonctions en l'état

Ces fichiers permettent de définir des ressources personnalisées `ControlPlaneTest` capables de :

* créer automatiquement un Pod ;
* recréer le Pod si il disparaît ;
* détecter une dérive entre l'état désiré (`spec.image`) et l'état réel du Pod ;
* supprimer/recréer le Pod afin de reconverger vers l'état désiré ;
* maintenir un `Status` synchronisé avec l'état réel du cluster.

## 🔎 Visualisation possibles

On y observe notamment :

* Comment récupérer une ressource depuis l'API Kubernetes avec `r.Get()`
* Comment effectuer des requêtes Kubernetes depuis un controller (`GET`, `POST`, `DELETE`, `UPDATE`)
* Comment vérifier l'existence d'une ressource Kubernetes
* Comment gérer la création automatique d'une ressource
* Comment gérer la suppression/recréation d'une ressource lors d'une dérive
* Comment gérer une relation propriétaire `CR -> Pod` avec `OwnerReferences`
* Comment mettre à jour le `Status` d'une CR
* Comment fonctionne la boucle de réconciliation Kubernetes
* Comment une modification du `Status` génère automatiquement une nouvelle réconciliation
* Comment éviter les boucles infinies de réconciliation grâce à la comparaison d'état

## 📦 Comment l'utiliser ?

1. Déplacer les fichiers dans le dossier `operator/` en remplaçant ceux existants
   *(attention : ne pas conserver deux définitions identiques de CRD/controller dans le même projet Operator)*

2. Régénérer les manifests depuis le dossier `operator/` :

```bash
make generate
make manifests
make install
make run
```

## ⚠️ Notes importantes

* Le cluster Kubernetes doit être démarré et accessible via le `kubeconfig`
* `make install` applique les CRDs au cluster courant
* `make run` lance l'opérateur localement et occupe le terminal
* Prévoir un second terminal pour manipuler le cluster (`kubectl`, tests, logs...)

## 🔄 Fonctionnement général obtenu

```text
Custom Resource
        ↓
Reconcile()
        ↓
Lecture état réel cluster
        ↓
Comparaison état désiré / état réel
        ↓
Création / suppression / recréation Pod
        ↓
Mise à jour Status
        ↓
Nouvelle reconciliation si nécessaire
```
