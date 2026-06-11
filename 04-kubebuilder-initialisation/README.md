# 🚀 Étape #4 - Mise en place du projet opérateur avec Kubebuilder

## 🎯 Objectif

Cette étape permet de mettre en place la structure complète d'un opérateur Kubernetes à l'aide de Kubebuilder.

L'objectif est de comprendre :

* l'initialisation d'un projet Operator ;
* la génération automatique des CRDs ;
* le fonctionnement du boilerplate Kubebuilder ;
* la génération des manifests Kubernetes ;
* l'extension de l'API Kubernetes ;
* la création des premières Custom Resources (CR).

Cette étape constitue la base nécessaire avant l'implémentation de la logique métier du controller.

---

## 📚 Prérequis & sources

### 🔧 Outils nécessaires

* Go
* kubectl
* Docker
* Minikube
* Kubebuilder
* Kubernetes actif et accessible via kubeconfig

### 🔗 Sources

* https://book.kubebuilder.io/
* https://github.com/kubernetes-sigs/kubebuilder
* https://pkg.go.dev/sigs.k8s.io/controller-runtime

---

## 📁 Structure du repository

```text
k8s-controlplanelab/
├── README.md
├── 01-cluster-setup/
├── 02-prometheus-manuel/
├── 03-observability-stack/
└── operator/
```

Création du dossier :

```bash
mkdir operator
cd operator
```

---

## 🔍 Vérification de Kubebuilder

```bash
kubebuilder version
```

Exemple :

```text
KubeBuilder: v4.14.0
Kubernetes: 1.35.0
Go OS/Arch: linux/amd64
```

---

## 🏗️ Initialisation du projet Kubebuilder

```bash
kubebuilder init \
--domain=lab.local \
--repo=github.com/Newfile01/k8s-controlplanelab/operator
```

Cette commande initialise :

* le projet Go ;
* la structure Kubebuilder ;
* les dépendances controller-runtime ;
* les manifests Kustomize ;
* le Makefile ;
* le manager Operator.

---

## 📂 Structure générée par Kubebuilder

```text
operator/
├── api/
├── bin/
├── cmd/
├── config/
├── hack/
├── internal/
├── test/
├── go.mod
├── go.sum
├── main.go
├── Makefile
└── PROJECT
```

| Élément                | Rôle                             |
| ---------------------- | -------------------------------- |
| `api/`                 | Définition des APIs et CRDs      |
| `internal/controller/` | Controllers et logique Reconcile |
| `config/crd/`          | Génération des CRDs              |
| `config/rbac/`         | RBAC générés automatiquement     |
| `config/manager/`      | Déploiement de l'opérateur       |
| `main.go`              | Point d'entrée du manager        |
| `Makefile`             | Génération et déploiement        |
| `PROJECT`              | Configuration Kubebuilder        |

---

## ⚙️ Création de l'API Operator

Création de l'API :

```bash
kubebuilder create api \
--group=controlplane \
--version=v1alpha1 \
--kind=ControlPlaneTest
```

Répondre :

```text
Create Resource [y/n] y
Create Controller [y/n] y
```

---

## 🧠 Ce que Kubebuilder génère

Cette commande génère automatiquement :

```text
api/v1alpha1/controlplanetest_types.go
api/v1alpha1/groupversion_info.go
internal/controller/controlplanetest_controller.go
config/crd/
config/rbac/
config/samples/
```

---

## 📄 Définition de la CRD

Le fichier :

```text
api/v1alpha1/controlplanetest_types.go
```

contient :

* le `Spec` ;
* le `Status` ;
* les structures Go ;
* les marqueurs Kubebuilder.

Exemple :

```go
type ControlPlaneTestSpec struct {
    Foo string `json:"foo, omitempty"`
}
```

Sans :

```go
omitempty
```

le champ devient automatiquement obligatoire dans la CRD OpenAPI générée.

---

## 🔍 Génération des manifests Kubernetes

```bash
make manifests
```

Cette commande utilise :

```text
controller-gen
```

pour générer automatiquement :

* CRDs ;
* schémas OpenAPI ;
* RBAC ;
* manifests Kustomize.

