# 🔄 Du Pod Controller à l’Opérateur Event-Driven

---

# 🎯 Objectif du lab

Ce répertoire a pour objectif de comprendre progressivement :
- le fonctionnement interne des controllers Kubernetes
- la logique de réconciliation
- l’architecture event-driven de Kubernetes
- controller-runtime
- les Operators Kubernetes modernes
- les mécanismes de convergence du control-plane

L’idée n’est pas simplement de “faire fonctionner” un Operator, mais surtout de comprendre :
- comment Kubernetes pilote réellement les ressources
- comment les événements circulent
- comment un controller interagit avec l’API Server
- comment construire un Operator scalable et observable

# 🧠 Qu’est-ce qu’un controller Kubernetes ?

Un controller Kubernetes est une boucle de contrôle :

```text
Etat désiré (Spec)
        VS
Etat réel (Cluster)
```

Le controller :
1. observe le cluster
2. compare état réel / désiré
3. corrige les différences (drift)
4. recommence continuellement

Le tout via l’API Kubernetes.

# 🔄 Architecture générale d’un Operator

```text
Custom Resource
        ↓
API Server Kubernetes
        ↓
Event Kubernetes
        ↓
controller-runtime
        ↓
Reconcile()
        ↓
GET état réel
        ↓
Comparaison Spec / Status
        ↓
Création / modification / suppression ressources
        ↓
Nouvelle convergence
```

# 🧠 Philosophie de progression du lab

Le lab a été découpé en plusieurs étapes afin de :
- comprendre les bases avant l’abstraction Kubernetes
- visualiser la logique réelle des controllers
- comprendre progressivement les mécanismes natifs Kubernetes
- terminer sur une architecture proche des vrais opérateurs production-grade

---

# 1️⃣ 01-pod-controller

## 🎯 Objectif

Comprendre les bases fondamentales :
- boucle de réconciliation
- GET / CREATE Kubernetes
- gestion du drift
- création manuelle de Pods
- fonctionnement de controller-runtime

## 🧠 Ce que cette étape met en lumière

Le controller devient :
- un client automatisé de l’API Kubernetes
- une boucle infinie de convergence

Cette étape permet surtout de comprendre :
- le rôle du Reconcile()
- les appels API Kubernetes
- la logique “état désiré VS état réel”

## 🔄 Fonctionnement

```text
Custom Resource
        ↓
Reconcile()
        ↓
GET CR
        ↓
GET Pod
        ↓
Pod absent ?
        ↓
CREATE Pod
```

## 📦 Ressources gérées

- Custom Resource
- Pod

## 🔧 Eléments importants ajoutés

### Reconcile()

```go
func (r *ControlPlaneTestReconciler) Reconcile(...)
```

### Lecture CR

```go
r.Get(...)
```

### Création Pod

```go
r.Create(...)
```

### Gestion drift

```text
Pod supprimé
        ↓
Nouvelle réconciliation
        ↓
Pod recréé
```

## 🧠 Limites de cette approche

Créer directement des Pods :
- contourne les mécanismes Kubernetes natifs
- ne scale pas bien
- ne bénéficie pas des ReplicaSets/Deployments
- rend la gestion complexe

➡️ passage au Deployment Controller.

---

# 2️⃣ 02-deployment-controller

## 🎯 Objectif

S’appuyer sur les mécanismes natifs Kubernetes.

## 🧠 Ce que cette étape met en lumière

Kubernetes sait déjà :
- gérer des Pods
- gérer le scaling
- recréer les Pods
- corriger les crashs
- gérer les ReplicaSets

➡️ le rôle du controller devient :

```text
piloter Kubernetes
plutôt que gérer les Pods lui-même
```

## 🔄 Fonctionnement

```text
Custom Resource
        ↓
Operator
        ↓
Deployment
        ↓
ReplicaSet
        ↓
Pods
```

## 📦 Ressources gérées

- Custom Resource
- Deployment

## 🔧 Eléments importants ajoutés

### Création Deployment

```go
appsv1.Deployment{}
```

### PodTemplate

```go
Spec.Template
```

### Replicas dynamiques

```go
Spec.Replicas
```

### Labels / Selectors

```go
MatchLabels
```

## 🎯 Apports majeurs

Grâce au Deployment :
- scaling natif
- auto-healing Kubernetes
- rolling updates
- ReplicaSets
- convergence native Kubernetes

➡️ le controller devient beaucoup plus simple.

## 🧠 Changement fondamental

Avant :

```text
Operator → Pods
```

Après :

```text
Operator → Deployment → Kubernetes gère les Pods
```

---

# 3️⃣ 03-deployment-controller-finalizers

## 🎯 Objectif

Comprendre la suppression contrôlée des ressources Kubernetes.

## 🧠 Ce que cette étape met en lumière

Une suppression Kubernetes n’est pas immédiate.

Kubernetes :
1. marque la ressource en suppression
2. attend les finalizers
3. laisse les controllers nettoyer
4. supprime réellement ensuite

## 🔄 Fonctionnement

```text
kubectl delete CR
        ↓
DeletionTimestamp
        ↓
Finalizer détecté
        ↓
Cleanup ressources
        ↓
Suppression finalizer
        ↓
Suppression réelle CR
```

## 📦 Ressources gérées

- Custom Resource
- Deployment
- Finalizers

## 🔧 Eléments importants ajoutés

### Finalizer

```go
controllerutil.AddFinalizer()
```

### Détection suppression

```go
DeletionTimestamp != nil
```

