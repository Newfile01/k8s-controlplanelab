# Passage de la stack manuelle à la stack automatisée

On cherche cette fois-ci à se servir d'une solution "clé-en-main" permettant de gérer automatiquement :
- Les configurations de scrape (auparavant inscrite dans la `config-map.yaml`)
- La découverte des éléments à surveiller (avec des CRDs MàJ par l'opérateur Prometheus : ServiceMonitor & PodMonitor)
- Pour préparer l'ensemble des services exposés pour Prometheus et Grafana
- Pour avoir une première configuration de Grafana

## Installation de la stack

```bash
# Ajout du repo qui contient le Helm chart
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Création d'un namespace dédié à la nouvelle stack
kubectl create namespace monitoring-stack

# "prometheus-community" already exists with the same configuration, skipping
# Hang tight while we grab the latest from your chart repositories...
# ...Successfully got an update from the "grafana-community" chart repository
# ...Successfully got an update from the "grafana" chart repository
# ...Successfully got an update from the "prometheus-community" chart repository
# ...Successfully got an update from the "stable" chart repository
# ...Successfully got an update from the "bitnami" chart repository
# Update Complete. ⎈Happy Helming!⎈
# namespace/monitoring-stack created

# Installation du helm chart
helm install monitoring-stack prometheus-community/kube-prometheus-stack -n monitoring-stack

# NAME: monitoring-stack
# LAST DEPLOYED: Wed Jun  3 19:37:18 2026
# NAMESPACE: monitoring-stack
# STATUS: deployed
# REVISION: 1
# DESCRIPTION: Install complete
# TEST SUITE: None
# NOTES:
# kube-prometheus-stack has been installed. Check its status by running:
#   kubectl --namespace monitoring-stack get pods -l "release=monitoring-stack"

# Get Grafana 'admin' user password by running:

#   kubectl --namespace monitoring-stack get secrets monitoring-stack-grafana -o jsonpath="{.data.admin-password}" | base64 -d ; echo

# Access Grafana local instance:

#   export POD_NAME=$(kubectl --namespace monitoring-stack get pod -l "app.kubernetes.io/name=grafana,app.kubernetes.io/instance=monitoring-stack" -oname)
#   kubectl --namespace monitoring-stack port-forward $POD_NAME 3000

# Get your grafana admin user password by running:

#   kubectl get secret --namespace monitoring-stack -l app.kubernetes.io/component=admin-secret -o jsonpath="{.items[0].data.admin-password}" | base64 --decode ; echo


# Visit https://github.com/prometheus-operator/kube-prometheus for instructions on how to create & configure Alertmanager and Prometheus instances using the Operator.
```

Etat du cluster avant installation :

```bash
kubectl get all -A
# NAMESPACE     NAME                                               READY   STATUS    RESTARTS        AGE
# default       pod/grafana-dashboard-6757bc8895-7xmz5             1/1     Running   0               99m
# default       pod/kube-state-metrics-7d4775494-8rtxf             1/1     Running   0               3h4m
# default       pod/nginx-b6485fcbb-fppkp                          1/1     Running   0               29m
# default       pod/nginx-b6485fcbb-mjjlg                          1/1     Running   0               29m
# default       pod/nginx-b6485fcbb-zdv5k                          1/1     Running   0               29m
# default       pod/node-exporter-prometheus-node-exporter-4dl7g   1/1     Running   0               7h30m
# default       pod/node-exporter-prometheus-node-exporter-9b589   1/1     Running   0               7h30m
# default       pod/node-exporter-prometheus-node-exporter-df8sb   1/1     Running   0               7h30m
# default       pod/node-exporter-prometheus-node-exporter-vbqfd   1/1     Running   0               7h30m
# kube-system   pod/coredns-7d764666f9-f7kgh                       1/1     Running   1 (16h ago)     28h
# kube-system   pod/etcd-control-plane-lab                         1/1     Running   1 (16h ago)     28h
# kube-system   pod/kindnet-kkvdp                                  1/1     Running   1 (16h ago)     28h
# kube-system   pod/kindnet-lz7s5                                  1/1     Running   1 (16h ago)     28h
# kube-system   pod/kindnet-mj6mj                                  1/1     Running   1 (16h ago)     28h
# kube-system   pod/kindnet-s578z                                  1/1     Running   1 (16h ago)     28h
# kube-system   pod/kube-apiserver-control-plane-lab               1/1     Running   1 (16h ago)     28h
# kube-system   pod/kube-controller-manager-control-plane-lab      1/1     Running   1 (16h ago)     28h
# kube-system   pod/kube-proxy-9k9t7                               1/1     Running   1 (16h ago)     28h
# kube-system   pod/kube-proxy-c4l64                               1/1     Running   1 (16h ago)     28h
# kube-system   pod/kube-proxy-jdr46                               1/1     Running   1 (16h ago)     28h
# kube-system   pod/kube-proxy-wfdl2                               1/1     Running   1 (16h ago)     28h
# kube-system   pod/kube-scheduler-control-plane-lab               1/1     Running   1 (16h ago)     28h
# kube-system   pod/metrics-server-9d74bb658-29cj4                 1/1     Running   2 (7h30m ago)   20h
# kube-system   pod/storage-provisioner                            1/1     Running   2 (7h31m ago)   28h
# monitoring    pod/prometheus-deployment-68f7f64854-jlnkq         1/1     Running   0               31m

# NAMESPACE     NAME                                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                  AGE
# default       service/grafana-dashboard                        ClusterIP   10.103.138.32   <none>        80/TCP                   99m
# default       service/kube-state-metrics                       ClusterIP   10.109.168.42   <none>        8080/TCP                 3h4m
# default       service/kubernetes                               ClusterIP   10.96.0.1       <none>        443/TCP                  28h
# default       service/node-exporter-prometheus-node-exporter   ClusterIP   10.102.138.68   <none>        9100/TCP                 7h30m
# kube-system   service/kube-dns                                 ClusterIP   10.96.0.10      <none>        53/UDP,53/TCP,9153/TCP   28h
# kube-system   service/metrics-server                           ClusterIP   10.98.169.79    <none>        443/TCP                  20h
# monitoring    service/prometheus-service                       NodePort    10.105.112.58   <none>        8090:30000/TCP           18h

# NAMESPACE     NAME                                                    DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR            AGE
# default       daemonset.apps/node-exporter-prometheus-node-exporter   4         4         4       4            4           kubernetes.io/os=linux   7h30m
# kube-system   daemonset.apps/kindnet                                  4         4         4       4            4           <none>                   28h
# kube-system   daemonset.apps/kube-proxy                               4         4         4       4            4           kubernetes.io/os=linux   28h

# NAMESPACE     NAME                                    READY   UP-TO-DATE   AVAILABLE   AGE
# default       deployment.apps/grafana-dashboard       1/1     1            1           99m
# default       deployment.apps/kube-state-metrics      1/1     1            1           3h4m
# default       deployment.apps/nginx                   3/3     3            3           29m
# kube-system   deployment.apps/coredns                 1/1     1            1           28h
# kube-system   deployment.apps/metrics-server          1/1     1            1           20h
# monitoring    deployment.apps/prometheus-deployment   1/1     1            1           18h

# NAMESPACE     NAME                                               DESIRED   CURRENT   READY   AGE
# default       replicaset.apps/grafana-dashboard-6757bc8895       1         1         1       99m
# default       replicaset.apps/kube-state-metrics-7d4775494       1         1         1       3h4m
# default       replicaset.apps/nginx-b6485fcbb                    3         3         3       29m
# kube-system   replicaset.apps/coredns-7d764666f9                 1         1         1       28h
# kube-system   replicaset.apps/metrics-server-9d74bb658           1         1         1       20h
# monitoring    replicaset.apps/prometheus-deployment-68f7f64854   1         1         1       31m
# monitoring    replicaset.apps/prometheus-deployment-6c9b7f86c5   0         0         0       5h47m
# monitoring    replicaset.apps/prometheus-deployment-845768596b   0         0         0       163m
# monitoring    replicaset.apps/prometheus-deployment-bccd966bc    0         0         0       6h46m
# monitoring    replicaset.apps/prometheus-deployment-c94dcc78     0         0         0       18h
```

Après installation :

```bash
kc get all -A
# NAMESPACE          NAME                                                         READY   STATUS    RESTARTS        AGE
# default            pod/grafana-dashboard-6757bc8895-7xmz5                       1/1     Running   0               102m
# default            pod/kube-state-metrics-7d4775494-8rtxf                       1/1     Running   0               3h7m
# default            pod/nginx-b6485fcbb-fppkp                                    1/1     Running   0               33m
# default            pod/nginx-b6485fcbb-mjjlg                                    1/1     Running   0               33m
# default            pod/nginx-b6485fcbb-zdv5k                                    1/1     Running   0               33m
# default            pod/node-exporter-prometheus-node-exporter-4dl7g             1/1     Running   0               7h34m
# default            pod/node-exporter-prometheus-node-exporter-9b589             1/1     Running   0               7h34m
# default            pod/node-exporter-prometheus-node-exporter-df8sb             1/1     Running   0               7h34m
# default            pod/node-exporter-prometheus-node-exporter-vbqfd             1/1     Running   0               7h34m
# kube-system        pod/coredns-7d764666f9-f7kgh                                 1/1     Running   1 (16h ago)     29h
# kube-system        pod/etcd-control-plane-lab                                   1/1     Running   1 (16h ago)     29h
# kube-system        pod/kindnet-kkvdp                                            1/1     Running   1 (16h ago)     28h
# kube-system        pod/kindnet-lz7s5                                            1/1     Running   1 (16h ago)     29h
# kube-system        pod/kindnet-mj6mj                                            1/1     Running   1 (16h ago)     29h
# kube-system        pod/kindnet-s578z                                            1/1     Running   1 (16h ago)     28h
# kube-system        pod/kube-apiserver-control-plane-lab                         1/1     Running   1 (16h ago)     29h
# kube-system        pod/kube-controller-manager-control-plane-lab                1/1     Running   1 (16h ago)     29h
# kube-system        pod/kube-proxy-9k9t7                                         1/1     Running   1 (16h ago)     28h
# kube-system        pod/kube-proxy-c4l64                                         1/1     Running   1 (16h ago)     29h
# kube-system        pod/kube-proxy-jdr46                                         1/1     Running   1 (16h ago)     28h
# kube-system        pod/kube-proxy-wfdl2                                         1/1     Running   1 (16h ago)     29h
# kube-system        pod/kube-scheduler-control-plane-lab                         1/1     Running   1 (16h ago)     29h
# kube-system        pod/metrics-server-9d74bb658-29cj4                           1/1     Running   2 (7h34m ago)   20h
# kube-system        pod/storage-provisioner                                      1/1     Running   2 (7h34m ago)   29h
# monitoring-stack   pod/alertmanager-monitoring-stack-kube-prom-alertmanager-0   2/2     Running   0               59s
# monitoring-stack   pod/monitoring-stack-grafana-745c96f497-mjppd                3/3     Running   0               68s
# monitoring-stack   pod/monitoring-stack-kube-prom-operator-766c56bbb4-hdd7z     1/1     Running   0               68s
# monitoring-stack   pod/monitoring-stack-kube-state-metrics-5657b445cb-mrcvs     1/1     Running   0               68s
# monitoring-stack   pod/monitoring-stack-prometheus-node-exporter-57x6l          0/1     Pending   0               68s
# monitoring-stack   pod/monitoring-stack-prometheus-node-exporter-dl2st          0/1     Pending   0               68s
# monitoring-stack   pod/monitoring-stack-prometheus-node-exporter-msfv8          0/1     Pending   0               68s
# monitoring-stack   pod/monitoring-stack-prometheus-node-exporter-nbs2j          0/1     Pending   0               68s
# monitoring-stack   pod/prometheus-monitoring-stack-kube-prom-prometheus-0       2/2     Running   0               59s
# monitoring         pod/prometheus-deployment-68f7f64854-jlnkq                   1/1     Running   0               34m

# NAMESPACE          NAME                                                         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                        AGE
# default            service/grafana-dashboard                                    ClusterIP   10.103.138.32    <none>        80/TCP                         102m
# default            service/kube-state-metrics                                   ClusterIP   10.109.168.42    <none>        8080/TCP                       3h7m
# default            service/kubernetes                                           ClusterIP   10.96.0.1        <none>        443/TCP                        29h
# default            service/node-exporter-prometheus-node-exporter               ClusterIP   10.102.138.68    <none>        9100/TCP                       7h34m
# kube-system        service/kube-dns                                             ClusterIP   10.96.0.10       <none>        53/UDP,53/TCP,9153/TCP         29h
# kube-system        service/metrics-server                                       ClusterIP   10.98.169.79     <none>        443/TCP                        20h
# kube-system        service/monitoring-stack-kube-prom-coredns                   ClusterIP   None             <none>        9153/TCP                       68s
# kube-system        service/monitoring-stack-kube-prom-kube-controller-manager   ClusterIP   None             <none>        10257/TCP                      68s
# kube-system        service/monitoring-stack-kube-prom-kube-etcd                 ClusterIP   None             <none>        2381/TCP                       68s
# kube-system        service/monitoring-stack-kube-prom-kube-proxy                ClusterIP   None             <none>        10249/TCP                      68s
# kube-system        service/monitoring-stack-kube-prom-kube-scheduler            ClusterIP   None             <none>        10259/TCP                      68s
# kube-system        service/monitoring-stack-kube-prom-kubelet                   ClusterIP   None             <none>        10250/TCP,4194/TCP,10255/TCP   60s
# monitoring-stack   service/alertmanager-operated                                ClusterIP   None             <none>        9093/TCP,9094/TCP,9094/UDP     59s
# monitoring-stack   service/monitoring-stack-grafana                             ClusterIP   10.107.245.223   <none>        80/TCP                         68s
# monitoring-stack   service/monitoring-stack-kube-prom-alertmanager              ClusterIP   10.99.107.10     <none>        9093/TCP,8080/TCP              68s
# monitoring-stack   service/monitoring-stack-kube-prom-operator                  ClusterIP   10.107.131.193   <none>        443/TCP                        68s
# monitoring-stack   service/monitoring-stack-kube-prom-prometheus                ClusterIP   10.110.221.42    <none>        9090/TCP,8080/TCP              68s
# monitoring-stack   service/monitoring-stack-kube-state-metrics                  ClusterIP   10.109.213.98    <none>        8080/TCP                       68s
# monitoring-stack   service/monitoring-stack-prometheus-node-exporter            ClusterIP   10.97.73.142     <none>        9100/TCP                       68s
# monitoring-stack   service/prometheus-operated                                  ClusterIP   None             <none>        9090/TCP                       59s
# monitoring         service/prometheus-service                                   NodePort    10.105.112.58    <none>        8090:30000/TCP                 18h

# NAMESPACE          NAME                                                       DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR            AGE
# default            daemonset.apps/node-exporter-prometheus-node-exporter      4         4         4       4            4           kubernetes.io/os=linux   7h34m
# kube-system        daemonset.apps/kindnet                                     4         4         4       4            4           <none>                   29h
# kube-system        daemonset.apps/kube-proxy                                  4         4         4       4            4           kubernetes.io/os=linux   29h
# monitoring-stack   daemonset.apps/monitoring-stack-prometheus-node-exporter   4         4         0       4            0           kubernetes.io/os=linux   68s

# NAMESPACE          NAME                                                  READY   UP-TO-DATE   AVAILABLE   AGE
# default            deployment.apps/grafana-dashboard                     1/1     1            1           102m
# default            deployment.apps/kube-state-metrics                    1/1     1            1           3h7m
# default            deployment.apps/nginx                                 3/3     3            3           33m
# kube-system        deployment.apps/coredns                               1/1     1            1           29h
# kube-system        deployment.apps/metrics-server                        1/1     1            1           20h
# monitoring-stack   deployment.apps/monitoring-stack-grafana              1/1     1            1           68s
# monitoring-stack   deployment.apps/monitoring-stack-kube-prom-operator   1/1     1            1           68s
# monitoring-stack   deployment.apps/monitoring-stack-kube-state-metrics   1/1     1            1           68s
# monitoring         deployment.apps/prometheus-deployment                 1/1     1            1           18h

# NAMESPACE          NAME                                                             DESIRED   CURRENT   READY   AGE
# default            replicaset.apps/grafana-dashboard-6757bc8895                     1         1         1       102m
# default            replicaset.apps/kube-state-metrics-7d4775494                     1         1         1       3h7m
# default            replicaset.apps/nginx-b6485fcbb                                  3         3         3       33m
# kube-system        replicaset.apps/coredns-7d764666f9                               1         1         1       29h
# kube-system        replicaset.apps/metrics-server-9d74bb658                         1         1         1       20h
# monitoring-stack   replicaset.apps/monitoring-stack-grafana-745c96f497              1         1         1       68s
# monitoring-stack   replicaset.apps/monitoring-stack-kube-prom-operator-766c56bbb4   1         1         1       68s
# monitoring-stack   replicaset.apps/monitoring-stack-kube-state-metrics-5657b445cb   1         1         1       68s
# monitoring         replicaset.apps/prometheus-deployment-68f7f64854                 1         1         1       34m
# monitoring         replicaset.apps/prometheus-deployment-6c9b7f86c5                 0         0         0       5h50m
# monitoring         replicaset.apps/prometheus-deployment-845768596b                 0         0         0       166m
# monitoring         replicaset.apps/prometheus-deployment-bccd966bc                  0         0         0       6h49m
# monitoring         replicaset.apps/prometheus-deployment-c94dcc78                   0         0         0       18h

# NAMESPACE          NAME                                                                    READY   AGE
# monitoring-stack   statefulset.apps/alertmanager-monitoring-stack-kube-prom-alertmanager   1/1     59s
# monitoring-stack   statefulset.apps/prometheus-monitoring-stack-kube-prom-prometheus       1/1     59s
```

### Observer les CRDs installées

```bash
kubectl get crds | grep monitoring.coreos.com
# alertmanagerconfigs.monitoring.coreos.com   2026-06-03T17:37:17Z
# alertmanagers.monitoring.coreos.com         2026-06-03T17:37:17Z
# podmonitors.monitoring.coreos.com           2026-06-03T17:37:17Z
# probes.monitoring.coreos.com                2026-06-03T17:37:18Z
# prometheusagents.monitoring.coreos.com      2026-06-03T17:37:18Z
# prometheuses.monitoring.coreos.com          2026-06-03T17:37:18Z
# prometheusrules.monitoring.coreos.com       2026-06-03T17:37:18Z
# scrapeconfigs.monitoring.coreos.com         2026-06-03T17:37:18Z
# servicemonitors.monitoring.coreos.com       2026-06-03T17:37:18Z
# thanosrulers.monitoring.coreos.com          2026-06-03T17:37:18Z
```

### Observer l'ensemble des ressources

```bash
kubectl get all -A | grep monitoring-stack
# monitoring-stack   pod/alertmanager-monitoring-stack-kube-prom-alertmanager-0   2/2     Running   0               5m59s
# monitoring-stack   pod/monitoring-stack-grafana-745c96f497-mjppd                3/3     Running   0               6m8s
# monitoring-stack   pod/monitoring-stack-kube-prom-operator-766c56bbb4-hdd7z     1/1     Running   0               6m8s
# monitoring-stack   pod/monitoring-stack-kube-state-metrics-5657b445cb-mrcvs     1/1     Running   0               6m8s
# monitoring-stack   pod/monitoring-stack-prometheus-node-exporter-57x6l          0/1     Pending   0               6m8s
# monitoring-stack   pod/monitoring-stack-prometheus-node-exporter-dl2st          0/1     Pending   0               6m8s
# monitoring-stack   pod/monitoring-stack-prometheus-node-exporter-msfv8          0/1     Pending   0               6m8s
# monitoring-stack   pod/monitoring-stack-prometheus-node-exporter-nbs2j          0/1     Pending   0               6m8s
# monitoring-stack   pod/prometheus-monitoring-stack-kube-prom-prometheus-0       2/2     Running   0               5m59s
# kube-system        service/monitoring-stack-kube-prom-coredns                   ClusterIP   None             <none>        9153/TCP                       6m8s
# kube-system        service/monitoring-stack-kube-prom-kube-controller-manager   ClusterIP   None             <none>        10257/TCP                      6m8s
# kube-system        service/monitoring-stack-kube-prom-kube-etcd                 ClusterIP   None             <none>        2381/TCP                       6m8s
# kube-system        service/monitoring-stack-kube-prom-kube-proxy                ClusterIP   None             <none>        10249/TCP                      6m8s
# kube-system        service/monitoring-stack-kube-prom-kube-scheduler            ClusterIP   None             <none>        10259/TCP                      6m8s
# kube-system        service/monitoring-stack-kube-prom-kubelet                   ClusterIP   None             <none>        10250/TCP,4194/TCP,10255/TCP   6m
# monitoring-stack   service/alertmanager-operated                                ClusterIP   None             <none>        9093/TCP,9094/TCP,9094/UDP     5m59s
# monitoring-stack   service/monitoring-stack-grafana                             ClusterIP   10.107.245.223   <none>        80/TCP                         6m8s
# monitoring-stack   service/monitoring-stack-kube-prom-alertmanager              ClusterIP   10.99.107.10     <none>        9093/TCP,8080/TCP              6m8s
# monitoring-stack   service/monitoring-stack-kube-prom-operator                  ClusterIP   10.107.131.193   <none>        443/TCP                        6m8s
# monitoring-stack   service/monitoring-stack-kube-prom-prometheus                ClusterIP   10.110.221.42    <none>        9090/TCP,8080/TCP              6m8s
# monitoring-stack   service/monitoring-stack-kube-state-metrics                  ClusterIP   10.109.213.98    <none>        8080/TCP                       6m8s
# monitoring-stack   service/monitoring-stack-prometheus-node-exporter            ClusterIP   10.97.73.142     <none>        9100/TCP                       6m8s
# monitoring-stack   service/prometheus-operated                                  ClusterIP   None             <none>        9090/TCP                       5m59s
# monitoring-stack   daemonset.apps/monitoring-stack-prometheus-node-exporter   4         4         0       4            0           kubernetes.io/os=linux   6m8s
# monitoring-stack   deployment.apps/monitoring-stack-grafana              1/1     1            1           6m8s
# monitoring-stack   deployment.apps/monitoring-stack-kube-prom-operator   1/1     1            1           6m8s
# monitoring-stack   deployment.apps/monitoring-stack-kube-state-metrics   1/1     1            1           6m8s
# monitoring-stack   replicaset.apps/monitoring-stack-grafana-745c96f497              1         1         1       6m8s
# monitoring-stack   replicaset.apps/monitoring-stack-kube-prom-operator-766c56bbb4   1         1         1       6m8s
# monitoring-stack   replicaset.apps/monitoring-stack-kube-state-metrics-5657b445cb   1         1         1       6m8s
# monitoring-stack   statefulset.apps/alertmanager-monitoring-stack-kube-prom-alertmanager   1/1     5m59s
# monitoring-stack   statefulset.apps/prometheus-monitoring-stack-kube-prom-prometheus       1/1     5m59s
```

## DESINSTALLATION STACK MANUELLE

Pour éviter les conflits de déploiement on devra supprimer l'ensemble des éléments installés en étape #02

```bash
# GRAFANA
helm uninstall grafana-dashboard
# NODE-EXPORTER
# helm uninstall node-exporter
# KUBE-STATE-METRICS
helm uninstall kube-state-metrics
# PROMETHEUS MANUEL
kubectl delete -f prometheus/
# NAMESPACE DEDIE
kubectl  namespace monitoring
kubectl delete -f prometheus/
# clusterrole.rbac.authorization.k8s.io "prometheus" deleted
# clusterrolebinding.rbac.authorization.k8s.io "prometheus" deleted
# Error from server (NotFound): error when deleting "prometheus/config-map.yaml": configmaps "prometheus-server-conf" not found
# Error from server (NotFound): error when deleting "prometheus/prometheus-deployment.yaml": deployments.apps "prometheus-deployment" not found
# Error from server (NotFound): error when deleting "prometheus/prometheus-service.yaml": services "prometheus-service" not found

# Suppression de la stack si déjà installée comme je l'avais fait
helm uninstall monitoring-stack prometheus-community/kube-prometheus-stack -n monitoring-stack

# Service résiduel
kubectl delete svc monitoring-stack-kube-prom-kubelet -n kube-system

# Suppression des CRDs résiduelles
kubectl delete crd \
alertmanagerconfigs.monitoring.coreos.com \
probes.monitoring.coreos.com \
prometheusagents.monitoring.coreos.com \
scrapeconfigs.monitoring.coreos.com \
thanosrulers.monitoring.coreos.com \
alertmanagers.monitoring.coreos.com \
podmonitors.monitoring.coreos.com \
prometheuses.monitoring.coreos.com \
prometheusrules.monitoring.coreos.com \
servicemonitors.monitoring.coreos.com

# Réinstallation de la stack complète après vérification de la suppression de l'ensemble des éléments monitoring-stack
kubectl get crds | grep monitoring.coreos.com

helm install monitoring-stack prometheus-community/kube-prometheus-stack -n monitoring-stack
```

Etat entre Désinstallation & Réinstallation : seuls les composants système de base restent

```bash
kc get all -A
# NAMESPACE     NAME                                            READY   STATUS    RESTARTS      AGE
# kube-system   pod/coredns-7d764666f9-f7kgh                    1/1     Running   1 (17h ago)   29h
# kube-system   pod/etcd-control-plane-lab                      1/1     Running   1 (17h ago)   29h
# kube-system   pod/kindnet-kkvdp                               1/1     Running   1 (17h ago)   29h
# kube-system   pod/kindnet-lz7s5                               1/1     Running   1 (17h ago)   29h
# kube-system   pod/kindnet-mj6mj                               1/1     Running   1 (17h ago)   29h
# kube-system   pod/kindnet-s578z                               1/1     Running   1 (17h ago)   29h
# kube-system   pod/kube-apiserver-control-plane-lab            1/1     Running   1 (17h ago)   29h
# kube-system   pod/kube-controller-manager-control-plane-lab   1/1     Running   1 (17h ago)   29h
# kube-system   pod/kube-proxy-9k9t7                            1/1     Running   1 (17h ago)   29h
# kube-system   pod/kube-proxy-c4l64                            1/1     Running   1 (17h ago)   29h
# kube-system   pod/kube-proxy-jdr46                            1/1     Running   1 (17h ago)   29h
# kube-system   pod/kube-proxy-wfdl2                            1/1     Running   1 (17h ago)   29h
# kube-system   pod/kube-scheduler-control-plane-lab            1/1     Running   1 (17h ago)   29h
# kube-system   pod/metrics-server-9d74bb658-29cj4              1/1     Running   2 (8h ago)    21h
# kube-system   pod/storage-provisioner                         1/1     Running   2 (8h ago)    29h

# NAMESPACE     NAME                     TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                  AGE
# default       service/kubernetes       ClusterIP   10.96.0.1      <none>        443/TCP                  29h
# kube-system   service/kube-dns         ClusterIP   10.96.0.10     <none>        53/UDP,53/TCP,9153/TCP   29h
# kube-system   service/metrics-server   ClusterIP   10.98.169.79   <none>        443/TCP                  21h

# NAMESPACE     NAME                        DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR            AGE
# kube-system   daemonset.apps/kindnet      4         4         4       4            4           <none>                   29h
# kube-system   daemonset.apps/kube-proxy   4         4         4       4            4           kubernetes.io/os=linux   29h

# NAMESPACE     NAME                             READY   UP-TO-DATE   AVAILABLE   AGE
# kube-system   deployment.apps/coredns          1/1     1            1           29h
# kube-system   deployment.apps/metrics-server   1/1     1            1           21h

# NAMESPACE     NAME                                       DESIRED   CURRENT   READY   AGE
# kube-system   replicaset.apps/coredns-7d764666f9         1         1         1       29h
# kube-system   replicaset.apps/metrics-server-9d74bb658   1         1         1       21h
```

# ERATUM : Retour à la case installation cluster

Après avoir observé des comportements étranges et nous être penchés sur les README du chart kube-prometheus-stack, nous nous sommes rendu compte qu'il serait problématique de continuer avec le cluster tel qu'il était en **étape #01**.

Nous avons donc collecté les paramètres à utiliser, supprimé le cluster initial tout entier et nous l'avons recréer avec les configurations nécessaires pour le fonctionnement correct et complet du chart que nous voulons installer : `kube-prometheus-stack`.

```bash
# Arrêt & suppression
minikube stop
minikube deletecomplète -p control-plane-lab

# Recréation avec configuration adaptée à kube-prometheus-stack
minikube start \
--profile=control-plane-lab \
--driver=docker \
--container-runtime=containerd \
--kubernetes-version=v1.35.1 \
--nodes=4 \
--memory=8192 \
--cpus=6 \
--bootstrapper=kubeadm \
--extra-config=kubelet.authentication-token-webhook=true \
--extra-config=kubelet.authorization-mode=Webhook \
--extra-config=scheduler.bind-address=0.0.0.0 \
--extra-config=controller-manager.bind-address=0.0.0.0

# Désactivation facultative de metrics-server
# (inutile si conservation volontaire de kubectl top)
# minikube addons disable metrics-server
minikube addons enable metrics-server

# Modification de la configuration kube-proxy
kubectl -n kube-system edit cm kube-proxy
# Modifier :
# metricsBindAddress: 127.0.0.1:10249
# par :
# metricsBindAddress: 0.0.0.0:10249

# Redémarrage kube-proxy
kubectl rollout restart daemonset kube-proxy -n kube-system


# Vérification du profile actif
minikube profile control-plane-lab
# Vérification du contexte kubectl
kubectl config current-context
```

Après installation voici ce que nous obtenons :

```bash
# Ensemble des éléments installés apr le chart dans le namespace dédié
kubectl get all -n monitoring-stack
# NAME                                                         READY   STATUS    RESTARTS   AGE
# pod/alertmanager-monitoring-stack-kube-prom-alertmanager-0   2/2     Running   0          2m37s
# pod/monitoring-stack-grafana-59c7794d5f-qmxdp                3/3     Running   0          2m45s
# pod/monitoring-stack-kube-prom-operator-5f6d448754-hs9t4     1/1     Running   0          2m45s
# pod/monitoring-stack-kube-state-metrics-5657b445cb-mlp9s     1/1     Running   0          2m45s
# pod/monitoring-stack-prometheus-node-exporter-k2kq6          1/1     Running   0          2m45s
# pod/monitoring-stack-prometheus-node-exporter-ld54x          1/1     Running   0          2m45s
# pod/monitoring-stack-prometheus-node-exporter-nd5gr          1/1     Running   0          2m45s
# pod/monitoring-stack-prometheus-node-exporter-tt9qz          1/1     Running   0          2m45s
# pod/prometheus-monitoring-stack-kube-prom-prometheus-0       2/2     Running   0          2m37s

# NAME                                                TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
# service/alertmanager-operated                       ClusterIP   None             <none>        9093/TCP,9094/TCP,9094/UDP   2m37s
# service/monitoring-stack-grafana                    ClusterIP   10.105.152.233   <none>        80/TCP                       2m46s
# service/monitoring-stack-kube-prom-alertmanager     ClusterIP   10.109.95.239    <none>        9093/TCP,8080/TCP            2m46s
# service/monitoring-stack-kube-prom-operator         ClusterIP   10.101.200.143   <none>        443/TCP                      2m46s
# service/monitoring-stack-kube-prom-prometheus       ClusterIP   10.111.225.243   <none>        9090/TCP,8080/TCP            2m46s
# service/monitoring-stack-kube-state-metrics         ClusterIP   10.105.137.100   <none>        8080/TCP                     2m46s
# service/monitoring-stack-prometheus-node-exporter   ClusterIP   10.97.242.165    <none>        9100/TCP                     2m46s
# service/prometheus-operated                         ClusterIP   None             <none>        9090/TCP                     2m37s

# NAME                                                       DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR            AGE
# daemonset.apps/monitoring-stack-prometheus-node-exporter   4         4         4       4            4           kubernetes.io/os=linux   2m45s

# NAME                                                  READY   UP-TO-DATE   AVAILABLE   AGE
# deployment.apps/monitoring-stack-grafana              1/1     1            1           2m45s
# deployment.apps/monitoring-stack-kube-prom-operator   1/1     1            1           2m45s
# deployment.apps/monitoring-stack-kube-state-metrics   1/1     1            1           2m45s

# NAME                                                             DESIRED   CURRENT   READY   AGE
# replicaset.apps/monitoring-stack-grafana-59c7794d5f              1         1         1       2m45s
# replicaset.apps/monitoring-stack-kube-prom-operator-5f6d448754   1         1         1       2m45s
# replicaset.apps/monitoring-stack-kube-state-metrics-5657b445cb   1         1         1       2m45s

# NAME                                                                    READY   AGE
# statefulset.apps/alertmanager-monitoring-stack-kube-prom-alertmanager   1/1     2m37s
# statefulset.apps/prometheus-monitoring-stack-kube-prom-prometheus       1/1     2m37s
# 

# CRDS installées par Prometheus Operator
kc get crds
# NAME                                        CREATED AT
# alertmanagerconfigs.monitoring.coreos.com   2026-06-03T21:57:54Z
# alertmanagers.monitoring.coreos.com         2026-06-03T21:57:54Z
# podmonitors.monitoring.coreos.com           2026-06-03T21:57:54Z
# probes.monitoring.coreos.com                2026-06-03T21:57:55Z
# prometheusagents.monitoring.coreos.com      2026-06-03T21:57:55Z
# prometheuses.monitoring.coreos.com          2026-06-03T21:57:55Z
# prometheusrules.monitoring.coreos.com       2026-06-03T21:57:55Z
# scrapeconfigs.monitoring.coreos.com         2026-06-03T21:57:55Z
# servicemonitors.monitoring.coreos.com       2026-06-03T21:57:55Z
# thanosrulers.monitoring.coreos.com          2026-06-03T21:57:55Z
# 

# Les CRs en cours d'exécution
kc get servicemonitors -A
# NAMESPACE          NAME                                                 AGE
# monitoring-stack   monitoring-stack-grafana                             3m22s
# monitoring-stack   monitoring-stack-kube-prom-alertmanager              3m22s
# monitoring-stack   monitoring-stack-kube-prom-apiserver                 3m22s
# monitoring-stack   monitoring-stack-kube-prom-coredns                   3m22s
# monitoring-stack   monitoring-stack-kube-prom-kube-controller-manager   3m22s
# monitoring-stack   monitoring-stack-kube-prom-kube-etcd                 3m22s
# monitoring-stack   monitoring-stack-kube-prom-kube-proxy                3m22s
# monitoring-stack   monitoring-stack-kube-prom-kube-scheduler            3m22s
# monitoring-stack   monitoring-stack-kube-prom-kubelet                   3m22s
# monitoring-stack   monitoring-stack-kube-prom-operator                  3m22s
# monitoring-stack   monitoring-stack-kube-prom-prometheus                3m22s
# monitoring-stack   monitoring-stack-kube-state-metrics                  3m22s
# monitoring-stack   monitoring-stack-prometheus-node-exporter            3m22s
```

On peut désormais comprendre le fonctionnement de l'opérateur en s'intéressant au CR servicemonitors :

```bash
kubectl get servicemonitors/monitoring-stack-kube-prom-apiserver -n monitoring-stack -o yaml
# apiVersion: monitoring.coreos.com/v1
# kind: ServiceMonitor
# metadata:
#   annotations:
#     meta.helm.sh/release-name: monitoring-stack
#     meta.helm.sh/release-namespace: monitoring-stack
#   creationTimestamp: "2026-06-03T21:58:25Z"
#   generation: 1
#   labels:
#     app: kube-prometheus-stack-apiserver
#     app.kubernetes.io/instance: monitoring-stack
#     app.kubernetes.io/managed-by: Helm
#     app.kubernetes.io/part-of: kube-prometheus-stack
#     app.kubernetes.io/version: 86.1.1
#     chart: kube-prometheus-stack-86.1.1
#     heritage: Helm
#     release: monitoring-stack
#   name: monitoring-stack-kube-prom-apiserver
#   namespace: monitoring-stack
#   resourceVersion: "1481"
#   uid: 6039d567-7388-4115-a699-1a0087298aec
# spec:
#   endpoints:
#   - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
#     metricRelabelings:
#     - action: drop
#       regex: (etcd_request|apiserver_request_slo|apiserver_request_sli|apiserver_request)_duration_seconds_bucket;(0\.15|0\.2|0\.3|0\.35|0\.4|0\.45|0\.6|0\.7|0\.8|0\.9|1\.25|1\.5|1\.75|2|3|3\.5|4|4\.5|6|7|8|9|15|20|40|45|50)(\.0)?
#       sourceLabels:
#       - __name__
#       - le
#     port: https
#     scheme: https
#     tlsConfig:
#       caFile: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
#       insecureSkipVerify: false
#       serverName: kubernetes
#   jobLabel: component
#   namespaceSelector:
#     matchNames:
#     - default
#   selector:
#     matchLabels:
#       component: apiserver
#       provider: kubernetes

kubectl describe servicemonitors/monitoring-stack-kube-prom-apiserver -n monitoring-stack
# Name:         monitoring-stack-kube-prom-apiserver
# Namespace:    monitoring-stack
# Labels:       app=kube-prometheus-stack-apiserver
#               app.kubernetes.io/instance=monitoring-stack
#               app.kubernetes.io/managed-by=Helm
#               app.kubernetes.io/part-of=kube-prometheus-stack
#               app.kubernetes.io/version=86.1.1
#               chart=kube-prometheus-stack-86.1.1
#               heritage=Helm
#               release=monitoring-stack
# Annotations:  meta.helm.sh/release-name: monitoring-stack
#               meta.helm.sh/release-namespace: monitoring-stack
# API Version:  monitoring.coreos.com/v1
# Kind:         ServiceMonitor
# Metadata:
#   Creation Timestamp:  2026-06-03T21:58:25Z
#   Generation:          1
#   Resource Version:    1481
#   UID:                 6039d567-7388-4115-a699-1a0087298aec
# Spec:
#   Endpoints:
#     Bearer Token File:  /var/run/secrets/kubernetes.io/serviceaccount/token
#     Metric Relabelings:
#       Action:  drop
#       Regex:   (etcd_request|apiserver_request_slo|apiserver_request_sli|apiserver_request)_duration_seconds_bucket;(0\.15|0\.2|0\.3|0\.35|0\.4|0\.45|0\.6|0\.7|0\.8|0\.9|1\.25|1\.5|1\.75|2|3|3\.5|4|4\.5|6|7|8|9|15|20|40|45|50)(\.0)?
#       Source Labels:
#         __name__
#         le
#     Port:    https
#     Scheme:  https
#     Tls Config:
#       Ca File:               /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
#       Insecure Skip Verify:  false
#       Server Name:           kubernetes
#   Job Label:                 component
#   Namespace Selector:
#     Match Names:
#       default
#   Selector:
#     Match Labels:
#       Component:  apiserver
#       Provider:   kubernetes
# Events:           <none>

```
## Fonctionnement ServiceMonitors

```text
ServiceMonitor
        ↓
sélectionne un Service
        ↓
Service → Endpoints → Pods
        ↓
Prometheus scrape
        ↓
HTTPS + certificat TLS
        ↓
Bearer Token
        ↓
RBAC Kubernetes
```
1️⃣ Quel Service surveiller
```
selector:
  matchLabels:
    Component: apiserver
    Provider: kubernetes
```
2️⃣ Dans quel namespace chercher
```
namespaceSelector:
  matchNames:
    - default
```
3️⃣ Quel port utiliser
`port: https`
4️⃣ Quel protocole utiliser
`scheme: https`
5️⃣ Comment authentifier le serveur
```
tlsConfig:
  caFile: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
```
6️⃣ Comment identifier le client
```
bearerTokenFile:
  /var/run/secrets/kubernetes.io/serviceaccount/token
```

En outre, le ServiceMonitor kubelet montre quelque chose d'extrêmement important :

`path: /metrics`
`path: /metrics/cadvisor`
`path: /metrics/probes`

- Donc : un même endpoint Kubernetes peut exposer plusieurs familles de métriques
- Les `metricRelabelings` montrent que la stack filtre certaines métriques avant ingestion Prometheus
- Petite nuance à saisir au niveau des namespace observés : les `servicemonitors` n'observent pas directement les pods des composants du control-plane. Ils observent le service `kubernetes` qui se trouve dans le namespace `default` : `kubernetes.default.svc`

---

# 📦 Fonctionnement de Prometheus Operator

| Élément | Rôle |
|---|---|
| `Prometheus` (CRD) | Déclare l'état désiré d'une instance Prometheus |
| `Prometheus Operator` | Observe les CRDs puis génère les ressources Kubernetes réelles |
| `ServiceMonitor` / `PodMonitor` | Déclarent automatiquement les targets à scraper |
| `Prometheus StatefulSet` | Instance Prometheus réellement exécutée |
| `config-reloader` | Recharge Prometheus dynamiquement lors d'un changement |
| `ConfigMaps` | Dashboards Grafana + règles Prometheus |
| `Secrets` | Configuration Prometheus finale + TLS + tokens |

## 🔄 Cycle réel de fonctionnement

```text
ServiceMonitor
        ↓
Prometheus Operator
        ↓
Configuration Prometheus générée
        ↓
Secret / ConfigMaps
        ↓
StatefulSet Prometheus
        ↓
Pod Prometheus
        ↓
Scraping automatique
```

## 🔍 Découverte automatique des targets

```yaml
serviceMonitorSelector:
  matchLabels:
    Release: monitoring-stack
```

```text
Prometheus
↓
sélectionne les ServiceMonitor ayant ce label
↓
Service -> Endpoints -> Pods
↓
scraping
```

## 🔄 Rechargement dynamique

```text
ServiceMonitor modifié
↓
Operator régénère la configuration
↓
Secret mis à jour
↓
config-reloader détecte le changement
↓
POST /-/reload
↓
Prometheus recharge sa configuration
```

## 📊 Dashboards Grafana

```text
ConfigMap contenant dashboard JSON
↓
Grafana détecte automatiquement le label grafana_dashboard=1
↓
Import automatique du dashboard
```

## 🧠 Idée fondamentale

```text
CRD
↓
Operator
↓
Reconciliation
↓
Ressources Kubernetes générées automatiquement
```

C'est exactement le pattern utilisé avec Kubebuilder
