#!/bin/bash
set -e
################################################################################
# ENREGISTREMENT DU TERMINAL
################################################################################

if [ "$1" != "--internal" ]; then

    TERMINAL_LOG="/tmp/campaign-$(date +%Y%m%d-%H%M%S).log"

    CMD="\"$0\" --internal"
    for arg in "$@"; do
        CMD+=" \"${arg}\""
    done

    script -q -c "${CMD}" "${TERMINAL_LOG}"
    RC=$?

    if [ ${RC} -eq 0 ]; then
        rm -f "${TERMINAL_LOG}"
    else
        echo
        echo "Le terminal a été sauvegardé dans :"
        echo "    ${TERMINAL_LOG}"
    fi

    exit ${RC}
fi

shift

################################################################################
# CONFIGURATION
################################################################################
AUTO_TEST="./auto_test.sh"
LOG_DIR="./logs"
SCRAPE_INTERVAL=10
SCENARIO_MARGIN=10      # %
PROTOCOL_MARGIN=5       # %
SCENARIO_PAUSE=120      # secondes entre deux scénarios
HARD_SCENARIO_PAUSE=600 
CR_NAMESPACE="operator-system"
mkdir -p "${LOG_DIR}"

################################################################################
# COULEURS
################################################################################
GREEN="\033[0;32m"; BLUE="\033[1;34m"; RED="\033[0;31m"; NC="\033[0m"

################################################################################
# AFFICHAGE
################################################################################
info()    { echo -e "${BLUE}🔄 $1${NC}"; }
success() { echo -e "${GREEN}✅ $1${NC}"; }
error()   { echo -e "${RED}❌ $1${NC}"; }

################################################################################
# HORODATAGE
################################################################################
timestamp() { date +"%Y%m%d-%H%M%S"; }

################################################################################
# SAUVEGARDE / RESTAURATION DE auto_test.sh
################################################################################
BACKUP_FILE="${AUTO_TEST}.bak"

backup_auto_test() { cp "${AUTO_TEST}" "${BACKUP_FILE}"; }

restore_auto_test() {
    if [ -f "${BACKUP_FILE}" ]; then mv "${BACKUP_FILE}" "${AUTO_TEST}"; fi
}

trap restore_auto_test EXIT INT TERM

################################################################################
# PARAMETRES DU SCENARIO
################################################################################
PARAMETERS=()

add_parameter() { PARAMETERS+=("$1=$2"); }

clear_parameters() { PARAMETERS=(); }


################################################################################
# MODIFICATION VARIABLES auto_test.sh
################################################################################
set_var() {
    sed -i "s|^$1=.*|$1=$2|" "${AUTO_TEST}"
}

set_var_string() {
    sed -i "s|^$1=.*|$1=\"$2\"|" "${AUTO_TEST}"
}

set_parameter() {
    set_var "$1" "$2"
    add_parameter "$1" "$2"
}

set_parameter_string() {
    set_var_string "$1" "$2"
    add_parameter "$1" "$2"
}


################################################################################
# LOGS
################################################################################
write_log_header() {
{
    echo "###############################################################"
    echo "# Kubernetes Control Plane Benchmark"
    echo "###############################################################"
    echo
    echo "Date        : $(date)"
    echo "Protocole   : ${PROTOCOL}"
    echo "Scénario    : ${SCENARIO}"
    echo
    echo "Configuration :"
    echo
    for PARAM in "${PARAMETERS[@]}"; do
        echo "    ${PARAM}"
    done
    echo
    echo "###############################################################"
    echo
} >> "${LOG_FILE}"
}

create_log() {
    PROTOCOL=$1; SCENARIO=$2
    mkdir -p "${LOG_DIR}/${PROTOCOL}"
    LOG_FILE="${LOG_DIR}/${PROTOCOL}/${SCENARIO}-$(timestamp).log"
    write_log_header
}

log() { echo "$1" >> "${LOG_FILE}"; }


################################################################################
# ETAT INITIAL
################################################################################
capture_cluster_state() {
    log ""; log "======================================================="
    log ""
    log "DATE : $(date)"
    log ""
    log "CLUSTER INITIAL"; log "======================================================="; log ""
    kubectl top nodes >> "${LOG_FILE}"

    log ""
    log "======================================================="
    log "PLAN DE CONTROLE INITIAL"
    log "======================================================="
    log ""

    kubectl top pods -A | awk '
    /etcd-control-plane-lab/ ||
    /kube-apiserver-control-plane-lab/ ||
    /kube-controller-manager-control-plane-lab/ ||
    /kube-scheduler-control-plane-lab/ ||
    /operator-controller-manager/
    ' >> "${LOG_FILE}"
}

