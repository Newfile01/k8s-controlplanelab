# 02 - REPERTOIRE POUR UN CONTROLLER DE DEPLOYMENT (GESTION DES PODS SOUS-TRAITEE À KUBERNETES)

Ce répertoire regroupe les fichiers `_types.go` et `_controller.go` permettant d'implémenter une logique d'Operator Kubernetes basée sur les mécanismes natifs de Kubernetes :

- rolling updates ;
- self-healing ;
- reconciliation ;
- convergence d'état.

L'Operator ne gère plus directement les Pods mais pilote désormais un `Deployment` Kubernetes.

Architecture obtenue :

```text
Custom Resource
        ↓
Deployment
        ↓
[Suite gérée par K8S]
ReplicaSet
        ↓
Pods
```

## 🎯 Objectifs atteints

Ces fichiers permettent de définir des ressources personnalisées `ControlPlaneTest` capables de :

- créer automatiquement un Deployment ;
- superviser la convergence entre état désiré et état réel ;
- mettre à jour automatiquement l'image et le nombre de replicas ;
- exploiter les rolling updates natifs Kubernetes ;
- exploiter le self-healing natif des Deployments ;
- maintenir un Status synchronisé avec l'état réel du cluster ;
- gérer automatiquement la dépendance `CR -> Deployment` avec `OwnerReferences`.

## 🔎 Ce que ce controller permet d'observer

### 🔄 Boucle de réconciliation Kubernetes

```text
CR créée/modifiée
        ↓
Reconcile()
        ↓
Lecture état réel
        ↓
Comparaison desired/current state
        ↓
CREATE / UPDATE si dérive
        ↓
Nouvelle reconciliation si nécessaire
```

### 📡 Requêtes Kubernetes depuis Go

Le controller effectue directement des opérations Kubernetes :

| Fonction Go | Requête Kubernetes |
|---|---|
| `r.Get()` | GET |
| `r.Create()` | POST |
| `r.Update()` | PUT/PATCH |
| `r.Status().Update()` | UPDATE `/status` |

### 🧠 Convergence déclarative

L'Operator ne manipule plus directement les Pods.

Il pilote uniquement :

```text
CR ↔ Deployment
```

Puis Kubernetes gère automatiquement :

- ReplicaSets ;
- Pods ;
- scheduling ;
- self-healing ;
- rolling updates.

### 🔥 Fonctionnalités observables

- Création automatique du Deployment
- Rolling update lors du changement d'image
- Scale up/down via `replicas`
- Self-healing lors de suppression manuelle de Pods
- Mise à jour automatique du Status
- Boucles de réconciliation déclenchées par les UPDATE Kubernetes

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

- Le cluster Kubernetes doit être démarré et accessible via le `kubeconfig`
- `make install` applique les CRDs au cluster courant
- `make run` lance l'opérateur localement et occupe le terminal
- Prévoir un second terminal pour manipuler le cluster (`kubectl`, tests, logs...)
- Chaque modification du `_types.go` nécessite une régénération des manifests

## 🧪 Exemple de Custom Resource

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest

metadata:
  name: nginx-test

spec:
  image: nginx:latest
  replicas: 3
```

## 🔍 Status observé

```yaml
status:
  deploymentName: nginx-test-deployment
  readyReplicas: 3
  availableReplicas: 3
```

## Tester les rolling updates & rollbacks

```bash
kubectl set image deployment/nginx-test-deployment nginx=nginx:stable-alpine3.23-perl
# deployment.apps/nginx-test-deployment image updated

kubectl set image deployment/nginx-test-deployment nginx=nginx:1.31.1-alpine3.23-perl
# deployment.apps/nginx-test-deployment image updated

kubectl rollout status deployment/nginx-test-deployment
# deployment "nginx-test-deployment" successfully rolled out

kubectl rollout history deployment/nginx-test-deployment
# deployment.apps/nginx-test-deployment
# REVISION  CHANGE-CAUSE
# 1         <none>
# 2         <none>
# 3         <none>
# 4         <none>
# 5         <none>

kubectl rollout undo deployment/nginx-test-deployment
# deployment.apps/nginx-test-deployment rolled back
```
