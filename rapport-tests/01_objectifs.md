# 🎯 Plan de test expérimental de l'opérateur Kubernetes

## 1. Objectifs de l'étude

L'objectif de ce travail était de développer un opérateur Kubernetes capable de générer des charges contrôlées sur le plan de contrôle Kubernetes (CP K8S) afin d'analyser son comportement et celui des principaux composants de Kubernetes dans différents scénarios de fonctionnement.

L'étude ne cherche pas uniquement à produire une forte charge sur le cluster. Elle vise avant tout à comprendre comment chaque type d'opération lié à l'opérateur sollicite les composants internes de Kubernetes et à identifier les facteurs les plus limitant (type de charge, composant, etc.).

Les expérimentations devront permettre de répondre notamment à des questions comme :

- Quels composants du plan de contrôle sont les plus sensibles à l'augmentation de la charge ?
- Quels types d'opérations (lecture, écriture, suppression, recréation, mise à jour) sont les plus coûteux en terme de CPU, RAM, Requêtes générées ou Latence de traitement ? Une mise à jour de ressources existantes est-elle moins coûteuse qu'une suppression par exemple ?
- Existe-t-il un seuil à partir duquel les performances se dégradent fortement ?
- Une montée en charge progressive produit elle le même comportement qu'une création massive et instantanée de ressources ?
- L'augmentation de la taille des objets Kubernetes est-elle plus pénalisante que l'augmentation de leur nombre ?
- Quels composants deviennent limitants lorsque plusieurs mécanismes de Kubernetes sont sollicités simultanément ?

---

# 2. Principes méthodologiques

Afin de pouvoir interpréter les résultats de manière fiable, chaque campagne de tests devra respecter les principes suivants.

## Un seul paramètre variable

Une campagne de test ne devra faire varier qu'un seul paramètre à la fois (voir section paramètres)
L'ensemble des autres paramètres devra rester constant sur la durée du test afin de pouvoir attribuer les variations observées au seul paramètre étudié.

Exemple :
- augmentation uniquement du nombre de Deployments
- augmentation uniquement du nombre de ConfigMaps
- augmentation uniquement de la fréquence de réconciliation

Ce genre d'approche permettra d'isoler l'impact de chaque élément sur le plan de contrôle.

## Progression des expérimentations

Les campagnes seront organisées du scénario le plus simple vers le plus complexe.

Cette progression permettra :
	- d'obtenir une référence de fonctionnement
	- d'identifier progressivement les limites du cluster
	- d'introduire des interactions entre plusieurs composants uniquement lorsque les comportements élémentaires seront connus.

## Reproductibilité

Chaque expérimentation devra être entièrement reproductible.
Les paramètres suivants devront être enregistrés :
	- date et heure du test
	- durée
	- paramètres de l'opérateur
	- paramètres de la CRD
	- version du cluster
	- version de l'opérateur
	- métriques collectées.

L'ensemble des fichiers générés (CR, logs, captures Grafana, export Prometheus) devra être conservé et accessible avec le rapport

## Comparabilité

Chaque campagne devra pouvoir être comparée à une ou plusieurs campagnes similaires.
Les comparaisons constitueront le cœur de l'analyse.

Les principaux axes de comparaison seront notamment :
	- montée en charge progressive vs création massive instantanée
	- mise à jour vs suppression / recréation
	- nombreux petits objets vs peu de gros objets
	- modification sur cluster peu chargé vs sur cluster déjà fortement sollicité
    - faible fréquence de réconciliation vs fréquence élevée

## Collecte des métriques

Pour chaque expérimentation, les métriques devront être relevées :
	- avant le début du test (état de référence) 
	- pendant les pics d'activité
	- après stabilisation du cluster

Les mesures devront être suffisamment fréquentes afin d'obtenir des courbes exploitables et de pouvoir observer les phénomènes transitoires (nombre d'échantillons >= 100)


## Résultat attendu

À l'issue des expérimentations, il devra être possible d'établir une corrélation entre :
- les paramètres de charge appliqués par l'opérateur
- les métriques observées sur le plan de contrôle
- les limites de fonctionnement des différents composants de Kubernetes

L'ensemble des conclusions devrait permettre en effet d'identifier les scénarios les plus coûteux pour le plan de contrôle ainsi que les composants les plus sensibles aux différentes stratégies de sollicitation.