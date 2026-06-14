# Construction & exploitation Helm chart opérateur à partir des sources Kubebuilder

## 🎯 Présentation

Cette étape consiste à :
- transformer les manifests Kubebuilder en chart Helm,
- templatiser les ressources,
- packager l'opérateur,
- publier sur un registre OCI,
- déployer depuis un artefact distant.

---

## 🏗️ Architecture & emplacement de travail

### Répertoire principal

```text
operator/helm/controlplane-operator
```

### Structure finale

```text
Chart.yaml
values.yaml
templates/
crds/
```

### Templates créés

```text
deployment.yaml
metrics.yaml
rbac.yaml
dashboards/
stress-tests/
```

---

## ⚙️ Actions effectuées

### Décomposition operator.yaml

Kubebuilder génère :

```text
operator.yaml
```

Découpage manuel :
- Deployment,
- Service,
- RBAC,
- ServiceMonitor,
- CRDs.

### Séparation CRDs

```text
templates/crd.yaml
→ crds/
```

### Templatisation Helm

Variables déplacées dans :

```text
values.yaml
```

Exemples :

```yaml
image.repository
image.tag
namespace
service.port
```

Utilisation :

```yaml
{{ .Values.image.repository }}
```

### Lint chart

```bash
helm lint .
```

### Génération manifests

```bash
helm template controlplane-operator .
```

### Installation locale

```bash
helm install controlplane-operator . \
-n operator-system \
--create-namespace
```

### Upgrade chart

```bash
helm upgrade controlplane-operator . \
-n operator-system
```

### Stress tests dynamiques

```yaml
stressTests.scheduler.enabled
stressTests.apiServer.enabled
stressTests.etcd.enabled
```

### Packaging chart

```bash
helm package .
```

Résultat :

```text
controlplane-operator-0.1.0.tgz
```

### Authentification GitHub OCI

```bash
helm registry login ghcr.io
```

### Push OCI

```bash
helm push controlplane-operator-0.1.0.tgz \
oci://ghcr.io/newfile01/charts
```

### Déploiement distant

```bash
helm install controlplane-operator \
oci://ghcr.io/newfile01/charts/controlplane-operator
```

---

## 🔎 Vérifications

### Vérifier release Helm

```bash
helm list -A
```

### Vérifier manifests

```bash
helm template .
```

### Vérifier CRDs

```bash
kubectl get crd
```

### Vérifier stress tests

```bash
kubectl get controlplanetest -A
```

---

## ✅ Bilan

L'opérateur est désormais :
- packagé sous Helm,
- configurable dynamiquement,
- versionné,
- publiable sur OCI,
- installable à distance,
- extensible via dashboards et scénarios de stress.