### Nettoyage ressources

```go
r.Delete(...)
```

### Retrait finalizer

```go
controllerutil.RemoveFinalizer()
```

## 🎯 Ce que cela apporte

Permet :
- cleanup propre
- gestion dépendances
- suppression contrôlée
- logique “pré-delete”

➡️ indispensable en production.

---

# 4️⃣ 04-owns-controller

## 🎯 Objectif

Construire un vrai Operator multi-ressources.

## 🧠 Ce que cette étape met en lumière

Un Operator ne gère pas une seule ressource.

Il pilote :
- Deployments
- Services
- ConfigMaps
- Secrets
- StatefulSets
- etc...

## 🔄 Fonctionnement

```text
Custom Resource
        ↓
Operator
        ├── Deployment
        ├── Service
        └── ConfigMap
```

## 📦 Ressources gérées

- Deployment
- Service
- ConfigMap

## 🔧 Eléments importants ajoutés

### OwnerReferences

```go
ctrl.SetControllerReference(...)
```

### Owns()

```go
Owns(&appsv1.Deployment{})
```

### Status

```go
Status.ReadyReplicas
Status.AvailableReplicas
```

### observedGeneration

```go
Status.ObservedGeneration
```

### Conditions

```go
metav1.Condition
```

## 🎯 Ce que cela apporte

Permet :
- ownership Kubernetes
- garbage collection native
- suivi état réel
- status avancé
- convergence observable

➡️ vraie architecture Operator Kubernetes.

## 🧠 observedGeneration

Permet de savoir :

```text
Spec utilisateur appliquée ?
```

## 🔄 Exemple

```text
metadata.generation = 7
status.observedGeneration = 6
```

➡️ Operator pas encore convergé.

---

# 5️⃣ 05-event-driven-controller

## 🎯 Objectif

Construire un vrai controller event-driven scalable.

## 🧠 Ce que cette étape met en lumière

Kubernetes est :

```text
massivement event-driven
```

Le controller ne boucle pas en permanence :
- il réagit aux événements cluster

## 🔄 Architecture finale

```text
Kubernetes Event
        ↓
Informer Cache
        ↓
controller-runtime
        ↓
Queue reconcile
        ↓
Reconcile()
        ↓
Convergence
```

## 🔧 Eléments importants ajoutés

### Watches()

```go
Watches(...)
```

### Owns()

```go
Owns(...)
```

### Predicates

```go
predicate.Funcs{}
```

### Event filtering

```go
CreateFunc
UpdateFunc
DeleteFunc
```

### Event Pod interception

```go
handler.EnqueueRequestsFromMapFunc(...)
```

### Indexation

```go
mgr.GetFieldIndexer()
```

### Mapping Pod → Deployment → CR

```text
Pod
 ↓
Deployment
 ↓
Custom Resource
```

## 🎯 Ce que cela apporte

Permet :
- réduction bruit events
- performances
- ciblage précis reconciliations
- réactivité temps réel
- architecture scalable

## 🧠 Event-driven réel

Le controller réagit désormais :
- aux Pods
- aux changements Ready
- aux suppressions
- aux changements phase
- aux ressources dépendantes

➡️ sans polling.

---

## 📊 Observabilité ajoutée

Le controller expose maintenant :
- métriques Prometheus
- histogrammes latence
- compteurs reconcile
- erreurs
- gauges Pods

## 🔧 RateLimiter / Backoff

Ajout :
- ExponentialBackoff
- BucketRateLimiter
- MaxConcurrentReconciles

## 🎯 Ce que cela apporte

Permet :
- éviter saturation API
- lisser charge
- éviter tempêtes reconcile
- améliorer stabilité

---

## 🔄 Fonctionnement final actuel

```text
Custom Resource
        ↓
API Server
        ↓
Events Kubernetes
        ↓
Informer Cache
        ↓
controller-runtime
        ↓
RateLimiter / Queue
        ↓
Reconcile()
        ↓
Deployment / Service / ConfigMap
        ↓
Kubernetes convergence native
        ↓
Status / Metrics / Conditions
```

## 🚀 Ce que l’Operator sait désormais faire

✅ Gestion drift  
✅ Gestion suppression  
✅ Multi-ressources  
✅ Status avancé  
✅ observedGeneration  
✅ Conditions  
✅ Event-driven  
✅ Predicates  
✅ Watches  
✅ Metrics Prometheus  
✅ Histogrammes latence  
✅ Backoff exponentiel  
✅ RateLimiting  
✅ Queue management  
✅ Ownership Kubernetes  
✅ Garbage Collection native  
✅ Réconciliation scalable  

## 🔥 Evolutions possibles

L’Operator peut encore évoluer vers :
- Webhooks
- Validation CRD
- Admission Controllers
- StatefulSets
- Jobs/CronJobs
- HorizontalPodAutoscaler
- Leader Election HA
- Prometheus Operator
- Grafana dashboards
- Helm packaging
- Multi-namespace watch
- Sharding controllers
- Profiling pprof
- Tracing OpenTelemetry
- Benchmarks control-plane massifs

---

# 🎯 Conclusion

Le controller initial :

```text
créait simplement des Pods
```

L’architecture finale :

```text
devient un véritable Operator Kubernetes
event-driven
observable
scalable
et proche des architectures production-grade
```

Le controller agit désormais comme :
- un orchestrateur de convergence
- un moteur de réconciliation distribué
- un consommateur d’évènements Kubernetes
- une extension native du control-plane Kubernetes