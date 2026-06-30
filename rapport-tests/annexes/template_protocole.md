La fiche suivante permet de donner le processus expérimental suivi lors de différents scenarios visant à faire varier les mêmes paramètres à différents niveaux. Elle permet de donner une direction aux tests et de centraliser le fonctionnement prévu pour un ensemble de tests.

Nous établierons ainsi plusieurs "Familles" de tests correspondant chacune à un protocole



```markdown
# 📋 Protocole expérimental n°X — <Titre du protocole>

## Objectif

<!-- 
Décrire en une ou deux phrases l'objectif de ce protocole expérimental.
-->

## Hypothèse de départ

<!-- 
Décrire l'hypothèse que ce protocole cherche à valider.
-->

## Paramètre étudié

<!-- 
Liste des paramètres variables
-->

## Paramètres constants

| Paramètre | Valeur |
|-----------|--------|
| Cluster | Minikube 1 Control Plane + 3 Workers |
| ResourceQuota | Activé |
| Monitoring | Activé |
| Image | nginx |
| CPU / RAM des Pods | Valeurs par défaut |
| Tous les autres paramètres | Inchangés | 


## Paramètres variables

<!-- | Paramètre | Valeurs étudiées |
|-----------|------------------|
| Nombre de [Charge] | |
| [Paramètre charge]| |
| Intervalle update CR | À définir |
| Durée test | À définir | -->

## Déroulement

1. Vérifier l'état initial du cluster
2. Déployer la Custom Resource
3. Observer les métriques
4. Exporter les graphiques Grafana
5. Attendre la stabilisation du cluster
6. Arrêter le scénario
7. Sauvegarder les résultats

# Métriques observées

### API Server

- Requêtes API/s
- Latence p99
- Erreurs HTTP
- Requêtes simultanées

### ETCD

- Requêtes/s
- Backend Commit
- WAL fsync
- Taille de la base

### Scheduler

- Tentatives de scheduling
- Durée de scheduling
- Pods en attente

### Controller Manager

- Workqueue Adds
- Workqueue Depth
- Retries

### Opérateur

- Réconciliations
- Temps de réconciliation
- Erreurs

### Cluster

- CPU
- RAM
- Nombre de charge

## Critères d'arrêt

<!--
Le protocole est arrêté lorsqu'au moins un des critères suivants est atteint :

- Durée prévue atteinte
- Nombre maximum de la charge atteinte (prédéfinie au départ)
- Timeout  (blocage du cluster)
- Erreurs empêchant la poursuite du scénario 
-->

## Relevés à effectuer

- Fichier de configuration utilisé
- Captures Grafana pertinentes (graphiques / tableaux)
- Comportement de l'opétaeur et logs si anomalies
- Evènements Kubernetes intéressants (ex. de la charge étudiée, des composants, etc.)
- Eventuellement description via `kubectl describe` si pertinente

## Questions auxquelles répondre

- Quel composant est sollicité en premier ?
- Quel composant devient limitant ?
- À partir de quel seuil la dégradation apparaît-elle ?
- La montée en charge est-elle linéaire ?
- Existe-t-il un effet de seuil ?
- Les performances reviennent-elles à l'état initial après stabilisation ?
- A quelle moment atteint-on la convergence état actuel vers état désiré ?

## Scénarios associés

Ce protocole pourra être décliné selon plusieurs scénarios présentés dans `09_scenarios.md`.

| Scénario | Description |
|----------|-------------|
| S1 | À compléter |
| S2 | À compléter |
| S3 | À compléter |
```