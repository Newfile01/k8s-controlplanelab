# Configuration manuelle de Prometheus et des composants de scraping

Entre chaque MàJ de la `config-map.yaml` il faudra exécuter les commandes suivantes (depuis le dossier contenant le sous-dossier `prometheus/`)

```text
(le nom par défaut est) 02-prometheus-manuel/
├── README.md
📍 (vous devez être ici avec votre terminal)
└── prometheus/
    ├── clusterRole.yaml
    ├── config-map.yaml
    ├── prometheus-deployment.yaml
    ├── prometheus-service.yaml
    └── kube-state-metrics/
        └── (installation Helm / manifests éventuels)
```

```bash
kubectl apply -f prometheus/config-map.yaml
# configmap/prometheus-server-conf configured
kubectl rollout restart deployment/prometheus-deployment \
-n monitoring
# deployment.apps/prometheus-deployment restarted
```

✅ Permettra de prendre en compte les modifications apportées aux manifests

---

## PROMETHEUS

```bash
# Recherche des APIs installées sur notre cluster par mot clé
kubectl api-resources | grep role
# clusterrolebindings                              rbac.authorization.k8s.io/v1      false        ClusterRoleBinding
# clusterroles                                     rbac.authorization.k8s.io/v1      false        ClusterRole
# rolebindings                                     rbac.authorization.k8s.io/v1      true         RoleBinding
# roles                                            rbac.authorization.k8s.io/v1      true         Role

# Recherche version de l'API correspondant aux RBACs
kubectl api-versions | grep rbac
# rbac.authorization.k8s.io/v1

# Vérification de la création du Role Prometheus
kubectl get clusterrole prometheus
# NAME         CREATED AT
# prometheus   2026-06-02T23:35:30Z

# Et du RoleBinding => association prometheus à un compte de service (ici default) sur un namespace donné
kubectl get clusterrolebinding prometheus
# NAME         ROLE                     AGE
# prometheus   ClusterRole/prometheus   5m22s

kubectl get all -n monitoring
# NAME                                       READY   STATUS    RESTARTS   AGE
# pod/prometheus-deployment-c94dcc78-cjv4x   1/1     Running   0          15m

# NAME                         TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
# service/prometheus-service   NodePort   10.105.112.58   <none>        8090:30000/TCP   15m

# NAME                                    READY   UP-TO-DATE   AVAILABLE   AGE
# deployment.apps/prometheus-deployment   1/1     1            1           15m

# NAME                                             DESIRED   CURRENT   READY   AGE
# replicaset.apps/prometheus-deployment-c94dcc78   1         1         1       15m



############
# ACCES
# Port-forward
kubectl port-forward \
svc/prometheus-service \
8090:8090 \
-n monitoring
# URL
# http://localhost:8090
```

---

## NODE-EXPORTER

