# Déploiement opérateur depuis les sources Kubebuilder

## 🎯 Présentation

Cette étape consiste à :
- générer l'opérateur depuis Kubebuilder,
- construire l'image,
- déployer l'opérateur dans Kubernetes,
- comprendre les ressources générées,
- vérifier son bon fonctionnement.

Ressources concernées :
- Deployment,
- ServiceAccount,
- RBAC,
- CRD,
- Controller Manager.

---

## 🏗️ Architecture & emplacement de travail

### Répertoire principal

```text
operator/
```

### Répertoires importants

```text
api/
internal/controller/
config/
```

### Architecture générale

```text
CRD
→ API Server
→ Controller Runtime
→ Reconcile Loop
→ Ressources Kubernetes
```

---

## ⚙️ Actions effectuées

### Génération du projet Kubebuilder

Commande :

```bash
kubebuilder init
kubebuilder create api
```

### Construction image Docker

```bash
make docker-build docker-push IMG=newfile01/operator-k8s:vX
```

### Déploiement opérateur

```bash
make deploy IMG=newfile01/operator-k8s:vX
```

### Vérification des ressources

```bash
kubectl get all -A | grep operator
```

### Vérification logs opérateur

```bash
kubectl logs -n operator-system deployment/operator-controller-manager -f
```

---

## 🔎 Vérifications

### Vérifier le pod opérateur

```bash
kubectl get pods -n operator-system
```

### Vérifier les CRDs

```bash
kubectl get crd
```

### Vérifier les événements

```bash
kubectl get events -A
```

---

## ✅ Bilan

L'opérateur Kubebuilder est maintenant :
- compilé,
- déployé,
- supervisé par Kubernetes,
- capable de réconcilier les CRs personnalisées.

Le fonctionnement général controller-runtime/Kubebuilder est désormais maîtrisé.