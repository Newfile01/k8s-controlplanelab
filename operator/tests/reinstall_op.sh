#!/usr/bin/bash

CHART_VERSION="0.1.10"
APP_VERSION="v15"
OCI_REGISTRY="ghcr.io/newfile01/charts"

#### SUPPRESSION CRs EN COURS

# Vérifier qu'il n'y a plus de CR active, si OUI les supprimer (c'est mieux)
kubectl get controlplanetests.controlplane.lab.local -A

# Supprimer les CR en cours en décommentant la ligne suivante avec le bon nom
# kubectl delete controlplanetests.controlplane.lab.local <nom_CR>



#### DESINSTALLATION/REINSTALLATION HELM CHART

helm list -A

# Le nom de la release "controlplane-operator" doit correspondre au champ "NAME" dans la commande précédente
helm uninstall controlplane-operator

kubectl delete crd controlplanetests.controlplane.lab.local

helm install controlplane-operator \
  oci://${OCI_REGISTRY}/controlplane-operator \
  --version ${CHART_VERSION} \
  -n operator-system \
  --create-namespace