################################################################################
# ATTENTES
################################################################################
wait_cleanup() {
    while kubectl get controlplanetests -n "${CR_NAMESPACE}" --no-headers 2>/dev/null | grep -q .; do
        sleep 5
    done
}

wait_margin() {
    local EXTRA=$(( $1 * SCENARIO_MARGIN / 100 ))
    sleep "${EXTRA}"
}

wait_cluster_stable() {

    echo "Attente de la stabilisation du cluster..."

    until kubectl get --raw=/readyz >/dev/null 2>&1
    do
        sleep 5
    done

    sleep 60
}

################################################################################
# EXECUTION auto_test.sh
################################################################################
run_auto_test() {
    chmod +x "${AUTO_TEST}"

    ./auto_test.sh || {
        error "Erreur durant auto_test.sh"
        exit 1
    }
}

################################################################################
# EXECUTION D'UN SCENARIO
################################################################################
run_scenario() {
    local PROTOCOL=$1; local SCENARIO=$2
    info "${PROTOCOL}-${SCENARIO} démarré $(timestamp)"
    create_log "${PROTOCOL}" "${SCENARIO}"
    capture_cluster_state
    run_auto_test
    wait_cleanup
    wait_cluster_stable
    wait_margin "${SCENARIO_DURATION}"
    info "Pause ${SCENARIO_PAUSE}s avant le scénario suivant..."
    sleep "${SCENARIO_PAUSE}"
    success "${PROTOCOL}-${SCENARIO} terminé $(timestamp)"
    clear_parameters
}

set_default_configuration() {

    set_parameter DEPLOYMENTS 0
    set_parameter REPLICAS_PER_DEP 0
    set_parameter CONFIGMAPS 0
    set_parameter SECRETS 0

    # Placement des Pods sur les Workers
    set_parameter SCHEDULER_ENABLED true
    set_parameter_string NODE_SELECTOR_KEY "minikube.k8s.io/primary"
    set_parameter_string NODE_SELECTOR_VALUE "false"

    # Désactivation de tous les stress spécifiques
    set_parameter APISERVER_ETCD_ENABLED false
    set_parameter CONTROLLER_MANAGER_ENABLED false
    set_parameter POD_LIFECYCLE_ENABLED false
    set_parameter OPERATOR_ENABLED false

    # Remise à zéro des options
    set_parameter FREQUENT_STATUS_UPDATES false
    set_parameter AGGRESSIVE_RECONCILE false
    set_parameter RECREATE_RESOURCES false
    set_parameter RECREATE_REPLICASETS false
    set_parameter AGGRESSIVE_GC false
    set_parameter DELETE_PODS_RANDOMLY false
    set_parameter CRASH_LOOP_SIMULATION false
    set_parameter RECONCILE_REQUEUE_DELAY 30
    set_parameter DELETE_PODS_DELAY 30
    set_parameter GC_DELAY 30
    set_parameter REPLICA_RECREATE_DELAY 30
}

################################################################################
# PROTOCOLE P1 — Variation du nombre de Deployments
################################################################################
run_protocol_P1() {
    echo; echo "==============================================================="
    echo "PROTOCOLE P1 - Variation du nombre de Deployments"
    echo "==============================================================="; echo

    set_default_configuration

    # S1
    SCENARIO_DURATION=90
    set_parameter DEPLOYMENTS 8;  set_parameter REPLICAS_PER_DEP 5
    run_scenario P1 S1

    # S2
    SCENARIO_DURATION=400
    set_parameter DEPLOYMENTS 36; set_parameter REPLICAS_PER_DEP 5
    run_scenario P1 S2

    # S3
    SCENARIO_DURATION=800
    set_parameter DEPLOYMENTS 72; set_parameter REPLICAS_PER_DEP 5
    run_scenario P1 S3
}

################################################################################
# PROTOCOLE P2 — Variation du nombre de Pods
################################################################################
run_protocol_P2() {
    echo; echo "==============================================================="
    echo "PROTOCOLE P2 - Variation du nombre de Pods"
    echo "==============================================================="; echo

    set_default_configuration

    # S1
    SCENARIO_DURATION=50
    set_parameter DEPLOYMENTS 10; set_parameter REPLICAS_PER_DEP 4
    run_scenario P2 S1

    # S2
    SCENARIO_DURATION=300
    set_parameter DEPLOYMENTS 10; set_parameter REPLICAS_PER_DEP 18
    run_scenario P2 S2

    # S3
    SCENARIO_DURATION=600
    set_parameter DEPLOYMENTS 10; set_parameter REPLICAS_PER_DEP 36
    run_scenario P2 S3
}

