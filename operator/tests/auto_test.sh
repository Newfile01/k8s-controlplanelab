#!/usr/bin/env bash

# ==========================================================
# DEFINITION DES PARAMETRES DU TEST
# ==========================================================
# Nomenclature label (pour les retrouver dans Grafana)
# D   : DEPLOYMENTS
# R   : REPLICAS_PER_DEP
# CM  : CONFIGMAPS
# SEC : SECRETS
#
# Ne pas mettre de '_' mais plutôt des '-'

TEST_LABEL="D10R1-R1-60s-30min"
CR_NAME="cr$(date '+%Y-%m-%d-%H%M')"

# Horodatage réel du lancement
START_TIME="$(date '+%Y-%m-%d %H:%M:%S')"
START_EPOCH="$(date +%s)"

# CHARGE INITIALE
DEPLOYMENTS=10
REPLICAS_PER_DEP=1
CPU_PER_POD="25m"
MEM_PER_POD="32Mi"
# Limits = Request x2
CPU_LIMIT_PER_POD="50m"
MEM_LIMIT_PER_POD="64Mi"

CONFIGMAPS=0
SIZE_KB_CM=0
SECRETS=0
SIZE_KB_SEC=0

# Limites
MAX_DEPLOYMENTS=10
MAX_REPLICAS_PER_DEP=300
MAX_CONFIGMAPS=0
MAX_SECRETS=0

# Définir la liste des ressources à faire varier
#
# Valeurs possibles :
# DEPLOYMENTS
# REPLICAS_PER_DEP
# CONFIGMAPS
# SECRETS
#
# (retour à la ligne entre chaque et valeur entre guillemets)
VARIABLE_LOADS=(
    "REPLICAS_PER_DEP"
)

INCREMENT=1
# secondes entre deux mises à jour de la CR
UPDATE_INTERVAL=60
# durée totale du test en secondes
TEST_DURATION=1800

# Définition de la fin du test
END_EPOCH=$((START_EPOCH + TEST_DURATION))
END_TIME="$(date \
    -d "@${END_EPOCH}" \
    '+%Y-%m-%d %H:%M:%S')"

# ==========================================================
# AUTRES PARAMETRES FIXES DE LA CR
# ==========================================================
IMAGE="nginx"
CONTAINER_NAME="app"

SERVICE_PORT=80
TARGET_PORT=80

SCHEDULER_ENABLED=false

NODE_SELECTOR_KEY="minikube.k8s.io/primary"
NODE_SELECTOR_VALUE="false"

TOPOLOGY_SPREAD=false

AFFINITY_MODE=""
AFFINITY_KEY=""
AFFINITY_VALUE=""

ANTI_AFFINITY_MODE=""
ANTI_AFFINITY_KEY=""
ANTI_AFFINITY_VALUE=""

APISERVER_ETCD_ENABLED=false
FREQUENT_STATUS_UPDATES=false
AGGRESSIVE_RECONCILE=false
RECONCILE_REQUEUE_DELAY=1000
RECREATE_RESOURCES=false

CONTROLLER_MANAGER_ENABLED=false
RECREATE_REPLICASETS=false
REPLICA_RECREATE_DELAY=1000
AGGRESSIVE_GC=false
GC_DELAY=1000

OPERATOR_ENABLED=true
OPERATOR_PROFILE="standard"

POD_LIFECYCLE_ENABLED=false
DELETE_PODS_RANDOMLY=false
DELETE_PODS_DELAY=1000
CRASH_LOOP_SIMULATION=false

# ==========================================================
# AFFICHAGE RECAP + CREATION LOGFILE
# ==========================================================

