# Étape #3 - kube-prometheus-stack complète (opérateur Prometheus + autres composants)

Ce qu'elle contient :

+ [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus/tree/main)
+ [prometheus-operator](https://github.com/prometheus-operator/prometheus-operator)
+ [kube-state-metrics](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-state-metrics)
+ [prometheus-node-exporter](https://github.com/prometheus-community/helm-charts/tree/main/charts/prometheus-node-exporter)
+ [grafana](https://github.com/grafana-community/helm-charts)
- Prometheus Adapter
- Prometheus black-box exporter.

Helm Repository : https://prometheus-community.github.io/helm-charts
OCI Artifact Registry : oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack

Repartir d'un cluster avec la configuration suivante (sur Minikube) :

```bash
# Arrêt & suppression du cluster
minikube stop
minikube delete -p control-plane-lab

# Après suppression, créer à nouveau le cluster avec les bons paramétrages avant l'installation du chart kube-prometheus-stack
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

# Puis
kubectl -n kube-system edit cm kube-proxy
# metricsBindAddress: 127.0.0.1:10249 devient 0.0.0.0:10249

# Et enfin actualiser
kubectl rollout restart daemonset kube-proxy -n kube-system

# Optionnel
minikube addons enable metrics-server
# Sinon désactiver le metrics-server si l'on souhaite utiliser le Prometheus Adapter qui donne : kubectl top, HPA, metrics API
# Ce dernier ayant été enlevé de la stack kube-prometheus-stack, il faudra le réinstaller manuellement
# Donnera : Custom Metrics API, External Metrics API, HPA Prometheus

# Vérif post-install
kubectl get servicemonitors -A
kubectl describe servicemonitor -A
# ServiceMonitor
# ↓
# Prometheus Operator
# ↓
# configuration Prometheus générée
# ↓
# scraping automatique
```

Ces paramètres sont importants pour :
```bash
--extra-config=kubelet.authentication-token-webhook=true \
# Prometheus
# → authentification via ServiceAccount token
# → accès sécurisé au kubelet

--extra-config=kubelet.authorization-mode=Webhook \
# kubelet
# → délègue l'autorisation à l'API Server
# → RBAC Kubernetes appliqué
# => Prometheus 
# → authentifié
# → puis autorisé via RBAC

--extra-config=scheduler.bind-address=0.0.0.0 \
# Par défaut : scheduler metrics → localhost uniquement, Prometheus ne peut pas les scraper
# Avec 0.0.0.0 les métriques deviennent accessibles depuis le cluster.

--extra-config=controller-manager.bind-address=0.0.0.0
# Même logique :
# controller-manager metrics
# → exposées hors localhost
# → scrapables par Prometheus
```

## Installer le chart kube-prometheus-stack


```bash
# Définition d'un namespace dédié
kubectl create namespace monitoring-stack

helm install monitoring-stack \
prometheus-community/kube-prometheus-stack \
--version 86.1.1 \
-n monitoring-stack
# NAME: monitoring-stack
# LAST DEPLOYED: Wed Jun  3 23:58:14 2026
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