################################################################################
# PROTOCOLE P3 — Variation ConfigMaps + Secrets
################################################################################
run_protocol_P3() {
    echo; echo "==============================================================="
    echo "PROTOCOLE P3 - Charge statique"
    echo "==============================================================="; echo

    set_default_configuration

    # S1
    SCENARIO_DURATION=600
    set_parameter CONFIGMAPS 50;   set_parameter SECRETS 50
    run_scenario P3 S1
    sleep "${HARD_SCENARIO_PAUSE}"

    # S2
    SCENARIO_DURATION=600
    set_parameter CONFIGMAPS 250;  set_parameter SECRETS 250
    run_scenario P3 S2
    sleep "${HARD_SCENARIO_PAUSE}"

    # S3
    SCENARIO_DURATION=1000
    set_parameter CONFIGMAPS 1000; set_parameter SECRETS 1000
    run_scenario P3 S3
    sleep "${HARD_SCENARIO_PAUSE}"
}

################################################################################
# PROTOCOLE P4 — Volume ConfigMaps + Secrets
################################################################################
run_protocol_P4() {
    echo; echo "==============================================================="
    echo "PROTOCOLE P4 - Volume des objets"
    echo "==============================================================="; echo
    
    set_default_configuration
    
    # S1
    SCENARIO_DURATION=600
    set_parameter CONFIGMAPS 10; set_parameter SECRETS 10
    set_parameter SIZE_KB_CM 100; set_parameter SIZE_KB_SEC 100
    run_scenario P4 S1
    sleep "${HARD_SCENARIO_PAUSE}"

    # S2    
    SCENARIO_DURATION=600
    set_parameter CONFIGMAPS 10; set_parameter SECRETS 10
    set_parameter SIZE_KB_CM 500; set_parameter SIZE_KB_SEC 500
    run_scenario P4 S2
    sleep "${HARD_SCENARIO_PAUSE}"

    # S3
    SCENARIO_DURATION=600
    set_parameter CONFIGMAPS 10; set_parameter SECRETS 10
    set_parameter SIZE_KB_CM 900; set_parameter SIZE_KB_SEC 900
    run_scenario P4 S3
    sleep "${HARD_SCENARIO_PAUSE}"
}

################################################################################
# PROTOCOLE P5 — Type d'opérations
################################################################################
run_protocol_P5() {
    echo; echo "==============================================================="
    echo "PROTOCOLE P5 - Type d'opérations"
    echo "==============================================================="; echo

    set_default_configuration

    # Activation du stress API Server / ETCD
    set_parameter APISERVER_ETCD_ENABLED true

    ###########################################################################
    # S1 - CREATE
    # Création initiale uniquement
    ###########################################################################
    SCENARIO_DURATION=600

    set_parameter RECREATE_RESOURCES false
    set_parameter FREQUENT_STATUS_UPDATES false
    set_parameter AGGRESSIVE_RECONCILE false

    run_scenario P5 S1

    ###########################################################################
    # S2 - UPDATE
    # Génération d'écritures répétées dans le Status
    ###########################################################################
    SCENARIO_DURATION=600

    set_parameter RECREATE_RESOURCES false
    set_parameter FREQUENT_STATUS_UPDATES true
    set_parameter AGGRESSIVE_RECONCILE false

    run_scenario P5 S2

    ###########################################################################
    # S3 - DELETE / RECREATE
    # Suppression puis recréation continue des Deployments
    ###########################################################################
    SCENARIO_DURATION=600

    set_parameter RECREATE_RESOURCES true
    set_parameter FREQUENT_STATUS_UPDATES false
    set_parameter AGGRESSIVE_RECONCILE false

    run_scenario P5 S3
}

