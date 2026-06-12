# 03 - 🧹 REPERTOIRE POUR LA GESTION DES FINALIZERS (SUPPRESSION CONTROLEE DES RESSOURCES)
Cette étape ajoute la gestion du cycle de suppression des Custom Resources grâce aux `Finalizers Kubernetes`.

L'objectif est de permettre à l'Operator :

* d'intercepter une suppression ;
* d'effectuer des opérations de nettoyage ;
* puis seulement ensuite autoriser la suppression réelle de la ressource.

---

# 🔄 Boucle complète observée

```text
CR créée
        ↓
Ajout automatique finalizer
        ↓
Utilisateur demande suppression
        ↓
deletionTimestamp ajouté par Kubernetes
        ↓
CR reste en Terminating
        ↓
Nouvelle reconciliation
        ↓
Cleanup custom
        ↓
Suppression finalizer
        ↓
Suppression réelle CR
        ↓
OwnerReferences suppriment Deployment
        ↓
ReplicaSet supprimé
        ↓
Pods supprimés
```

---

# 🎯 Objectif des Finalizers

Sans finalizer :

```text
DELETE CR
        ↓
Suppression immédiate
        ↓
OwnerReferences suppriment le Deployment
```

Avec finalizer :

```text
DELETE CR
        ↓
Ressource passe en "Terminating"
        ↓
L'Operator garde temporairement la ressource
        ↓
Nettoyage / vérifications
        ↓
Suppression du finalizer
        ↓
Suppression réelle
```

---

# 🧠 Fonctionnement détaillé

## 📌 1. Création normale de la CR

Exemple :

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest
metadata:
  name: nginx-test
spec:
  image: nginx:latest
  replicas: 3
```

L'Operator ajoute automatiquement :

```yaml
metadata:
  finalizers:
  - controlplanetest.lab.local/finalizer
```

---

## 📌 2. Demande de suppression

Commande :

```bash
kubectl delete controlplanetest nginx-test
```

Kubernetes NE supprime PAS immédiatement la ressource.

L'API Server :

* ajoute automatiquement un `deletionTimestamp` ;
* conserve la ressource ;
* place la CR en état `Terminating`.

Exemple :

```yaml
metadata:
  deletionTimestamp: "2026-06-09T14:00:00Z"
  finalizers:
  - controlplanetest.lab.local/finalizer
```

---

## 📌 3. Nouvelle boucle de réconciliation

La modification de la ressource déclenche automatiquement :

```text
UPDATE événement
        ↓
Reconcile()
```

Le controller détecte alors :

```go
if !controlPlaneTest.ObjectMeta.DeletionTimestamp.IsZero()
```

ce qui signifie :

```text
"la ressource est en cours de suppression"
```

---

## 📌 4. Nettoyage avant suppression réelle

Tant que le finalizer existe :

```text
Kubernetes refuse la suppression réelle
```

L'Operator peut alors :

* attendre une condition ;
* supprimer des ressources externes ;
* sauvegarder des données ;
* désenregistrer un élément ;
* effectuer un cleanup personnalisé.

Dans notre cas pédagogique :

```text
CR -> Deployment -> ReplicaSet -> Pods
```

Les `OwnerReferences` suffisent déjà à supprimer automatiquement le Deployment et ses Pods.

Le Finalizer sert ici principalement à comprendre le fonctionnement du lifecycle complet d'un Operator Kubernetes.

---

## 📌 5. Suppression réelle

Quand le nettoyage est terminé :

```go
controllerutil.RemoveFinalizer(...)
```

Puis :

```go
r.Update(...)
```

Kubernetes constate alors :

```text
plus aucun finalizer
```

et autorise finalement :

```text
la suppression réelle de la ressource
```


---

# ⚠️ Point très important

Le controller-runtime boucle PAS automatiquement en permanence. Une nouvelle reconciliation est déclenchée uniquement par :

* création ;
* update ;
* suppression ;
* watch ;
* requeue manuel.

Si une condition externe doit être attendue :

```go
return ctrl.Result{
    RequeueAfter: 10 * time.Second,
}, nil
```

permet de relancer périodiquement la reconciliation.

---

# 📦 Comment utiliser cette version ?

Depuis le dossier `operator/` (fichers `..._types.go`et `..._controller.go` remplacés):

```bash
make generate
make manifests
make install
make run
```

Puis créer une CR :

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest
metadata:
  name: nginx-test
spec:
  image: nginx:latest
  replicas: 3
```

```bash
# En se positionant dans le dossier du manifeste correspondant
kubectl apply -f nginx-test.yaml
```

Suppression :

```bash
kubectl delete controlplanetest nginx-test
```

Observer :

```bash
# ajouter 'watch -n 1' devant la commande pour l'observer en continu
kubectl get controlplanetest
kubectl describe controlplanetest nginx-test
```

La ressource passera temporairement en `Terminating` tant que le `Finalizer` sera présent.


