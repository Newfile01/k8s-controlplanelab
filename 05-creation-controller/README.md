# 🔄 Comprendre les appels Kubernetes dans un controller Kubebuilder

## 🧠 Principe général

Un controller Kubernetes ne dialogue jamais directement avec etcd.

Toutes les opérations passent par :

```text
Controller Go
        ↓
controller-runtime client
        ↓
REST API Kubernetes
        ↓
API Server
        ↓
etcd
```

Kubebuilder et `controller-runtime` fournissent un client Kubernetes Go simplifié permettant d'effectuer des opérations sur les ressources du cluster.

---

## 📦 Le rôle du Reconciler

Dans :

```go
func (r *ControlPlaneTestReconciler) Reconcile(...)
```

`r` représente le Reconciler et contient notamment :

```go
client.Client
```

fourni par :

```text
sigs.k8s.io/controller-runtime
```

---

## 🔍 Requête GET Kubernetes

Code Go :

```go
err := r.Get(ctx, req.NamespacedName, controlPlaneTest)
```

Interprétation :

```text
Récupère une ressource Kubernetes
depuis l'API Server
```

Équivalent REST :

```http
GET /apis/controlplane.lab.local/v1alpha1/namespaces/default/controlplanetests/controlplanetest-sample
```

---

## 📄 Paramètres du `Get()`

| Élément | Rôle |
|---|---|
| `ctx` | contexte d'exécution (timeout, logs, annulation...) |
| `req.NamespacedName` | nom + namespace de la ressource recherchée |
| `controlPlaneTest` | structure Go recevant la réponse de l'API |

Après le `Get()` :

```go
controlPlaneTest.Name
controlPlaneTest.Namespace
controlPlaneTest.Spec.Image
```

sont automatiquement remplis.

---

## 🚀 Requête CREATE Kubernetes

Code Go :

```go
err = r.Create(ctx, pod)
```

Interprétation :

```text
Créer une ressource Kubernetes
via l'API Server
```

Équivalent REST :

```http
POST /api/v1/namespaces/default/pods
```

avec le Pod envoyé dans le body :

```json
{
  "kind": "Pod",
  ...
}
```

---

## 🔄 Correspondance des opérations Kubernetes

| Fonction Go | Requête REST Kubernetes |
|---|---|
| `r.Get()` | GET |
| `r.List()` | LIST |
| `r.Create()` | POST |
| `r.Update()` | PUT / PATCH |
| `r.Delete()` | DELETE |

---

## 🔁 Fonctionnement réel d'un controller

```text
Event Kubernetes
        ↓
Reconcile()
        ↓
Lecture état réel via GET
        ↓
Comparaison état désiré / état réel
        ↓
Création / modification / suppression
        ↓
Nouvelle réconciliation
```

---

## 🎯 Exemple réel du lab

```text
Custom Resource
        ↓
Reconcile()
        ↓
r.Get() -> lecture CR
        ↓
r.Create() -> création Pod nginx
        ↓
API Server
        ↓
Scheduler
        ↓
Node
```

---

## 🧠 Idée fondamentale

Kubebuilder simplifie l'accès à l'API Kubernetes :

```text
Requête REST Kubernetes
↓
abstraction Go controller-runtime
↓
controller/operator Kubernetes
```

Le controller agit finalement comme :

```text
un client automatisé de l'API Kubernetes
```