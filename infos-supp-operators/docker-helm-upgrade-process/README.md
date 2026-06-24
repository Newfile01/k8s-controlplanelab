# PROCEDURE DE MÀJ OPERATEUR

> Les commandessuivantes  permettent de :
>
> - Générer & tester les manifests Kubernetes créés avec Kubebuilder
> - Générer une image Docker de notre opérateur et la publier sur un repository DockerHub
> - Générer & Vérifier un Helm chart issu de notre code opérateur créé via kubebuilder
> - Packager & Publier l'opérateur sous forme de Helm Chart sur un registre publique OCI (accessible via Token uniquement)
> - Mettre à jour l'opérateur déjà en place dans le cluster depuis le registre OCI (source distante)

**NOTE** : Modifier les variables en fonction de ses paramètres

---

# 📝🐋 MANIFESTS & DOCKER

Depuis `/operator`

Ici : 

📁 `/k8s-controlplanelab/operator`

```text

|   ...
├── Dockerfile
├── Makefile
├── ...
├── api
│   └── v1alpha1
│       ├── controlplanetest_types.go
│       └── ...
├── bin
│   └── ...
├── cmd
│   └── main.go
├── config
│   ├── crd
│   │   ├── bases
│   │   └── ...
│   ├── default
│   │   ├── ...
│   │   └── metrics_service.yaml
│   ├── manager
│   │   └── ...
│   ├── network-policy
│   │   └── ...
│   ├── prometheus
│   │   └── ...
│   │   ├── monitor.yaml
│   │   ...
│   ├── rbac
│   │   ├── ...
│   │   ├── role.yaml
│   │   ├── role_binding.yaml
│   │   └── service_account.yaml
│   └── ...
├── ;...
├── hack
│   └── ...
├── helm
│   └── ...
├── internal
│   └── controller
│       ├── controlplanetest_controller.go
│       └── ...
├── operator.yaml
└──  test
    ├── e2e
    │   └── ...
    └── utils
        └── ...
```

```bash
# ========================================
# VARIABLES
# ========================================

DOCKER_VERSION="v7"
DOCKER_REPO="newfile01/operator-k8s"

# ============================================================
# PARTIE KUBEBUILDER
# ============================================================

# Régénération des CRDs Kubernetes et manifests associés
# à partir des annotations Kubebuilder présentes dans le code Go
make manifests

# Régénération du code Go généré automatiquement
# (DeepCopy, objets Kubernetes, etc.)
make generate

# Formatage automatique de tout le code Go du projet
# selon les conventions officielles Go
make fmt

# Analyse statique du code Go
# détecte erreurs, variables inutilisées, types incorrects, etc.
make vet
# ============================================================
# CONSTRUCTION ET PUBLICATION IMAGE DOCKER
# ============================================================
docker login

docker build \
  -t ${DOCKER_REPO}:${DOCKER_VERSION} .

docker push \
  ${DOCKER_REPO}:${DOCKER_VERSION}

# ============================================================
# DEPLOIEMENT DIRECT DEPUIS LES SOURCES KUBEBUILDER
# ============================================================
# (utile pour le développement sans Helm)

make undeploy

make deploy \
  IMG=${DOCKER_REPO}:${DOCKER_VERSION}
```

---

## 📦 HELM CHART

Depuis `/operator/helm/<operateur-directory>`

Ici :

📁 `/k8s-controlplanelab/operator/helm/controlplane-operator`

```text
├── Chart.yaml
├── charts
├── crds
│   └── controlplanetests.controlplane.lab.local.yaml
├── templates
│   ├── NOTES.txt
│   ├── _helpers.tpl
│   ├── deployment.yaml
│   ├── metrics.yaml
│   ├── namespace.yaml
│   └── rbac.yaml
└── values.yaml
```

```bash
# ========================================
# VARIABLES
# ========================================
CHART_VERSION="0.1.10"
APP_VERSION="v15"

OCI_REGISTRY="ghcr.io/newfile01/charts"
GH_TOKEN="<github_token>"

# ============================================================
# PREPARATION DU HELM CHART
# ============================================================

# Mettre à jour manuellement :
#
# Chart.yaml
#   version: ${CHART_VERSION}
#   appVersion: "${APP_VERSION}"
#
# values.yaml
#   image:
#     repository: ${DOCKER_REPO}
#     tag: ${DOCKER_VERSION}

# Ou automatiquement

# ============================================================
# MàJ CHART.YAML
# ============================================================
# version: X.Y.Z
# appVersion: vX
sed -i \
  -e "s/^version:.*/version: ${CHART_VERSION}/" \
  -e "s/^appVersion:.*/appVersion: \"${APP_VERSION}\"/" \
  Chart.yaml


grep -E '^(version|appVersion):' Chart.yaml

# ============================================================
# MàJ VALUES.YAML
# ============================================================
sed -i \
  -e "s|^\([[:space:]]*repository:\).*|\1 ${DOCKER_REPO}|" \
  -e "s|^\([[:space:]]*tag:\).*|\1 ${DOCKER_VERSION}|" \
  values.yaml

grep -A2 '^image:' values.yaml

# IMPORTANT : bien supprimé ancienne archive créé en packageant
rm -rf *.tgz

# ============================================================
# VERIFICATION DU HELM CHART & PACKAGING
# ============================================================

# Vérification de la cohérence du chart :
# syntaxe YAML, templates Helm, références .Values, etc.
helm lint .

# Génération des manifests Kubernetes finaux
# (aucun déploiement, simple rendu des templates Helm)
helm template controlplane-operator . > rendered.yaml

# Simulation complète d'une installation Helm
# avec affichage détaillé du rendu et des éventuelles erreurs
helm install test . \
  --namespace operator-system \
  --create-namespace \
  --dry-run --debug

# Génération du package Helm
helm package .

# ============================================================
# PUBLICATION DU HELM CHART SUR LE REGISTRE OCI
# ============================================================

echo "${GH_TOKEN}" | helm registry login ghcr.io \
  -u Newfile01 \
  --password-stdin

helm push \
  controlplane-operator-${CHART_VERSION}.tgz \
  oci://${OCI_REGISTRY}

# ============================================================
# INSTALLATION / MISE A JOUR DEPUIS LE REGISTRE OCI
# ============================================================

helm list -A

helm uninstall controlplane-operator
# IMPORTANT SI CHANGEMENTS DANS LES CRDs
kubectl delete crd controlplanetests.controlplane.lab.local

helm install controlplane-operator \
  oci://${OCI_REGISTRY}/controlplane-operator \
  --version ${CHART_VERSION} \
  -n operator-system \
  --create-namespace

# OU

helm upgrade --install controlplane-operator \
  oci://${OCI_REGISTRY}/controlplane-operator \
  --version ${CHART_VERSION} \
  -n operator-system \
  --create-namespace
```

---

## 🔎 VERIFICATION FINALES (MINIMALES)

```bash
# ============================================================
# TEST RAPIDE DE L'OPERATEUR
# ============================================================

kubectl get pods -n operator-system
kubectl get servicemonitor -A
kubectl get clusterrole operator-manager-role -o yaml
kubectl logs -n operator-system \
  deployment/operator-controller-manager -f

```

