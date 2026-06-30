# 🧪 Scénario S01 — Montée en charge progressive des Pods

## Protocole associé

Protocole expérimental n°1 — Influence du nombre de Pods

---

## Objectif

Étudier l'évolution des performances du plan de contrôle lors d'une augmentation progressive du nombre de Pods.

---

## Paramètres de la campagne

| Paramètre | Valeur |
|-----------|--------|
| Nom de la campagne | D10-R1-30s-180min |
| Date | 25/06/2026 |
| Durée | 180 min |
| Intervalle | 30 s |
| Condition d'arrêt | 360 Pods ou timeout |

---

## Paramètres variables

| Paramètre | Valeur |
|-----------|--------|
| Deployments | 10 |
| Réplicas par Deployment | 1 → 36 |
| Incrément | +1 Pod toutes les 30 s |

---

## Paramètres constants

- Cluster Minikube (1 CP + 3 Workers)
- ResourceQuota activé
- Monitoring déplacé sur le Control Plane
- Image Nginx
- Scheduler Stress désactivé
- Controller Manager Stress désactivé

---

## Configuration `auto_test.sh`

```bash
DEPLOYMENTS=10
REPLICAS_PER_DEP=1
CONFIGMAPS=0
SECRETS=0

INTERVAL=30
DURATION=10800

SCHEDULER_ENABLED=false
APISERVER_ETCD_ENABLED=false
CONTROLLER_MANAGER_ENABLED=false
```

---

## Ressource `ControlPlaneTest`

```yaml
spec:

  deploymentCount: 10

  replicasPerDeployment: 1

  configMapCount: 0

  secretCount: 0
```

---

## Déroulement

- Déploiement de la CR
- Vérification des Deployments
- Observation des métriques Grafana
- Capture des courbes toutes les 10 minutes
- Arrêt à 360 Pods ou après 180 minutes

---

## Relevés à effectuer

- CPU du cluster
- RAM du cluster
- Latence API Server
- Backend Commit ETCD
- Scheduler
- Controller Manager
- Nombre de Pods
- Évènements Kubernetes

---

## Résultats observés

À compléter après expérimentation.

---

## Conclusion

À compléter après analyse.