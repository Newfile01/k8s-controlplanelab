# 👀 REPERTOIRE POUR LA SURVEILLANCE DES RESSOURCES ENFANTS AVEC `Owns()`

Cette étape transforme le controller précédent en véritable Operator réactif capable de surveiller automatiquement les ressources Kubernetes qu'il génère.

Jusqu'ici :

```text
CR modifiée
        ↓
Reconcile()
        ↓
Mise à jour Deployment
```

Mais :

```text
Modification manuelle du Deployment
        ↓
Aucune reconciliation automatique
```

L'ajout de `Owns()` permet désormais :

```text
Modification CR
        ↓
OU
        ↓
Modification Deployment
        ↓
Reconcile()
        ↓
Correction automatique des drifts
```

---

# 🎯 Objectif

Permettre à l'Operator de :

* surveiller les Deployments qu'il possède ;
* détecter les modifications manuelles ;
* reconverger automatiquement vers l'état désiré défini dans la CR.

Exemple :

```text
CR : replicas = 4
        ↓
Utilisateur modifie Deployment à 10 replicas
        ↓
Operator détecte la dérive
        ↓
Reconcile()
        ↓
Retour automatique à 4 replicas
```

---

# 🧠 Problème observé avant `Owns()`

Même avec :

```go
controllerutil.SetControllerReference(...)
```

le Deployment :

* dépendait bien de la CR ;
* était bien supprimé automatiquement ;
* MAIS n'était pas surveillé.

Donc :

```text
kubectl scale deployment nginx-test-deployment --replicas=10
```

créait une divergence sans déclencher de nouvelle reconciliation :

```text
Etat désiré (CR) = 4 replicas
Etat réel (Deployment) = 10 replicas
```



---

# 🔥 Différence entre `OwnerReference` et `Owns()`

| Fonction                   | Rôle                                             |
| -------------------------- | ------------------------------------------------ |
| `SetControllerReference()` | crée la relation parent/enfant                   |
| `Owns()`                   | surveille automatiquement les ressources enfants |

---

# 🚀 Modification du controller

Dans `SetupWithManager()` (à la fin de `...controller.go`):

Avant :

```go
return ctrl.NewControllerManagedBy(mgr).
	For(&controlplanev1alpha1.ControlPlaneTest{}).
	Named("controlplanetest").
	Complete(r)
```

Après :

```go
return ctrl.NewControllerManagedBy(mgr).
	For(&controlplanev1alpha1.ControlPlaneTest{}).
	Owns(&appsv1.Deployment{}).
	Named("controlplanetest").
	Complete(r)
```

---

# 🔍 Ce que fait réellement `Owns()`

Le controller-runtime ajoute automatiquement :

```text
WATCH sur les Deployments
```

Lorsqu'un Deployment possédé par la CR change :

```text
UPDATE événement
        ↓
Nouvelle reconciliation
```

---

# 🔄 Fonctionnement obtenu

```text
ControlPlaneTest
        ↓
Operator crée Deployment
        ↓
Deployment possède OwnerReference
        ↓
Operator surveille Deployment grâce à Owns()
        ↓
Modification manuelle détectée
        ↓
Nouvelle reconciliation
        ↓
Correction automatique du drift
```

---

# 🧪 Exemple de test

## 📦 Relancer l'Operator
Depuis le dossier `operator/` :

```bash
make run
```

ou :

```bash
make manifests
make install
```

## Créer les ressources & modifiers les enfants
```bash
# Depuis le dossier du manifeste
kubectl apply -f nginx-test.yaml

# Vérif, éventuellement rajouter '-o yaml'
kubectl get controlplanetest

# Modif nombre de replicas
kubectl scale deployment nginx-test-deployment \
--replicas=10
```

Dans un second terminal :

```bash
watch -n 1 kubectl get pods -w
```

Résultat attendu :

```text
10 replicas
↓
Reconcile()
↓
4 replicas
```

---

# 🧠 Ce que cette étape apporte réellement

Cette étape transforme l'Operator :

```text
Controller CRUD simple
        ↓
Operator déclaratif réactif
```

Le controller ne réagit plus seulement aux modifications de la CR **il surveille désormais l'état réel du cluster** et corrige automatiquement les dérives observées.