{
    echo "=============================================================="
    printf "%-18s : %-30s %-18s : %s\n" \
        "TEST_LABEL" "${TEST_LABEL}" \
        "CR_NAME" "${CR_NAME}"
    echo
    printf "%-18s : %-30s %-18s : %s\n" \
        "Début" "${START_TIME}" \
        "Fin prévue" "${END_TIME}"
    echo
    printf "%-18s : %-30s %-18s : %s\n" \
        "DEPLOYMENTS" "${DEPLOYMENTS}" \
        "CPU" "${CPU_PER_POD}"
    printf "%-18s : %-30s %-18s : %s\n" \
        "REPLICAS_PER_DEP" "${REPLICAS_PER_DEP}" \
        "MEM" "${MEM_PER_POD}"
    echo
    printf "%-18s : %-30s %-18s : %s\n" \
        "CONFIGMAPS" "${CONFIGMAPS}" \
        "SIZE_CM" "${SIZE_KB_CM}kB"
    printf "%-18s : %-30s %-18s : %s\n" \
        "SECRETS" "${SECRETS}" \
        "SIZE_SEC" "${SIZE_KB_SEC}kB"
    echo
    printf "%-18s : %s\n" \
        "VARIABLE_LOADS" "${VARIABLE_LOADS[*]}"
    printf "%-18s : %-30s %-18s : %s\n" \
        "INCREMENT" "${INCREMENT}" \
        "UPDATE_INTERVAL" "${UPDATE_INTERVAL}s"
    printf "%-18s : %s\n" \
        "TEST_DURATION" "${TEST_DURATION}s"
    echo "=============================================================="
} | tee "${CR_NAME}.log"


# ==========================================================
# CREATION CR
# ==========================================================
generate_cr() {
    
cat > "${CR_NAME}.yaml" <<EOF
apiVersion: controlplane.lab.local/v1alpha1
kind: ControlPlaneTest
metadata:
    name: ${CR_NAME}
    namespace: operator-system

spec:
    image: "${IMAGE}"

    containerName: "${CONTAINER_NAME}"

    deploymentCount: ${DEPLOYMENTS}
    replicasPerDeployment: ${REPLICAS_PER_DEP}

    servicePort: ${SERVICE_PORT}
    targetPort: ${TARGET_PORT}

    customLabel:
        key: "stress"
        value: "${TEST_LABEL}"

    resourcesPerPod:
        cpuRequest: "${CPU_PER_POD}"
        cpuLimit: "${CPU_LIMIT_PER_POD}"
        memoryRequest: "${MEM_PER_POD}"
        memoryLimit: "${MEM_LIMIT_PER_POD}"

    configMapCount: ${CONFIGMAPS}
    configMapSizeKB: ${SIZE_KB_CM}

    secretCount: ${SECRETS}
    secretSizeKB: ${SIZE_KB_SEC}

    # ==========================================================
    # SCHEDULER STRESS
    # ==========================================================

    schedulerStress:
        enabled: ${SCHEDULER_ENABLED}

        nodeSelector:
            ${NODE_SELECTOR_KEY}: "${NODE_SELECTOR_VALUE}"

        topologySpread: ${TOPOLOGY_SPREAD}

        affinityMode: "${AFFINITY_MODE}"

        affinitySelector:
            key: "${AFFINITY_KEY}"
            value: "${AFFINITY_VALUE}"

        antiAffinityMode: "${ANTI_AFFINITY_MODE}"

        antiAffinitySelector:
            key: "${ANTI_AFFINITY_KEY}"
            value: "${ANTI_AFFINITY_VALUE}"

    # ==========================================================
    # API SERVER / ETCD STRESS
    # ==========================================================

    apiServerEtcdStress:
        enabled: ${APISERVER_ETCD_ENABLED}

        frequentStatusUpdates: ${FREQUENT_STATUS_UPDATES}

        aggressiveReconcile: ${AGGRESSIVE_RECONCILE}
        reconcileRequeueDelay: ${RECONCILE_REQUEUE_DELAY}

        recreateResources: ${RECREATE_RESOURCES}

    # ==========================================================
    # CONTROLLER MANAGER STRESS
    # ==========================================================

    controllerManagerStress:
        enabled: ${CONTROLLER_MANAGER_ENABLED}

        recreateReplicaSets: ${RECREATE_REPLICASETS}
        replicaRecreateDelay: ${REPLICA_RECREATE_DELAY}

        aggressiveGarbageCollection: ${AGGRESSIVE_GC}
        garbageCollectionDelay: ${GC_DELAY}

    # ==========================================================
    # OPERATOR STRESS
    # ==========================================================

    operatorStress:
        enabled: ${OPERATOR_ENABLED}

        profile: "${OPERATOR_PROFILE}"

    # ==========================================================
    # POD LIFECYCLE STORM
    # ==========================================================

    podLifecycleStorm:
        enabled: ${POD_LIFECYCLE_ENABLED}

        deletePodsRandomly: ${DELETE_PODS_RANDOMLY}
        deletePodsDelay: ${DELETE_PODS_DELAY}

        crashLoopSimulation: ${CRASH_LOOP_SIMULATION}

EOF

}

