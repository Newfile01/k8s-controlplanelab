# 05 - 🚀 OPERATEUR EVENT-DRIVEN

## 🎯 Objectif global

Depuis l’étape `observedGeneration` jusqu’aux optimisations avancées du `controller-runtime`, l’objectif a été de transformer l’opérateur :

```text
d'un simple controller Kubernetes "fonctionnel"
```

vers :

```text
un opérateur réellement scalable, événementiel et optimisé
```

Ces étapes ont permis d’ajouter :

- suivi de génération observée (`observedGeneration`)
- ressources secondaires (`Service`, `ConfigMap`)
- interception d’évènements Kubernetes (`Watch`)
- mapping Pod → Deployment → CR
- indexation locale
- filtrage avancé d’évènements (`Predicates`)
- filtrage sémantique (`PodReady`, `Phase`, `Delete`)
- limitation de charge (`RateLimiter`)
- limitation concurrence (`MaxConcurrentReconciles`)
- backoff exponentiel (`ExponentialFailureRateLimiter`)
- limitation débit global (`BucketRateLimiter`)

---

# 📊 ARCHITECTURE AVANT / APRES

## ⚠️ Avant les optimisations

```text
Kubernetes Event
        ↓
controller-runtime Watch
        ↓
Reconcile()
        ↓
GET / UPDATE
        ↓
Nouvelle Reconcile
```

Problèmes :
- beaucoup trop d’évènements
- aucune limitation
- retries agressifs
- bruit important
- nombreuses réconciliations inutiles

---

## ✅ Après toutes les optimisations

```text
Kubernetes Event
        ↓
Predicate Filtering
        ↓
Watch Pod / Deployment / Service / ConfigMap
        ↓
Mapping Pod → Deployment → CR
        ↓
Workqueue controller-runtime
        ↓
RateLimiter / BucketLimiter / Backoff
        ↓
Workers limités
        ↓
Reconcile()
        ↓
PATCH / UPDATE minimal
        ↓
Status / Conditions / observedGeneration
```

Résultats :
- moins de bruit
- moins de charge API Server
- moins de charge etcd
- retries contrôlés
- meilleure stabilité
- meilleure scalabilité
- logique réellement event-driven

---

# 🧩 ETAPE 3 — observedGeneration

## 🎯 Rôle

`observedGeneration` permet de savoir :

```text
si le Status correspond bien à la dernière version du Spec
```

Kubernetes incrémente automatiquement :

```yaml
metadata.generation
```

à chaque modification du `spec`.

L’opérateur recopie ensuite cette valeur dans :

```yaml
status.observedGeneration
```

➡️ Cela permet de savoir si la dernière génération utilisateur a bien été traitée.

---

## 📍 Code à ajouter

### Dans `_types.go`

```go
type ControlPlaneTestStatus struct {
	DeploymentName     string             `json:"deploymentName,omitempty"`
	ReadyReplicas      int32              `json:"readyReplicas,omitempty"`
	AvailableReplicas  int32              `json:"availableReplicas,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}
```

---

### Dans `Reconcile()`

```go
controlPlaneTest.Status.ObservedGeneration =
	controlPlaneTest.Generation
```

Avant :

```go
err = r.Status().Update(ctx, controlPlaneTest)
```

---

## 🧪 Vérification

```bash
kubectl patch controlplanetest nginx-test \
--type merge \
-p '{"spec":{"replicas":4}}'

kubectl get controlplanetest nginx-test -o yaml
```

---

## ✅ Résultat attendu

```yaml
metadata:
  generation: 6

status:
  observedGeneration: 6
```

---

# 🧩 ETAPE 4 — Service + ConfigMap

## 🎯 Rôle

Ajout de ressources secondaires gérées automatiquement :

- `Service`
- `ConfigMap`

➡️ démonstration d’un opérateur multi-ressources.

---

## 📍 Code à ajouter

### Création ConfigMap

```go
configMap := &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:      configMapName,
		Namespace: req.Namespace,
	},
	Data: map[string]string{
		"index.html": "Bonjour depuis ConfigMap",
	},
}
```

---

### Création Service

```go
service := &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:      serviceName,
		Namespace: req.Namespace,
	},
	Spec: corev1.ServiceSpec{
		Selector: map[string]string{
			"app": controlPlaneTest.Name,
		},
		Ports: []corev1.ServicePort{
			{
				Port: 80,
			},
		},
	},
}
```

---

## 🧪 Vérification

```bash
kubectl get deployment
kubectl get service
kubectl get configmap
```

---

## ✅ Résultat attendu

```text
nginx-test-deployment
nginx-test-service
nginx-test-configmap
```

---

# 🧩 ETAPE 5 — Watches / Owns

## 🎯 Rôle

Permettre à l’opérateur de réagir :

```text
aux modifications des ressources secondaires
```

sans passer par une modification directe de la CR.

---

## 📍 Code à ajouter

### Dans `SetupWithManager()`

```go
return ctrl.NewControllerManagedBy(mgr).
	For(&controlplanev1alpha1.ControlPlaneTest{}).
	Owns(&appsv1.Deployment{}).
	Owns(&corev1.Service{}).
	Owns(&corev1.ConfigMap{}).
