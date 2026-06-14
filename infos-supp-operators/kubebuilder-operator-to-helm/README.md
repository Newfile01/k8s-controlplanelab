# Construction & exploitation Helm chart opérateur à partir des sources Kubebuilder

## 🎯 Présentation

Cette étape consiste à :

* transformer les manifests Kubebuilder en chart Helm
* templatiser les ressources Kubernetes
* rendre l’installation dynamique et configurable
* packager l’opérateur
* publier le chart sur un registre OCI
* déployer l’opérateur depuis un artefact distant

Le chart Helm permet désormais :

* l’installation rapide de l’opérateur
* l’activation dynamique des scénarios de stress
* le déploiement des dashboards Grafana
* le versionning de l’opérateur

---

## 🏗️ Architecture & emplacement de travail

* Répertoire principal :

```text
operator/helm/controlplane-operator
```

### Structure finale

```text
Chart.yaml
values.yaml
templates/
crds/
charts/
```

### Structure des templates

```text
templates/
├── deployment.yaml
├── metrics.yaml
├── rbac.yaml
├── namespace.yaml
├── dashboards/
│   └── <dashboards-grafana>
└── stress-tests/
    └── <templates-scenarios>
```

### Ressources Helm générées

```text
Deployment
Service
ServiceMonitor
RBAC
CRD
Dashboards Grafana
Custom Resources de stress
```

---

## ⚙️ Actions effectuées

## 📦 Génération du chart Helm

Création initiale :

```bash
helm create controlplane-operator
```

Puis nettoyage du chart généré automatiquement :

* suppression ingress
* suppression service inutile
* suppression autoscaling
* adaptation Kubebuilder

---

## 🔧 Décomposition du `operator.yaml`

Kubebuilder génère un manifest global :

```text
operator/operator.yaml
```

Ce manifest contient :

* CRDs
* RBAC
* Deployment
* ServiceAccount
* ServiceMonitor
* Services

Pour Helm, il faut découper manuellement ce fichier en plusieurs manifests :

* `deployment.yaml`
* `rbac.yaml`
* `metrics.yaml`
* `namespace.yaml`

---

## 📚 Séparation des CRDs

Les CRDs doivent être placées dans :

```text
crds/
```

Exemple :

```text
crds/controlplanetests.controlplane.lab.local.yaml
```

Pourquoi :

* Helm installe automatiquement les CRDs avant les autres ressources
* évite les erreurs de dépendances CRD inexistantes
* séparation claire entre API Kubernetes et templates Helm

---

## 🧩 Templatisation Helm

Les manifests statiques sont remplacés par des variables Helm :

### Variables déplacées dans `values.yaml`

```yaml
namespace: operator-system

image:
  repository: newfile01/operator-k8s
  tag: v6
  pullPolicy: Always

service:
  port: 8443
```

### Utilisation dans les templates

```yaml
namespace: {{ .Values.namespace }}

image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
```

Cela permet :

* plusieurs environnements
* changement dynamique d’image
* activation conditionnelle des scénarios
* packaging réutilisable

---

## 🧪 Gestion des scénarios de stress

Des templates Helm dédiés sont ajoutés :

```text
templates/stress-tests/
```

Exemples :

* `scheduler-stress.yaml`
* `apiserver-stress.yaml`
* `etcd-stress.yaml`

Activation conditionnelle :

```yaml
{{- if .Values.stressTests.scheduler.enabled }}
```

Configuration via :

```yaml
stressTests:
  scheduler:
    enabled: true
```

---

## 📊 Intégration dashboards Grafana

Ajout d’un dossier :

```text
templates/dashboards/
```

Contenant :

* dashboards JSON
* ConfigMaps Grafana
* labels Grafana sidecar

Les dashboards sont automatiquement importés par Grafana via :

* `grafana_dashboard=1`

---

## 🔎 Vérifications du chart

Depuis la racine du chart :

```text
operator/helm/controlplane-operator/
```

### Génération des manifests

```bash
helm template controlplane-operator .
```

Permet :

* vérifier le rendu YAML final
* valider les variables Helm
* détecter les erreurs de templating

