# ☸️ Environnements d'expériences

## Plateforme Minikube

Les expérimentations sont réalisées sur un cluster Minikube multi-nœuds.

Configuration :

- Minikube v1.38.1
- Kubernetes v1.35.1
- 1 nœud Control Plane
- 3 nœuds Workers
- Driver Docker
- Monitoring : kube-prometheus-stack

## Calibration du banc d'essai

### Premières observations sans mécanisme de limitation

Les premiers essais ont consisté à augmenter progressivement le nombre de Pods déployés sans appliquer de mécanisme de limitation des ressources. Le cluster est devenu instable à partir d'environ 360 Pods, alors que sa capacité théorique était estimée à environ 440 Pods. [ANNEXES_A_RAJOUTER](tableaux+graphiques_résultats.txt)

La saturation a tout d'abord été provoquée par un manque de mémoire disponible sur les nœuds, entraînant une augmentation importante de l'activité du plan de contrôle. Les temps de commit d'ETCD se sont fortement allongés, accompagnés d'une hausse de la latence des requêtes de l'API Server, des erreurs HTTP, du nombre de requêtes adressées à ETCD ainsi que de l'activité des workqueues du Controller Manager.

Les Pods non déployés sont restés principalement dans les états Pending, ContainerCreating ou ImagePullBackOff, tandis que les composants du plan de contrôle continuaient à tenter de converger vers l'état demandé. Les réconciliations successives de l'opérateur, associées aux mises à jour de statut des ressources, ont vraisemblablement entretenu une boucle de traitements de plus en plus coûteuse, aggravant progressivement la saturation d'ETCD.

Une fois ce niveau de saturation atteint, le cluster ne répondait pratiquement plus aux requêtes Kubernetes et a nécessité une intervention manuelle (suppression de la Custom Resource, suppression progressive des Deployments puis arrêt du cluster) afin de retrouver un fonctionnement normal.

Ces premières expérimentations ont mis en évidence la nécessité de préserver un niveau minimal de ressources pour les composants critiques du plan de contrôle et d'introduire des mécanismes de limitation afin d'éviter le blocage complet du cluster.

### Observations après mise en place d'un LimitRange

Afin de limiter les ressources consommées par chaque Pod, un LimitRange [ANNEXE_A_RAJOUTER](limitrange.yaml) a ensuite été introduit dans le namespace de test. Les premières valeurs retenues (25 mCPU, 32 MiB de mémoire demandée et 40 MiB de mémoire maximale par conteneur) étaient dimensionnées pour permettre le fonctionnement normal des applications tout en augmentant le nombre maximal de Pods pouvant être déployés.

Une première observation importante est que le LimitRange ne s'applique qu'à la création des Pods. Les Pods déjà existants conservent leurs anciennes valeurs de Requests et Limits, ce qui impose un redémarrage des Deployments (ou une recréation de la Custom Resource) pour appliquer les nouvelles contraintes.

Avec ces valeurs, le comportement global du cluster reste proche de celui observé précédemment, avec une légère augmentation de la consommation de ressources du plan de contrôle, probablement liée au fait que les ressources réservées par Pod deviennent supérieures à leur consommation réelle.

Dans un second temps, les limites ont été fortement réduites afin d'observer le comportement du cluster dans une situation de manque de ressources. Les Pods ne parviennent alors plus à démarrer correctement et passent successivement par les états ContainerCreating, Error, OOMKilled puis Terminating. Les mécanismes de réconciliation recréent continuellement les Pods manquants, générant une activité importante du Scheduler, du Controller Manager et de l'API Server.

Contrairement au scénario sans limitation, le cluster reste néanmoins opérationnel. Les composants du plan de contrôle continuent de répondre aux requêtes Kubernetes et il demeure possible d'observer ou de modifier les ressources du cluster malgré l'instabilité permanente des Pods. Ce comportement suggère que les mécanismes de limitation empêchent une consommation excessive des ressources disponibles et permettent au Control Plane de conserver une capacité minimale de fonctionnement.

Ces expérimentations montrent toutefois que le LimitRange agit uniquement sur les ressources attribuées à chaque conteneur et ne limite pas le nombre total de Pods pouvant être créés. Il constitue donc une première protection contre la saturation des ressources, mais ne permet pas à lui seul de maîtriser la charge globale appliquée au cluster.

L'ensemble  qui justifie l'utilisation d'un ResourceQuota dans les campagnes de tests suivantes.


# Limitations suite à observations

Afin de pouvoir tester le cluster minikube intégralement (jusqu'au maximum de ses ressources) nous avons dû limiter les ressources exploitables dans le namespace où son déployées les tests.

Pour éviter la saturation complète du cluster observée lors des essais préliminaires (≈360 Pods), un `ResourceQuota` est appliqué au namespace `operator-system` durant l'ensemble des campagnes de tests. [ANNEXE_A_RAJOUTER](quota.yaml)

| Paramètre | Valeur | Objectif |
|-----------|---------|----------|
| Pods max | 360 | Limiter le nombre maximal de Pods et les ressources CPU/RAM consommées |
| Requests CPU | 8 | Réservation minimale |
| Requests RAM | 11 Gi | Réservation mémoire 
| Limits CPU | 10 | Limitation CPU maximale |
| Limits RAM | 12 Gi | Limitation mémoire maximale |