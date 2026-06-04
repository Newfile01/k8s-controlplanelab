# 🔭 Étape #2 - Mise en place d'une pile d'observabilité Prometheus & Grafana (sans opérateur)

Cette étape permet de comprendre les rouages de la collecte de métriques dans Kubernetes avec Prometheus avant l'utilisation d'un opérateur Kubernetes dédié.

On commencera par récupérer les éléments se trouvant dans `sources/`, les copier et les modifier pour arriver aux éléments finaux se trouvant dans `prometheus/`

Architecture finale obtenue :

```text
node-exporter ───────────────┐
                             │
API Server ──────────────────┤
                             │
kube-state-metrics ──────────┤
                             ▼
                        Prometheus ──► Grafana
```

## 🚀 Déploiement manuel de Prometheus

Création du namespace dédié :

```bash
kubectl create namespace monitoring
```

Déploiement des manifestes :

```bash
kubectl apply -f prometheus/
```

Les manifestes utilisés :

```text
prometheus/
├── clusterRole.yaml
├── config-map.yaml
├── prometheus-deployment.yaml
└── prometheus-service.yaml
```

Le déploiement Prometheus crée :

- 1 pod contenant le serveur Prometheus
- 1 service Kubernetes exposant Prometheus
- 1 ConfigMap contenant les jobs de scraping
- 1 ClusterRole et 1 ClusterRoleBinding permettant à Prometheus d'interroger l'API Kubernetes

## 🔐 Correction des API Versions RBAC

Les ressources RBAC utilisent une ancienne API non compatible avec Kubernetes `1.35.1`.

Modifier :

```yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
```

par :

```yaml
apiVersion: rbac.authorization.k8s.io/v1
```

pour :

- `ClusterRole`
- `ClusterRoleBinding`

Vérifications utiles :

```bash
kubectl api-resources | grep role
kubectl api-versions | grep rbac
```

## ⚙️ Fonctionnement du Prometheus manuel

Le serveur Prometheus écoute dans le pod sur :

```text
localhost:9090
```

Le service Kubernetes expose ce port :

```yaml
ports:
  - port: 8090
    targetPort: 9090
```

Accès externe :

```bash
kubectl port-forward \
--address 0.0.0.0 \
svc/prometheus-service \
8090:8090 \
-n monitoring
```

Accès :

```text
http://localhost:8090
```

## 📄 Structure de la ConfigMap Prometheus

Le fichier :

```text
prometheus/config-map.yaml
```

contient les jobs de scraping :

```yaml
data:
  prometheus.yml: |-
```

Structure générale :

```yaml
scrape_configs:
  - job_name: 'nom-du-job'
    static_configs:
      - targets:
        - 'service.namespace.svc:port'
```

Le deployment monte automatiquement cette ConfigMap dans :

```text
/etc/prometheus/prometheus.yml
```

grâce à :

```yaml
volumeMounts:
  - mountPath: /etc/prometheus/
```

Le paramètre :

```yaml
defaultMode: 420
```

correspond aux permissions Linux :

```text
0644
rw-r--r--
```

## 📦 Installation du Node-exporter

Prérequis : Helm

Helm permet de templatiser et déployer des applications Kubernetes complexes.

Installation utilisée ici :

- méthode : `From Script`
- version : `4.1.4`

Documentation :

- https://helm.sh/docs/intro/install/

Ajout du repository :

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts

helm repo update
```

Recherche des charts :

```bash
helm search repo node-exporter
```

Installation :

```bash
helm install node-exporter \
oci://ghcr.io/prometheus-community/charts/prometheus-node-exporter
```

Le chart crée :

- 1 DaemonSet
- 1 pod `node-exporter` par node
- 1 Service ClusterIP
- 1 ServiceAccount

Le DaemonSet garantit :

```text
1 pod node-exporter par node
```

Les métriques Linux exposées concernent notamment :

- CPU
- RAM
- disque
- réseau
- filesystem

## 🎯 Ajout du scraping Node-exporter

Modification de :

```text
prometheus/config-map.yaml
```

Ajout :

```yaml
- job_name: 'node-exporter'
  static_configs:
    - targets:
      - 'node-exporter-prometheus-node-exporter.default.svc:9100'