```

---

## 🧪 Vérification

```bash
kubectl delete deployment nginx-test-deployment
```

---

## ✅ Résultat attendu

Le Deployment est automatiquement recréé.

---

# 🧩 ETAPE 6 — Indexation locale

## 🎯 Rôle

Créer un index mémoire local dans `controller-runtime` afin de retrouver rapidement :

```text
Deployment → CR parent
```

➡️ évite des recherches coûteuses sur tout le cluster.

---

## 📍 Code à ajouter

### Dans `SetupWithManager()`

```go
if err := mgr.GetFieldIndexer().IndexField(
	context.Background(),
	&appsv1.Deployment{},
	deploymentOwnerKey,
	func(rawObj client.Object) []string {

		deployment := rawObj.(*appsv1.Deployment)

		owner := metav1.GetControllerOf(deployment)

		if owner == nil {
			return nil
		}

		if owner.APIVersion != controlplanev1alpha1.GroupVersion.String() ||
			owner.Kind != "ControlPlaneTest" {
			return nil
		}

		return []string{owner.Name}
	},
); err != nil {
	return err
}
```

---

## 🧪 Vérification

```bash
kubectl delete deployment nginx-test-deployment
```

---

## ✅ Résultat attendu

Réconciliation immédiate via le mapping indexé.

---

# 🧩 ETAPE 7 — Watches Pod + Predicates

## 🎯 Rôle

Intercepter directement les évènements Pod utiles :

- suppression
- changement Ready
- changement Phase

Tout en filtrant :
- le bruit Kubernetes
- les Pods non concernés

---

## 📍 Code à ajouter

### Predicate avancé

```go
podPredicate := predicate.Funcs{

	UpdateFunc: func(e event.UpdateEvent) bool {

		oldPod := e.ObjectOld.(*corev1.Pod)
		newPod := e.ObjectNew.(*corev1.Pod)

		if oldPod.Status.Phase != newPod.Status.Phase {
			return true
		}

		oldReady := isPodReady(oldPod)
		newReady := isPodReady(newPod)

		return oldReady != newReady
	},
}
```

---

### Watch Pod

```go
.Watches(
	&corev1.Pod{},
	handler.EnqueueRequestsFromMapFunc(
		func(ctx context.Context, obj client.Object) []reconcile.Request {
```

---

## 🧪 Vérification

```bash
kubectl delete pod <pod-name>
```

---

## ✅ Résultat attendu

Logs :

```text
☸️🗑️📦 │ Suppression Pod détectée
☸️🔄📊 │ Changement Ready Pod détecté
```

---

# 🧩 ETAPE 8 — RateLimiter / Backoff / Workqueue

## 🎯 Rôle

Limiter :
- les réconciliations simultanées
- les retries agressifs
- les tempêtes d’évènements Kubernetes

---

## 📍 Code à ajouter

### Imports

```go
import (
	"time"

	"golang.org/x/time/rate"

	"k8s.io/client-go/util/workqueue"

	"sigs.k8s.io/controller-runtime/pkg/controller"
)
```

---

### Dans `SetupWithManager()`

```go
.WithOptions(controller.Options{
	MaxConcurrentReconciles: 2,
	RateLimiter: workqueue.NewTypedMaxOfRateLimiter(
		workqueue.NewTypedItemExponentialFailureRateLimiter[reconcile.Request](
			1*time.Second,
			30*time.Second,
		),
		&workqueue.TypedBucketRateLimiter[reconcile.Request]{
			Limiter: rate.NewLimiter(
				rate.Limit(10),
				100,
			),
		},
	),
})
```

---

## 🧪 Vérification

### Génération massive d’évènements

```bash
kubectl delete pod -l app=nginx-test
```

---

## ✅ Résultat attendu

- réconciliations progressives
- pas de tempête CPU
- retries espacés
- opérateur stable

---

# 🔥 SYNTHESE GLOBALE — CHAINE COMPLETE

```text
═══════════════════════════════════════════════

[KUBERNETES]

- Evènement cluster :
  Pod / Deployment / CR / Service / ConfigMap

═══════════════════════════════════════════════

[CONTROLLER-RUNTIME]

- Watch
- Informer
- Cache
- Predicate filtering
- Mapping ressources
- Indexation locale
- Workqueue
- RateLimiter
- BucketRateLimiter
- Backoff exponentiel
- Workers reconcile

═══════════════════════════════════════════════

[OPERATEUR]

Reconcile()

- GET ressources
- Vérification convergence
- Création / Update / Delete
- Mise à jour Status
- Conditions
- observedGeneration

═══════════════════════════════════════════════

[KUBERNETES]

- Application état désiré
- ReplicaSet
- Pods
- Self-healing
- RollingUpdate

═══════════════════════════════════════════════

[CONTROLLER-RUNTIME]

- Nouveaux évènements
- Nouvelle queue
- Nouvelle reconcile si nécessaire

═══════════════════════════════════════════════
```