```bash
#################
# NODE-EXPORTER & HELM
#################

# Ajout du repo helm
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
# Trouver les charts correspondant
helm search repo node-exporter
# Installation
helm install node-exporter oci://ghcr.io/prometheus-community/charts/prometheus-node-exporter
# "prometheus-community" already exists with the same configuration, skipping
# Hang tight while we grab the latest from your chart repositories...
# ...Successfully got an update from the "grafana-community" chart repository
# ...Successfully got an update from the "prometheus-community" chart repository
# ...Successfully got an update from the "grafana" chart repository
# ...Successfully got an update from the "stable" chart repository
# ...Successfully got an update from the "bitnami" chart repository
# Update Complete. ⎈Happy Helming!⎈
# NAME                                            CHART VERSION   APP VERSION     DESCRIPTION
# bitnami/node-exporter                           4.5.19          1.9.1           Prometheus exporter for hardware and OS metrics...
# prometheus-community/prometheus-node-exporter   4.55.0          1.11.1          A Helm chart for prometheus node-exporter
# stable/prometheus-node-exporter                 1.11.2          1.0.1           DEPRECATED A Helm chart for prometheus node-exp...
# Pulled: ghcr.io/prometheus-community/charts/prometheus-node-exporter:4.55.0
# Digest: sha256:a0e0d89c5a4036722282dc94c91cea1309e53b56a5b48276f48c280d617dafab
# NAME: node-exporter
# LAST DEPLOYED: Wed Jun  3 12:04:23 2026
# NAMESPACE: default
# STATUS: deployed
# REVISION: 1
# DESCRIPTION: Install complete
# TEST SUITE: None
# NOTES:
# 1. Get the application URL by running these commands:
#   export POD_NAME=$(kubectl get pods --namespace default -l "app.kubernetes.io/name=prometheus-node-exporter,app.kubernetes.io/instance=node-exporter" -o jsonpath="{.items[0].metadata.name}")
#   echo "Visit http://127.0.0.1:9100 to use your application"
#   kubectl port-forward --namespace default $POD_NAME 9100

# On remarque la création d'un pod sur chaque noeud (nécessaire pour récupérer les métriques propres à chacun)
# Création d'un service permettant à Prometheus de dialoguer avec (d'atteindre les endpoints que présente node-exporter)
# Création d'un démon sur chaque noeud
kc get all
# NAME                                               READY   STATUS    RESTARTS   AGE
# pod/node-exporter-prometheus-node-exporter-4dl7g   1/1     Running   0          7m6s
# pod/node-exporter-prometheus-node-exporter-9b589   1/1     Running   0          7m6s
# pod/node-exporter-prometheus-node-exporter-df8sb   1/1     Running   0          7m6s
# pod/node-exporter-prometheus-node-exporter-vbqfd   1/1     Running   0          7m6s

# NAME                                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
# service/kubernetes                               ClusterIP   10.96.0.1       <none>        443/TCP    21h
# service/node-exporter-prometheus-node-exporter   ClusterIP   10.102.138.68   <none>        9100/TCP   7m6s

# NAME                                                    DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR            AGE
# daemonset.apps/node-exporter-prometheus-node-exporter   4         4         4       4            4           kubernetes.io/os=linux   7m6s

# Ensemble des endpoints exposés
kc get endpoints
# Warning: v1 Endpoints is deprecated in v1.33+; use discovery.k8s.io/v1 EndpointSlice
# NAME                                     ENDPOINTS                                                           AGE
# kubernetes                               192.168.67.2:8443                                                   22h
# node-exporter-prometheus-node-exporter   192.168.67.2:9100,192.168.67.3:9100,192.168.67.4:9100 + 1 more...   48m



############
# ACCES
# Port-forward
kc port-forward node-exporter-prometheus-node-exporter-4dl7g 9100:9100
# URL
# http://localhost:9100/metrics
```

---

## API-SERVER METRICS

Cette partie concerne l'implantation d'un point de scraping pour l'api-server dans Prometheus. Permettra d'observer les interactions de Kubernetes en action.
Tout se passe dans la `config-map.yaml`

```yaml
### Scraping de l'API-SERVER
      - job_name: 'kubernetes-apiservers'
      # lancement du SD (Service Discovery) pour découvrir les Endpoints Kubernetes
        kubernetes_sd_configs:
          - role: endpoints
      # Dialogue avec l'api-server forcément en HTTPS
        scheme: https
      # On indique le certificat pour le dialogue
        tls_config:
          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      # Ainsi qu'un token du ServiceAccount qui est monté directement dans le pod
        bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      # On filtre pour aller chercher namespace "default", service "kubernetes", port "https" (=443)
      # Ce qui correspond à l'API-SERVER
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

### 🔍 Bien comprendre la relation entre Services et Endpoints

Les Endpoints représentent les IP:Port réels vers lesquels un Service peut router le trafic (généralement des Pods). Ils sont unitaires et mis à jour régulièrement par le Endpoint Controller (qui se trouve dans le kube-controller-manager). Les services, représenté sous forme de nom de domaine suivant le schéma DNS de Kubernetes, s'appuyent sur ce controller pour mettre à jour les Endpoints vers lesquels ils pointent sur la base des "selector" définis dans leur manifestes.

## 📋 RBAC (Role Based Access Control) pour Prometheus

Les RBAC servent à définir "QUI peut faire QUOI sur QUELLE ressource ?" dans un cluster. Ici on défini au travers de `clusterRole.yaml` un rôle pour Prometheus (ayant accès à différents éléments via certaine méthodes : GET, LIST, WATCH) et on le l'associe ("bind") à un serviceAccount. Ici c'est au serviceAccount "default" du namespace "monitoring" (dont fait partie Prometheus)

Il est possible d'observer les autorisation d'un rôle ainsi que le ou les serviceAccounts auxquels il est associé comme suit :

```bash
kubectl describe clusterrole prometheus
# Name:         prometheus
# Labels:       <none>
# Annotations:  <none>
# PolicyRule:
#   Resources                    Non-Resource URLs  Resource Names  Verbs
#   ---------                    -----------------  --------------  -----
#   endpoints                    []                 []              [get list watch]
#   nodes/proxy                  []                 []              [get list watch]
#   nodes                        []                 []              [get list watch]
#   pods                         []                 []              [get list watch]
#   services                     []                 []              [get list watch]
#   ingresses.networking.k8s.io  []                 []              [get list watch]
#                                [/metrics]         []              [get]