```

Le target correspond au service Kubernetes :

```text
<service>.<namespace>.svc:<port>
```

Application des modifications :

```bash
kubectl apply -f prometheus/config-map.yaml

kubectl rollout restart deployment/prometheus-deployment \
-n monitoring
```

## ☸️ Ajout du scraping API Server Kubernetes

Prometheus va maintenant interroger directement l'API Kubernetes grâce à :

```yaml
kubernetes_sd_configs
```

Ajout dans :

```text
prometheus/config-map.yaml
```

```yaml
- job_name: 'kubernetes-apiservers'

  kubernetes_sd_configs:
    - role: endpoints

  scheme: https

  tls_config:
    ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt

  bearer_token_file:
    /var/run/secrets/kubernetes.io/serviceaccount/token

  relabel_configs:
    - source_labels:
        [
          __meta_kubernetes_namespace,
          __meta_kubernetes_service_name,
          __meta_kubernetes_endpoint_port_name
        ]
      action: keep
      regex: default;kubernetes;https
```

Cette configuration :

- interroge l'API Kubernetes
- découvre automatiquement les endpoints
- utilise HTTPS/TLS
- utilise le token du ServiceAccount
- filtre uniquement le service `kubernetes`

Architecture obtenue :

```text
Prometheus
    │
    ├── vérifie le certificat du serveur Kubernetes
    │
    └── envoie son token
            │
            ▼
      API Server
            │
            └── vérifie les RBAC
```

Le certificat :

```text
ca.crt
```

sert à authentifier le serveur Kubernetes.

Le token :

```text
bearer_token_file
```

sert à identifier Prometheus auprès de Kubernetes.

Les RBAC définissent :

```text
QUI peut faire QUOI sur QUELLE ressource
```

## 📊 Installation de kube-state-metrics

Ajout du chart :

```bash
helm install kube-state-metrics \
prometheus-community/kube-state-metrics
```

Architecture :

```text
API Server → kube-state-metrics → Prometheus
```

`kube-state-metrics` transforme l'état des objets Kubernetes en métriques Prometheus :

- Pods
- Deployments
- ReplicaSets
- Nodes
- DaemonSets
- StatefulSets
- Jobs
- PVC
- etc.

Contrairement à `node-exporter` :

```text
node-exporter → observe Linux
kube-state-metrics → observe Kubernetes
```

## 🎯 Ajout du scraping kube-state-metrics

Ajout dans :

```text
prometheus/config-map.yaml
```

```yaml
- job_name: 'kube-state-metrics'
  static_configs:
    - targets:
      - 'kube-state-metrics.default.svc:8080'
```

Application des modifications :

```bash
kubectl apply -f prometheus/config-map.yaml

kubectl rollout restart deployment/prometheus-deployment \
-n monitoring
```

## 📈 Installation de Grafana

Ajout du repository :

```bash
helm repo add grafana-community \
https://grafana-community.github.io/helm-charts

helm repo update
```

Installation :

```bash
helm install grafana-dashboard \
grafana-community/grafana
```

Le chart crée notamment :

- 1 Deployment Grafana
- 1 Service Kubernetes
- 1 Secret contenant le mot de passe admin

Récupération du mot de passe :

```bash
kubectl get secret grafana-dashboard \
-o jsonpath="{.data.admin-password}" \
| base64 -d
```

Exposition Grafana :

```bash
kubectl port-forward \
svc/grafana-dashboard \
3000:80
```

Accès :

```text
http://localhost:3000
```

Utilisateur :

```text
admin
```

## 🔗 Ajout de Prometheus comme datasource Grafana

Dans Grafana :

```text
Connections → Data Sources → Add data source
```

Choisir :

```text
Prometheus
```

Datasource :

```text
http://prometheus-service.monitoring.svc:8090
```

Le format DNS Kubernetes utilisé est :

```text
<service>.<namespace>.svc
```

Grafana interroge uniquement Prometheus :

```text
Exporters → Prometheus → Grafana
```

Grafana ne scrape pas directement les exporters.

## 📉 Dashboard Kubernetes

Import du dashboard :

```text
Dashboard ID : 7249
```

Ce dashboard permet de visualiser :

- Nodes
- CPU
- RAM
- Pods
- Control-plane
- métriques Kubernetes
- métriques Linux