### Validation Helm

```bash
helm lint .
```

Permet :

* vérifier structure chart
* détecter YAML invalides
* détecter erreurs Helm

---

## 🚀 Installation locale

### Installation standard

```bash
helm install controlplane-operator . \
-n operator-system \
--create-namespace
```

### Installation avec stress test

```bash
helm install controlplane-operator . \
-n operator-system \
--create-namespace \
--set stressTests.scheduler.enabled=true
```

---

## 🔄 Upgrade du chart

Après modification des manifests Helm :

```bash
helm upgrade controlplane-operator . \
-n operator-system
```

Exemple avec activation d’un scénario :

```bash
helm upgrade controlplane-operator . \
-n operator-system \
--set stressTests.scheduler.enabled=true
```

Important :

* seuls les manifests impactés sont mis à jour
* les CRDs existantes ne sont pas supprimées
* les CRs existantes sont conservées

---

## 📦 Packaging Helm

### Génération package OCI

```bash
helm package .
```

Résultat :

```text
controlplane-operator-0.1.0.tgz
```

---

## 🔐 Authentification GitHub OCI

Créer un token GitHub :

* `read:packages`
* `write:packages`

Puis :

```bash
export CR_PAT=<github_token>

echo $CR_PAT | helm registry login ghcr.io \
-u <github_user> \
--password-stdin
```

---

## ☁️ Push OCI Registry

```bash
helm push controlplane-operator-0.1.0.tgz \
oci://ghcr.io/newfile01/charts
```

Le chart devient installable à distance depuis GHCR.

---

## 🌍 Déploiement distant

### Depuis OCI Registry

```bash
helm install controlplane-operator \
oci://ghcr.io/newfile01/charts/controlplane-operator \
-n operator-system \
--create-namespace
```

### Depuis package local

```bash
helm install controlplane-operator \
./controlplane-operator-0.1.0.tgz \
-n operator-system \
--create-namespace
```

---

## 🧰 Commandes utiles

```bash
# Génération manifests
helm template controlplane-operator .

# Validation chart
helm lint .

# Installation
helm install controlplane-operator . \
-n operator-system \
--create-namespace

# Upgrade
helm upgrade controlplane-operator . \
-n operator-system

# Historique release
helm history controlplane-operator \
-n operator-system

# Rollback
helm rollback controlplane-operator 1 \
-n operator-system

# Packaging
helm package .

# Installation depuis package
helm install controlplane-operator \
./controlplane-operator-0.1.0.tgz \
-n operator-system \
--create-namespace

# Désinstallation
helm uninstall controlplane-operator \
-n operator-system

# Suppression CRD
kubectl delete crd \
controlplanetests.controlplane.lab.local
```

---

## 🔄 Mise à jour du chart OCI

### 1) Modifier `Chart.yaml`

```yaml
version: 0.1.1
appVersion: "v7"
```

### 2) Regénérer le package

```bash
helm package .
```

### 3) Republier OCI

```bash
helm push controlplane-operator-0.1.1.tgz \
oci://ghcr.io/newfile01/charts
```

### 4) Upgrade distant

```bash
helm upgrade controlplane-operator \
oci://ghcr.io/newfile01/charts/controlplane-operator \
--version 0.1.1 \
-n operator-system
```

---

## 🔎 Vérifications

```bash
# Vérifier release Helm
helm list -A

# Vérifier ressources opérateur
kubectl get all -n operator-system

# Vérifier CRDs
kubectl get crd

# Vérifier scénarios
kubectl get controlplanetest -A

# Vérifier ServiceMonitor
kubectl get servicemonitor -A

# Vérifier logs opérateur
kubectl logs -n operator-system \
deployment/operator-controller-manager -f

# Vérifier historique Helm
helm history controlplane-operator \
-n operator-system
```

---

## ✅ Bilan

L’opérateur est désormais :

* packagé sous Helm
* configurable dynamiquement
* versionné
* publiable sur OCI
* installable localement ou à distance
* extensible via dashboards et scénarios de stress
* facilement maintenable et distribuable
