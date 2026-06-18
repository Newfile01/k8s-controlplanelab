#!/usr/bin/env bash

set -e

PROFILE="control-plane-lab"

echo "============================================================"
echo "VERIFICATION PROFIL MINIKUBE"
echo "============================================================"

# Vérifie si le profil existe réellement
if ! minikube profile list | grep -q "${PROFILE}"; then
    echo "[ERREUR] Profil Minikube '${PROFILE}' introuvable"
    echo "Créer le cluster manuellement avant d'utiliser ce script"
    exit 1
fi

echo "[OK] Profil ${PROFILE} trouvé"

# Vérifie si le contexte kubectl est correct
CURRENT_CONTEXT=$(kubectl config current-context 2>/dev/null || true)

if [[ "${CURRENT_CONTEXT}" != *"${PROFILE}"* ]]; then
    echo "[INFO] Activation du profil ${PROFILE}"
    minikube profile ${PROFILE}
fi

echo ""
echo "============================================================"
echo "DEMARRAGE MINIKUBE"
echo "============================================================"

minikube start \
--profile=control-plane-lab \
--driver=docker \
--container-runtime=containerd \
--kubernetes-version=v1.35.1 \
--addons=[metrics-server] \
--nodes=4 \
--memory=4096 \
--cpus=2 \
--bootstrapper=kubeadm \
--extra-config=kubelet.authentication-token-webhook=true \
--extra-config=kubelet.authorization-mode=Webhook \
--extra-config=scheduler.bind-address=0.0.0.0 \
--extra-config=controller-manager.bind-address=0.0.0.0


echo ""
echo "============================================================"
echo "ATTENTE CLUSTER"
echo "============================================================"

kubectl wait --for=condition=Ready nodes --all --timeout=300s
minikube addons enable metrics-server
minikube addons enable yakd

echo "[OK] Cluster prêt"

echo ""
echo "============================================================"
echo "LANCEMENT PORT-FORWARD GRAFANA"
echo "============================================================"

# Nettoyage ancien process éventuel
pkill -f "port-forward.*monitoring-stack-grafana" || true

nohup kubectl port-forward \
svc/monitoring-stack-grafana \
3000:80 \
-n monitoring-stack \
--address 0.0.0.0 \
> grafana.log 2>&1 &

sleep 5

echo ""
echo "============================================================"
echo "LANCEMENT PORT-FORWARD PROMETHEUS"
echo "============================================================"

# Nettoyage ancien process éventuel
pkill -f "port-forward.*monitoring-stack-kube-prom-prometheus" || true

nohup kubectl port-forward \
svc/monitoring-stack-kube-prom-prometheus \
9090:9090 \
-n monitoring-stack \
--address 0.0.0.0 \
> prometheus.log 2>&1 &

sleep 5

echo ""
echo "============================================================"
echo "VERIFICATION GRAFANA"
echo "============================================================"

GRAFANA_HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
http://localhost:3000)

if [ "${GRAFANA_HTTP_CODE}" = "302" ] || \
   [ "${GRAFANA_HTTP_CODE}" = "200" ]; then
    echo "[OK] Grafana accessible : http://localhost:3000"
else
    echo "[NOK] Grafana inaccessible"
fi

echo ""
echo "============================================================"
echo "VERIFICATION PROMETHEUS"
echo "============================================================"

PROM_HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
http://localhost:9090)

if [ "${PROM_HTTP_CODE}" = "302" ] || \
   [ "${PROM_HTTP_CODE}" = "200" ]; then
    echo "[OK] Prometheus accessible : http://localhost:9090"
else
    echo "[NOK] Prometheus inaccessible"
fi

echo ""
echo "============================================================"
echo "CLUSTER OPERATIONNEL"
echo "============================================================"

kubectl get nodes