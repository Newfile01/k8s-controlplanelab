# Prise en charge du monitoring par Prometheus Operator

## 🎯 Présentation

L'objectif de cette étape est de rendre l'opérateur observable par Prometheus Operator. 

Pour ce faire cette étape introduit :
- Service Kubernetes
- ServiceMonitor Prometheus Operator
- exposition endpoint `/metrics`

Le but est de permettre à Prometheus de découvrir automatiquement l'opérateur afin de :
- collecter ses métriques,
- superviser son comportement,
- visualiser ses performances dans Grafana.

L'opérateur expose ses métriques via la librairie Prometheus intégrée à `controller-runtime`.

---

## 🏗️ Architecture & emplacement de travail

### Répertoire principal

```text
operator/config/default/
    - kustomization.yaml
        → Liste les manifests à appliquer pour le déploiement standard de l'opérateur
        (les éléments à déployer pour les métriques de l'opérateur custom)

    - manager_metrics_patch.yaml
        → Patch appliqué au Deployment manager afin d'activer/configurer le serveur métriques exposé par controller-runtime
        (les options définies pour le serveur dans main.go)

    - metrics_service.yaml
        → Manifest Service permettant d'exposer le port metrics du controller-manager dans le cluster
        (manifest pour la création d'un service d'exposition des metrics)

operator/config/prometheus/
    - kustomization.yaml
        → Liste les manifests liés à Prometheus Operator
        (éléments à déployer pour prometheus)

    - monitor.yaml
        → Manifest ServiceMonitor permettant à Prometheus Operator de découvrir automatiquement le Service metrics de l'opérateur
        (manifest pour un serviceMonitor incluant le endpoint du service à scraper et la méthode pour le faire (ex. HTTPS avec token))

operator/config/rbac/
    - {roles | rolebindings | serviceaccount}.yaml
        → Ensemble des RBAC générées automatiquement par Kubebuilder pour :
            - le controller
            - le leader election
            - les métriques
            - les accès Kubernetes
```

### Ressources générées

```text
Service
ServiceMonitor
ClusterRole metrics
ClusterRoleBinding metrics
ServiceAccount
```

### Chaîne de supervision

```text
Operator
→ Endpoint /metrics
→ Service
→ ServiceMonitor
→ Prometheus Operator
→ Prometheus
→ Grafana
```

Important :
- Prometheus Operator NE scrape PAS directement les Pods
- il observe les objets `ServiceMonitor`
- puis génère automatiquement la configuration Prometheus correspondante

---

## ⚙️ Actions effectuées

### Exposition du port métriques

Ajout des arguments dans le manager (en fonction de ce qui est défini dans `main.go`) :

```yaml
--metrics-bind-address=:8443
```

Fichier :

```text
operator/config/default/manager_metrics_patch.yaml
```

Explication :
- `controller-runtime` embarque automatiquement un serveur HTTP metrics Prometheus
- ce patch permet de choisir :
    - le port d'écoute
    - l'adresse d'écoute
- ici :
    - port `8443`
    - écoute HTTPS

Le serveur métriques est démarré directement par le manager dans `main.go`.

### Création du Service métriques

Fichier :

```text
operator/config/default/metrics_service.yaml
```

Service créé :

```yaml
kind: Service
port: 8443
targetPort: 8443
```

Explication :
- le Pod expose le port metrics
- mais Prometheus scrape un Service Kubernetes
- le Service fournit donc :
    - une IP stable
    - un endpoint stable
    - une abstraction réseau Kubernetes

### Création du ServiceMonitor

Fichier :

```text
operator/config/prometheus/monitor.yaml
```

Ajout :
- endpoint `/metrics`
- port `https`
- `tlsConfig`
- `bearerTokenFile`

Explication :
- le `ServiceMonitor` est une CRD fournie par Prometheus Operator
- Prometheus Operator observe automatiquement ces objets
- puis génère dynamiquement la configuration Prometheus nécessaire au scraping

Le `bearerTokenFile` permet :
- l'authentification Kubernetes du scrapeur Prometheus

Le `tlsConfig` permet :
- la configuration TLS/HTTPS du scraping

Attention à la configuration définie pour la supervision :
- en production il est important d'utiliser :
    - HTTPS
    - certificats valides
    - validation TLS
    - authentification stricte
- ici :
    - `insecureSkipVerify: true`
    - uniquement pour simplifier le lab

### RBAC métriques

Fichiers :

```text
operator/config/rbac/metrics_reader_role.yaml
operator/config/rbac/metrics_auth_role.yaml
operator/config/rbac/metrics_auth_role_binding.yaml
```

Explication :
- Prometheus doit pouvoir :
    - accéder au endpoint metrics
    - effectuer certaines validations/authentifications Kubernetes

Les RBAC associées autorisent notamment :
- `tokenreviews`
- `subjectaccessreviews`

Ces permissions sont nécessaires lorsque :
- HTTPS sécurisé,
- authentification Kubernetes,
- webhook auth,
- ou metrics server sécurisé sont utilisés.

---

## 🔎 Vérifications

### Vérifier le Service et le ServiceMonitor

```bash
kubectl get svc -n operator-system
kubectl get servicemonitor -n operator-system
```

### Vérifier la target Prometheus

Dans Grafana/Prometheus :

```text
Status > Targets
```

La target attendue doit apparaître :
- UP
- sans erreur TLS/authentification

### Vérifier directement les métriques

```bash
kubectl port-forward svc/operator-controller-manager-metrics-service 8443:8443 -n operator-system
```

Puis :

```text
https://localhost:8443/metrics
```

Le endpoint doit retourner :
- des métriques Prometheus,
- au format texte Prometheus exposition format.

Exemples :
- métriques Go runtime,
- métriques controller-runtime,
- métriques custom éventuelles.

---

## ✅ Bilan

L'opérateur peut désormais exposer automatiquement ses métriques Prometheus :
- temps de réconciliation,
- nombre de réconciliations,
- erreurs,
- métriques runtime Go,
- métriques controller-runtime.

Prometheus Operator découvre automatiquement l'opérateur via le `ServiceMonitor`.

Les métriques deviennent alors directement exploitables dans :
- Prometheus,
- Grafana,
- alerting,
- dashboards,
- supervision du control-plane.