Les CRDs sont générées dans :

```text
config/crd/bases/
```

---

## 🔐 Génération automatique des RBAC

Kubebuilder génère automatiquement en s'appuyant notamment sur des marqueurs de génération de codes : `+kubebuilder:...`

```text
config/rbac/
```

Ces manifests permettent :

* au controller d'accéder à l'API Kubernetes ;
* de lire/écrire des ressources ;
* d'observer les événements du cluster.

Le controller fonctionnera ensuite avec :

```text
ServiceAccount
+
Role / ClusterRole
+
RoleBinding / ClusterRoleBinding
```

comme tout composant Kubernetes classique.

---

## ☸️ Installation des CRDs dans le cluster

```bash
make install
```

Cette commande :

```text
génère les manifests
↓
applique les CRDs au cluster Kubernetes
```

---

## ⚠️ Vérification du cluster actif

Le cluster Kubernetes doit être actif et le kubeconfig doit pointer vers le bon contexte.

Vérification :

```bash
kubectl config current-context
```

Exemple attendu :

```text
control-plane-lab
```

Informations cluster :

```bash
kubectl cluster-info
```

---

## 🔄 Changement de profile Minikube

Si le mauvais cluster est utilisé :

```bash
minikube profile control-plane-lab
```

Puis vérifier :

```bash
kubectl config current-context
```

---

## 🔍 Vérification des CRDs

```bash
kubectl get crds | grep controlplane
```

Résultat attendu :

```text
controlplanetests.controlplane.lab.local
```

---

## ▶️ Exécution locale de l'opérateur

```bash
make run
```

L'opérateur s'exécute alors :

```text
depuis la machine locale
```

et dialogue avec Kubernetes via :

```text
~/.kube/config
```

Logs attendus :

```text
Starting Controller
Starting workers
```

---

## 📄 Création d'une première Custom Resource

Kubebuilder génère automatiquement un exemple dans :

```text
config/samples/
```

Exemple :

```yaml
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest
metadata:
  name: controlplanetest-sample
spec:
  foo: xxx
```

Ici nous avons copier cet exemple dans un dossier `tests/` à l'intérieur du dossier `operator/` qui constitue le dossier contenant le projet et les dépendances de notre opérateur.

---

## 🚀 Déploiement de la ressource personnalisée

On se positionne dans le dossier `operator/tests/` puis :

```bash
kubectl apply -f test1_cr.yaml

# Possibilité d'éditer le YAML à la création directement
kubectl create -f test1_cr.yaml --edit -o yaml
```

Vérification :

```bash
kubectl get controlplanetest
```

Résultat :

```text
NAME                        AGE
controlplanetest-sample     1m
```

---

## 🔍 Observation détaillée de la ressource

```bash
kubectl describe controlplanetest
```

Exemple :

```text
Spec:
  Foo: xxx
```

---

## 🔐 Validation OpenAPI automatique

Si le champ obligatoire `foo` est absent :

```bash
kubectl apply -f fichier.yaml
```

Erreur :

```text
The ControlPlaneTest "..." is invalid: spec: Required value
```

Kubebuilder :

```text
lit les structs Go
↓
génère le schéma OpenAPI
↓
Kubernetes valide automatiquement les CRs
```

---

## 🧠 Fonctionnement global obtenu

```text
CRD
↓
Custom Resource
↓
Kubebuilder Controller
↓
Reconciliation Loop
↓
Ressources Kubernetes générées automatiquement
```

Il s'agit exactement du même modèle que :

```text
Prometheus Operator
kube-prometheus-stack
ServiceMonitor
```

observés précédemment.

---

Il faut bien noter qu'à chaque modification des fichiers `customkind_types.go` ou `customKind_controller.go` il faudra effectuer à nouveau les commandes `make manifests`, pour regénérer les manifests et `make install` pour les installer à nouveau dans le cluster Kubernetes. Sans ça il y aura un décalage entre les manifests sur lesquels nous travaillons dans le dossier projet de l'opérateur et ceux réellement connus de Kubernetes. Un redémarrage de l'opérateur sera également nécessaire avec `make run`ou `go cmd/main.go`