kubectl describe clusterrolebinding prometheus
# Name:         prometheus
# Labels:       <none>
# Annotations:  <none>
# Role:
#   Kind:  ClusterRole
#   Name:  prometheus
# Subjects:
#   Kind            Name     Namespace
#   ----            ----     ---------
#   ServiceAccount  default  monitoring

kubectl auth can-i list endpoints \
--as=system:serviceaccount:monitoring:default
# yes

#  Cette dernière commande permet de tester des méthodes en précisant le serviceAccount utilisé afin de vérifier les autorisations
```

### 🔀 Note sur le scraping de l'API-SERVER

Pour récupérer l'ensemble des métriques sur l'API-SERVER il faut pouvoir interroger son service. 

Le `kubekubernetes_sd_configs` permet d'utiliser le Service Discovery (SD) de Kubernetes pour aller découvrir l'ensemble des endpoints de Kubernetes suivant un schéma spécifique. Ici le schéma doit employer HTTPS et donc un certificat. Pour dialoguer avec un endpoint il faudra en plus un token que l'on vient récupérer sur le serviceAccount associé à Prometheus (aux pods qu'il génère) et donc par l'intermédiaire de son ClusterRole et ClusterRoleBinding. La partie `relabels` sert quant à elle à filtre sur l'ensemble des endpoints présentés pour aller chercher le service se trouvant dans le namespace "default" et portant le nom "kubernetes" en ne s'intéressant qu'au port "https" (càd 443 ici redirigé visiblement vers le 8443). 

Ce qui correspond finalement à pointer sur le service qui pointe lui-même sur le endpoint de l'api-server



### 🔐 Certificat, Token et RBAC Kubernetes


| Élément                   | Description                                                                                                                                                                              |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Contexte                  | `Prometheus = client` → `API Server Kubernetes = serveur HTTPS`                                                                                                                          |
| Séquence                  | `Prometheus → vérifie le certificat du serveur → envoie son token → API Server → vérifie les RBAC`                                                                                       |
| Certificat TLS (`ca.crt`) | Authentifie le serveur Kubernetes. Prometheus compare le certificat présenté par l'API Server avec la CA Kubernetes présente dans `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt` |
| TLS                       | Chiffre et sécurise la connexion HTTPS                                                                                                                                                   |
| Bearer Token              | Identifie le client Kubernetes : `Prometheus → ServiceAccount monitoring/default`                                                                                                        |
| RBAC                      | Autorise/refuse les actions : `ServiceAccount → ClusterRoleBinding → ClusterRole → permissions`                                                                                          |
| Exemple RBAC              | `list pods` | `watch endpoints` | `get services`                                                                                                                                         |
| Refus                     | `403 Forbidden`                                                                                                                                                                          |


---

## KUBE-STATE-METRICS

Kube-state-meetrics permet d'observer l'enemble des objet de Kubernetes

```bash
# KUBE-STATE-METRICS
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
# "prometheus-community" already exists with the same configuration, skipping
# Hang tight while we grab the latest from your chart repositories...
# ...Successfully got an update from the "grafana-community" chart repository
# ...Successfully got an update from the "grafana" chart repository
# ...Successfully got an update from the "prometheus-community" chart repository
# ...Successfully got an update from the "stable" chart repository
# ...Successfully got an update from the "bitnami" chart repository
# Update Complete. ⎈Happy Helming!⎈

helm install kube-state-metrics prometheus-community/charts/kube-state-metrics
# Error: INSTALLATION FAILED: chart "charts/kube-state-metrics" matching  not found in prometheus-community index. (try 'helm repo update'): no chart name found