generate_cr
# kubectl apply -f "${CR_NAME}.yaml" || {
#     echo "❌ Erreur application CR"
#     exit 1
# }

echo "=== YAML généré ==="
cat "${CR_NAME}.yaml"
echo

kubectl apply -f "${CR_NAME}.yaml"

echo "RC=$?"

# Fonction de logging à activer à chaque update
log_update() {
    local timestamp
    timestamp="$(date '+%Y-%m-%d %H:%M:%S')"

    {
        echo
        echo "[$timestamp] UPDATE CR ${CR_NAME}"

        for variable in "${VARIABLE_LOADS[@]}"; do
            echo "  ${variable}=${!variable}"
        done

    } >> "${CR_NAME}.log"
}

all_max_reached=false

# ==========================================================
# BOUCLE DE TEST
# ==========================================================
while [ "$(date +%s)" -lt "${END_EPOCH}" ]; do
    sleep "${UPDATE_INTERVAL}"

    for variable in "${VARIABLE_LOADS[@]}"; do
        case "${variable}" in
            DEPLOYMENTS)
                if [ "${DEPLOYMENTS}" -lt "${MAX_DEPLOYMENTS}" ]; then
                    ((DEPLOYMENTS+=INCREMENT))
                fi
                ;;
            REPLICAS_PER_DEP)
                if [ "${REPLICAS_PER_DEP}" -lt "${MAX_REPLICAS_PER_DEP}" ]; then
                    ((REPLICAS_PER_DEP+=INCREMENT))
                fi
                ;;
            CONFIGMAPS)
                if [ "${CONFIGMAPS}" -lt "${MAX_CONFIGMAPS}" ]; then
                    ((CONFIGMAPS+=INCREMENT))
                fi
                ;;
            SECRETS)

                if [ "${SECRETS}" -lt "${MAX_SECRETS}" ]; then
                    ((SECRETS+=INCREMENT))
                fi
                ;;
        esac
    done

    # Modification CR & Logs
    generate_cr

    kubectl apply -f "${CR_NAME}.yaml" || {
        echo "❌ Erreur application CR"
    exit 1
}
    log_update

    all_max_reached=true

    for variable in "${VARIABLE_LOADS[@]}"; do
        case "${variable}" in
            DEPLOYMENTS)
                [ "${DEPLOYMENTS}" -lt "${MAX_DEPLOYMENTS}" ] && all_max_reached=false
                ;;
            REPLICAS_PER_DEP)
                [ "${REPLICAS_PER_DEP}" -lt "${MAX_REPLICAS_PER_DEP}" ] && all_max_reached=false
                ;;
            CONFIGMAPS)
                [ "${CONFIGMAPS}" -lt "${MAX_CONFIGMAPS}" ] && all_max_reached=false
                ;;
            SECRETS)
                [ "${SECRETS}" -lt "${MAX_SECRETS}" ] && all_max_reached=false
                ;;
        esac
    done

    if ${all_max_reached}; then 
        break
    fi
done

# ==========================================================
# FIN DU TEST
# ==========================================================
echo
echo "Suppression CR ${CR_NAME}"

kubectl delete controlplanetest "${CR_NAME}" \
    -n operator-system
echo "[$(date '+%Y-%m-%d %H:%M:%S')] FIN DU TEST" \
    >> "${CR_NAME}.log"
