# 📋 Protocole expérimental P1 — Variation  de charge dynamique

## Objectif

Étudier l'impact d'une augmentation progressive du nombre de Pods sur les performances du plan de contrôle Kubernetes. Ce protocole constitue la référence de l'étude et permettra de comparer les résultats obtenus lors des autres campagnes de tests.

## Paramètre étudié

Le paramètre étudié est le nombre total de Pods déployés simultanément dans le cluster.

La charge est augmentée progressivement en faisant varier le nombre de réplicas de plusieurs Deployments, tout en conservant constants l'ensemble des autres paramètres de la CR.

## Paramètre impacté

- `spec.deploymentCount`
- `spec.replicasPerDeployment`

## Composants principalement sollicités

- Scheduler
- Controller Manager
- API Server
- ETCD

## Hypothèse

L'augmentation progressive du nombre de Pods devrait entraîner une augmentation quasi linéaire de l'activité du plan de contrôle jusqu'à l'apparition d'un premier seuil de saturation.

Au-delà de ce seuil, une augmentation de la latence des traitements est attendue, suivie d'une dégradation progressive des performances des différents composants. ETCD est pressenti comme le premier composant limitant en raison de son rôle central dans le stockage des objets Kubernetes et de la synchronisation de l'ensemble du plan de contrôle.

## Campagne de tests

| Scénario | Nombre de Pods | Description |
|----------|----------------|-------------|
| S1 | 40 | Faible charge |
| S2 | 120 | Charge modérée |
| S3 | 180 | Charge intermédiaire |
| S4 | 270 | Forte charge |
| S5 | 360 | Limite expérimentale du cluster |

## Résultats attendus

- Augmentation du nombre de requêtes API
- Augmentation du nombre de requêtes ETCD
- Augmentation du nombre de décisions de scheduling
- Augmentation du nombre de réconciliations du Controller Manager
- Augmentation progressive de la latence des composants
- Apparition d'un seuil de dégradation avant la saturation complète du cluster

## Questions auxquelles répondre

- L'évolution des performances est-elle linéaire avec le nombre de Pods ?
- Quel composant présente les premiers signes de saturation ?
- Quel composant devient le facteur limitant du plan de contrôle ?
- À partir de combien de Pods les premières dégradations apparaissent-elles ?
- Les temps de réconciliation évoluent-ils de manière proportionnelle à la charge ?
- Observe-t-on une augmentation du nombre de retries ou d'erreurs avant la saturation ?
- La convergence entre l'état désiré et l'état observé reste-t-elle stable tout au long de la montée en charge ?