helm install kube-state-metrics oci://ghcr.io/prometheus-community/charts/kube-state-metrics
# Pulled: ghcr.io/prometheus-community/charts/kube-state-metrics:7.4.0
# Digest: sha256:48ebaf7dfbe9df5a1e6af5ab78992de3026dcef647168c38088554796a01b932
# NAME: kube-state-metrics
# LAST DEPLOYED: Wed Jun  3 16:31:19 2026
# NAMESPACE: default
# STATUS: deployed
# REVISION: 1
# DESCRIPTION: Install complete
# TEST SUITE: None
# NOTES:
# kube-state-metrics is a simple service that listens to the Kubernetes API server and generates metrics about the state of the objects.
# The exposed metrics can be found here:
# https://github.com/kubernetes/kube-state-metrics/blob/master/docs/README.md#exposed-metrics

# The metrics are exported on the HTTP endpoint /metrics on the listening port.
# In your case, kube-state-metrics.default.svc.cluster.local:8080/metrics

# They are served either as plaintext or protobuf depending on the Accept header.
# They are designed to be consumed either by Prometheus itself or by a scraper that is compatible with scraping a Prometheus client endpoint.

kc get pods
# NAME                                           READY   STATUS    RESTARTS   AGE
# kube-state-metrics-7d4775494-8rtxf             1/1     Running   0          56s
# node-exporter-prometheus-node-exporter-4dl7g   1/1     Running   0          4h27m
# node-exporter-prometheus-node-exporter-9b589   1/1     Running   0          4h27m
# node-exporter-prometheus-node-exporter-df8sb   1/1     Running   0          4h27m
# node-exporter-prometheus-node-exporter-vbqfd   1/1     Running   0          4h27m

kc get all
# NAME                                               READY   STATUS    RESTARTS   AGE
# pod/kube-state-metrics-7d4775494-8rtxf             1/1     Running   0          69s
# pod/node-exporter-prometheus-node-exporter-4dl7g   1/1     Running   0          4h28m
# pod/node-exporter-prometheus-node-exporter-9b589   1/1     Running   0          4h28m
# pod/node-exporter-prometheus-node-exporter-df8sb   1/1     Running   0          4h28m
# pod/node-exporter-prometheus-node-exporter-vbqfd   1/1     Running   0          4h28m

# NAME                                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
# service/kube-state-metrics                       ClusterIP   10.109.168.42   <none>        8080/TCP   69s
# service/kubernetes                               ClusterIP   10.96.0.1       <none>        443/TCP    25h
# service/node-exporter-prometheus-node-exporter   ClusterIP   10.102.138.68   <none>        9100/TCP   4h28m

# NAME                                                    DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR            AGE
# daemonset.apps/node-exporter-prometheus-node-exporter   4         4         4       4            4           kubernetes.io/os=linux   4h28m

# NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
# deployment.apps/kube-state-metrics   1/1     1            1           69s

# NAME                                           DESIRED   CURRENT   READY   AGE
# replicaset.apps/kube-state-metrics-7d4775494   1         1         1       69s



############
# ACCES
# Port-forward
kubectl port-forward svc/kube-state-metrics 8080:8080
# URL
# http://localhost:8080/metrics
```

---

## GRAFANA

Ajout de Grafana pour permettre la visualisation avancée des données Prometheus

```bash
helm repo add grafana-community https://grafana-community.github.io/helm-charts
helm repo update
helm install grafana-dashboard grafana-community/grafana

# Récupération du mot de passe admin
kubectl get secret --namespace default grafana-dashboard -o jsonpath="{.data.admin-password}" | base64 --decode ; echo

# Vérif installation
kc get all -A |grep grafana
# default       pod/grafana-dashboard-6757bc8895-7xmz5             1/1     Running   0               2m
# default       service/grafana-dashboard                        ClusterIP   10.103.138.32   <none>        80/TCP                   2m
# default       deployment.apps/grafana-dashboard       1/1     1            1           2m
# default       replicaset.apps/grafana-dashboard-6757bc8895       1         1         1       2m

###
# ACCES
# Port-Forward
kubectl port-forward svc/grafana-dashboard 3000:80
# URL
```

