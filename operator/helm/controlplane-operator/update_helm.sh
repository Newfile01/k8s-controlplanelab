```bash
#!/bin/bash

set -e

############################################################
# CONFIGURATION
############################################################

CHART_PATH_FILE="$HOME/.helm_chart_path"
DEFAULT_CHART_PATH="/home/nico/k8s-controlplanelab/operator/helm/controlplane-operator"

GHCR_REGISTRY="ghcr.io"
GHCR_REPO="oci://ghcr.io/newfile01/charts"

############################################################
# INITIALISATION DU CHEMIN
############################################################

if [ ! -f "$CHART_PATH_FILE" ]; then
    echo "[INFO] Aucun chemin Helm enregistré"
    echo "[INFO] Enregistrement du chemin par défaut :"
    echo "       $DEFAULT_CHART_PATH"

    echo "$DEFAULT_CHART_PATH" > "$CHART_PATH_FILE"
fi

CHART_PATH=$(cat "$CHART_PATH_FILE")

############################################################
# VERIFICATION DOSSIER
############################################################

if [ ! -d "$CHART_PATH" ]; then
    echo "[ERREUR] Dossier Helm introuvable :"
    echo "          $CHART_PATH"
    exit 1
fi

############################################################
# POSITIONNEMENT DANS LE DOSSIER
############################################################

cd "$CHART_PATH"

echo "============================================================"
echo "HELM CHART"
echo "============================================================"
echo "[INFO] Dossier courant :"
echo "       $(pwd)"

############################################################
# RECUPERATION VERSION ACTUELLE
############################################################

CHART_VERSION=$(grep '^version:' Chart.yaml | awk '{print $2}')
APP_VERSION=$(grep '^appVersion:' Chart.yaml | awk '{print $2}' | tr -d '"v')

echo
echo "============================================================"
echo "VERSION ACTUELLE"
echo "============================================================"

echo "Chart version : $CHART_VERSION"
echo "App version   : v$APP_VERSION"

############################################################
# INCREMENTATION VERSION
############################################################

echo
echo "============================================================"
echo "INCREMENTATION VERSION"
echo "============================================================"

read -p "Incrementer version chart ? (y/n) : " INC_CHART

if [ "$INC_CHART" = "y" ]; then

    MAJOR=$(echo "$CHART_VERSION" | cut -d. -f1)
    MINOR=$(echo "$CHART_VERSION" | cut -d. -f2)
    PATCH=$(echo "$CHART_VERSION" | cut -d. -f3)

    PATCH=$((PATCH + 1))

    NEW_CHART_VERSION="$MAJOR.$MINOR.$PATCH"

    sed -i "s/^version: .*/version: $NEW_CHART_VERSION/" Chart.yaml

    CHART_VERSION="$NEW_CHART_VERSION"

    echo "[OK] Nouvelle chart version : $CHART_VERSION"
fi

read -p "Incrementer appVersion ? (y/n) : " INC_APP

if [ "$INC_APP" = "y" ]; then

    APP_VERSION=$((APP_VERSION + 1))

    sed -i "s/^appVersion: .*/appVersion: \"v$APP_VERSION\"/" Chart.yaml

    echo "[OK] Nouvelle appVersion : v$APP_VERSION"
fi

############################################################
# NETTOYAGE ANCIENS PACKAGES
############################################################

echo
echo "============================================================"
echo "SUPPRESSION DES .TGZ"
echo "============================================================"

rm -f ./*.tgz

echo "[OK] Packages supprimés"

############################################################
# HELM LINT
############################################################

echo
echo "============================================================"
echo "HELM LINT"
echo "============================================================"

if ! helm lint .; then
    echo
    echo "[ERREUR] helm lint a échoué"
    exit 1
fi

echo
echo "[OK] helm lint valide"

############################################################
# PACKAGE HELM
############################################################

echo
echo "============================================================"
echo "PACKAGE HELM"
echo "============================================================"

helm package .

PACKAGE_NAME=$(ls *.tgz)

echo
echo "[OK] Package généré :"

ls -lh *.tgz

############################################################
# PUSH REGISTRE OCI
############################################################

echo
echo "============================================================"
echo "PUSH OCI"
echo "============================================================"

helm push "$PACKAGE_NAME" "$GHCR_REPO"

echo
echo "[OK] Push terminé"

############################################################
# RESUME
############################################################

echo
echo "============================================================"
echo "RESUME"
echo "============================================================"

echo "Chart version : $CHART_VERSION"
echo "App version   : v$APP_VERSION"
echo "Package       : $PACKAGE_NAME"

echo
echo "[OK] TERMINE"
```