################################################################################
# PROTOCOLE P6 — Fréquence de réconciliation
################################################################################
run_protocol_P6() {
    echo; echo "==============================================================="
    echo "PROTOCOLE P6 - Fréquence de réconciliation"
    echo "==============================================================="; echo

    set_default_configuration
    set_parameter APISERVER_ETCD_ENABLED true

    # S1
    SCENARIO_DURATION=600
    set_parameter AGGRESSIVE_RECONCILE true
    set_parameter RECONCILE_REQUEUE_DELAY 30
    run_scenario P6 S1

    # S2
    SCENARIO_DURATION=600
    set_parameter AGGRESSIVE_RECONCILE true
    set_parameter RECONCILE_REQUEUE_DELAY 10
    run_scenario P6 S2

    # S3
    SCENARIO_DURATION=600
    set_parameter AGGRESSIVE_RECONCILE true
    set_parameter RECONCILE_REQUEUE_DELAY 1
    run_scenario P6 S3
}

################################################################################
# PROTOCOLE P7 — Complexité du Scheduler
################################################################################
run_protocol_P7() {
    echo; echo "==============================================================="
    echo "PROTOCOLE P7 - Complexité du Scheduler"
    echo "==============================================================="; echo

    set_default_configuration

    # Charge fixe ≈80 %
    set_parameter DEPLOYMENTS 72
    set_parameter REPLICAS_PER_DEP 4
    set_parameter SCHEDULER_ENABLED true

    # S1 NodeSelector
    SCENARIO_DURATION=600
    set_parameter_string NODE_SELECTOR_KEY "minikube.k8s.io/primary"
    set_parameter_string NODE_SELECTOR_VALUE "true"
    set_parameter TOPOLOGY_SPREAD false
    set_parameter_string AFFINITY_MODE "None"
    set_parameter_string ANTI_AFFINITY_MODE "None"
    run_scenario P7 S1

    # S2 Affinity
    SCENARIO_DURATION=600
    set_parameter_string AFFINITY_MODE "Required"
    run_scenario P7 S2

    # S3 TopologySpread
    SCENARIO_DURATION=600
    set_parameter TOPOLOGY_SPREAD true
    run_scenario P7 S3
}

################################################################################
# PROTOCOLE P8 — Self-Healing
################################################################################
run_protocol_P8() {
    echo; echo "==============================================================="
    echo "PROTOCOLE P8 - Self-Healing"
    echo "==============================================================="; echo

    set_default_configuration

    # Charge fixe
    set_parameter DEPLOYMENTS 72
    set_parameter REPLICAS_PER_DEP 4
    set_parameter POD_LIFECYCLE_ENABLED true
    set_parameter DELETE_PODS_RANDOMLY true

    # S1
    SCENARIO_DURATION=600
    set_parameter DELETE_PODS_DELAY 60
    run_scenario P8 S1

    # S2
    SCENARIO_DURATION=600
    set_parameter DELETE_PODS_DELAY 6
    run_scenario P8 S2

    # S3
    SCENARIO_DURATION=600
    set_parameter DELETE_PODS_DELAY 1
    run_scenario P8 S3
}

################################################################################
# PROTOCOLE P9 — Crash d'un Worker
################################################################################
run_protocol_P9() {
    echo; echo "==============================================================="
    echo "PROTOCOLE P9 - Crash d'un Worker"
    echo "==============================================================="; echo

    set_default_configuration

    SCENARIO_DURATION=1200

    set_parameter DEPLOYMENTS 72
    set_parameter REPLICAS_PER_DEP 4

    echo
    echo "###########################################################"
    echo "Arrêter manuellement le kubelet du Worker après stabilisation"
    echo "Le scénario se poursuivra pendant 20 minutes."
    echo "###########################################################"
    echo

    run_scenario P9 S1
}

################################################################################
# PROTOCOLE P10 — Pression API
################################################################################
run_protocol_P10() {
    echo; echo "==============================================================="
    echo "PROTOCOLE P10 - Pression API"
    echo "==============================================================="; echo

    set_default_configuration

    set_parameter APISERVER_ETCD_ENABLED true
    set_parameter AGGRESSIVE_RECONCILE true
    set_parameter FREQUENT_STATUS_UPDATES true

    # S1
    SCENARIO_DURATION=600
    set_parameter RECONCILE_REQUEUE_DELAY 30
    run_scenario P10 S1

    # S2
    SCENARIO_DURATION=600
    set_parameter RECONCILE_REQUEUE_DELAY 10
    run_scenario P10 S2

    # S3
    SCENARIO_DURATION=600
    set_parameter RECONCILE_REQUEUE_DELAY 1
    run_scenario P10 S3
}

