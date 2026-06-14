# Déploiement opérateur depuis les sources Kubebuilder

## 🎯 Présentation

Cette étape consiste à :
- générer l'opérateur depuis Kubebuilder
- construire l'image
- déployer l'opérateur dans Kubernetes
- comprendre les ressources générées
- vérifier son bon fonctionnement

Ressources concernées :
- CRD
- CR
- Deployment
- ConfigMaps
- Services
- ServiceAccount
- RBAC
- Controller Manager

---

## 🏗️ Architecture & emplacement de travail

### Répertoire principal

(Celui généré par Kubebuilder suite à `kubebuilder init`et `kubebuilder create api`)

```text
operator/
```

### Architecture générale

```text
CRD
→ API Server
→ etcd
→ Controller Runtime
→ Reconcile Loop
→ Ressources Kubernetes
```

Explication :
- la CRD définit un nouveau type de ressource Kubernetes
- l'API Server enregistre cette ressource
- etcd stocke les objets créés
- controller-runtime observe les événements Kubernetes
- le reconcile loop applique l'état désiré défini par la CR

### Répertoires importants générés par Kubebuilder

```text
api/
    → définition des types CRD

internal/controller/
    → logique métier du controller

config/
    → manifests Kubernetes générés

cmd/main.go
    → point d'entrée principal du manager/controller-runtime
```

---

## ⚙️ Actions effectuées

(à partir d'un projet Kubebuilder déjà initié/généré)

* Ce déplacer dans le dossier `operator/`
* Construction image Docker `make docker-build docker-push IMG=newfile01/operator-k8s:vX`
* Déploiement opérateur `make deploy IMG=newfile01/operator-k8s:vX`

* Remplacer 'X' par le numéro de version du tag 

Ex. de workflow :

```bash
# Rebuild
docker build -t newfile01/operator-k8s:v6 .
docker push newfile01/operator-k8s:v6
make undeploy
make deploy IMG=newfile01/operator-k8s:v6
```
Kubebuilder utilise :
- `config/default/kustomization.yaml`

Qui référence :
- RBAC
- manager
- metrics
- CRDs
- patches

Puis :
- génère un manifest global
- applique le tout avec `kubectl apply`

---

## 🔎 Vérifications

```bash
# Lancer une CR de test
kubectl apply -f tests/scenario_base.yaml
# Vérification de lancement correcte
# - Pod operator doit apparaître
# - Service operator doit apparaître
# - "kubectl get servicemonitor -A" doit faire apparaître 'operator...'
kubectl get pods -n operator-system
kubectl get clusterrole operator-manager-role -o yaml
# ressources opérateur
kubectl get all -A | grep operator
# logs opérateur
kubectl logs -n operator-system deployment/operator-controller-manager -f
# CRDs
kubectl get crd
# events
kubectl get events -A
# Vérification ServiceMonitor
kubectl get servicemonitor -A
```

**Observations attendues**

Le Pod operator doit :
- être `Running`
- rester stable
- ne pas redémarrer en boucle

Les ressources suivantes doivent apparaître :
- Deployment
- ServiceAccount
- Service
- ClusterRole
- ClusterRoleBinding
- CRD

Le ServiceMonitor doit apparaître si :
- Prometheus Operator est installé
- les manifests metrics ont été appliqués

---

## ✅ Bilan

L'opérateur Kubebuilder est maintenant :
- compilé
- containerisé
- publié sur un registre Docker
- déployé dans Kubernetes
- supervisé par Kubernetes
- capable de réconcilier les CRs personnalisées

Maîtrise du fonctionnement général de :
- Kubebuilder
- controller-runtime
- reconcile loop
- CRD/CR
- RBAC
- manager controller-runtime

Le cluster est maintenant prêt pour :
- les stress tests control-plane
- l'observabilité Prometheus/Grafana
- le packaging Helm
- les scénarios avancés Kubernetes