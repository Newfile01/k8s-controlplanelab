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
--wait=apiserver \
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

NAMESPACE="operator-system"
echo "🖼️ Contexte de travail : ${NAMESPACE}"
kubectl config set-context --current --namespace=${NAMESPACE}

VARIABLE_LOADS=(
    "controlplanetests"
    "limitranges"
    "resourcequotas"
    "deployments"
    "replicasets"
    "pods"
    "services"
    "secrets"
    "configmaps"
    "endpoints"
    "endpointslices"
)

mkdir -p launch_logs || {
    echo "❌ Dossier de logs introuvable !"
    exit 1
}
echo "✅ Dossier de logs OK"

START_TIME="$(date '+%Y-%m-%d_%H-%M-%S')"
LOG_FILE="launch_logs/launch_${START_TIME}.log"
touch "$LOG_FILE"
echo "✅ Logfile créé"

sleep 30s
# Usage de substitution de commande BASH (nécessite lancement script avec `bash <script>.sh`)
# `exec` : indique de modifier la sortie standard (stdout) pour tous le reste du script
# `>` (tout seul) : redirection stdout
# `>(...)` : Process Substitution créé un pipe avec la redirection précédente
# `tee -a` : dupplique stdout sur l'écran et dans le fichier défini. Le `-a` permet de ne pas écraser le contenu du fichier ("append")
# `2>&1` : redirige les erreurs ('2') au même endroit que  stdout (donc dans le fichier et à l'écran)
#
# Chaque commande suivant `exec ...` suivra donc le processus établi : 
# redirection vers tee, qui dupplique écran/fichier sans écrasement, avec inclulsion des erreurs
#
#
# Astuce : possible d'inclure ceci au milieu d'un fichier plutôt qu'à la fin en sauvegardant les flux :
# exec 3>&1 4>&2
# 
# Ensuite, après le script à exploitere avec `exec ...` on réhabilite le flux de base :
# exec 1>&3 2>&4
#
# Possible de fermer les flux temporairement créés :
# exec 3>&- 4>&-
exec 3>&1 4>&2
exec > >(tee -a "$LOG_FILE") 2>&1
for variable in "${VARIABLE_LOADS[@]}"; do
    echo
    echo "===== ${variable} ====="
    kubectl get "${variable}" -o wide
    [ "$variable" = "limitranges" ] && echo "SI ABSENT >> PROBABLEMENT PREVU ✌️"
    [ "$variable" = "resourcequotas" ] && echo "SI ABSENT >> VERIFIER LE DIMENSIONNEMENT DES PODS !!! 🫣"
done

echo
echo "===== Nœuds ====="
echo "Cluster infos :"
kubectl get nodes
kubectl top nodes
exec 1>&3 2>&4
exec 3>&- 4>&-