################################################################################
# PROTOCOLE P11 — Charge combinée
################################################################################
run_protocol_P11() {
    echo; echo "==============================================================="
    echo "PROTOCOLE P11 - Charge combinée"
    echo "==============================================================="; echo

    set_default_configuration

    SCENARIO_DURATION=1200

    set_parameter DEPLOYMENTS 50
    set_parameter REPLICAS_PER_DEP 4
    set_parameter APISERVER_ETCD_ENABLED true

    set_parameter CONFIGMAPS 300
    set_parameter SECRETS 300
    set_parameter SIZE_KB_CM 300; set_parameter SIZE_KB_SEC 300

    set_parameter FREQUENT_STATUS_UPDATES true
    set_parameter AGGRESSIVE_RECONCILE true
    set_parameter RECONCILE_REQUEUE_DELAY 10

    run_scenario P11 S1
}

################################################################################
# PROTOCOLE P12 — Stress du Controller Manager
################################################################################
run_protocol_P12() {
    echo; echo "==============================================================="
    echo "PROTOCOLE P12 - Controller Manager Stress"
    echo "==============================================================="; echo

    set_default_configuration

    # Configuration de base
    set_parameter SCHEDULER_ENABLED true
    set_parameter CONTROLLER_MANAGER_ENABLED true

    set_parameter APISERVER_ETCD_ENABLED false
    set_parameter POD_LIFECYCLE_ENABLED false
    set_parameter OPERATOR_ENABLED false

    # Charge fixe (~80 %)
    set_parameter DEPLOYMENTS 72
    set_parameter REPLICAS_PER_DEP 4

    ###########################################################################
    # S1 - Recréation des ReplicaSets
    ###########################################################################
    SCENARIO_DURATION=600

    set_parameter RECREATE_REPLICASETS true
    set_parameter REPLICA_RECREATE_DELAY 10

    set_parameter AGGRESSIVE_GC false

    run_scenario P12 S1

    ###########################################################################
    # S2 - Garbage Collection agressive
    ###########################################################################
    SCENARIO_DURATION=600

    set_parameter RECREATE_REPLICASETS false

    set_parameter AGGRESSIVE_GC true
    set_parameter GC_DELAY 10

    run_scenario P12 S2

    ###########################################################################
    # S3 - Combinaison des deux
    ###########################################################################
    SCENARIO_DURATION=600

    set_parameter RECREATE_REPLICASETS true
    set_parameter REPLICA_RECREATE_DELAY 10

    set_parameter AGGRESSIVE_GC true
    set_parameter GC_DELAY 10

    run_scenario P12 S3
}

################################################################################
# MAIN
################################################################################
main() {
    echo
    echo "==============================================================="
    echo " Kubernetes Control Plane Benchmark"
    echo "==============================================================="
    echo

    # Vérifications
    [ -f "${AUTO_TEST}" ] || { error "Fichier ${AUTO_TEST} introuvable"; exit 1; }
    kubectl cluster-info >/dev/null 2>&1 || { error "Cluster Kubernetes inaccessible"; exit 1; }

    # Sauvegarde de auto_test.sh
    restore_auto_test 2>/dev/null || true
    backup_auto_test

    info "Début de la campagne : $(timestamp)"
    echo

    ################################################################################
    # Liste des protocoles demandés
    ################################################################################
    if [ $# -eq 0 ]; then
        PROTOCOLS=(P1 P2 P3 P4 P5 P6 P7 P8 P9 P10 P11 P12)
    else
        PROTOCOLS=($(printf "%s\n" "$@" | sort -V | uniq))
    fi

    for P in "${PROTOCOLS[@]}"; do
        case "$P" in
            P1)  run_protocol_P1 ;;
            P2)  run_protocol_P2 ;;
            P3)  run_protocol_P3 ;;
            P4)  run_protocol_P4 ;;
            P5)  run_protocol_P5 ;;
            P6)  run_protocol_P6 ;;
            P7)  run_protocol_P7 ;;
            P8)  run_protocol_P8 ;;
            P9)  run_protocol_P9 ;;
            P10) run_protocol_P10 ;;
            P11) run_protocol_P11 ;;
            P12) run_protocol_P12 ;;
            *)
                echo "Protocole inconnu : ${P}"
                ;;
        esac

        info "Pause entre protocoles (${PROTOCOL_MARGIN}%)..."
        sleep $((90 * (100 + PROTOCOL_MARGIN) / 100))
    done

    echo
    success "Campagne terminée avec succès."
    info "Les journaux sont disponibles dans : ${LOG_DIR}"
}

main "$@"