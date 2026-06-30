# ⚙️ Paramètres expérimentaux

## Objectif

Les campagnes de tests reposent sur un ensemble de paramètres permettant de contrôler précisément la charge appliquée au Control Plane Kubernetes.

Chaque scénario présenté dans la suite de ce document correspondra à une combinaison de ces paramètres.

---

## 📈 Paramètres de charge

Ces paramètres permettent de faire varier le nombre de ressources manipulées par le cluster.

| Paramètre | Impact principal | Composants principalement sollicités |
|-----------|------------------|--------------------------------------|
| Nombre de Deployments | Nombre d'applications déployées | API Server, Controller Manager, Scheduler, ETCD |
| Nombre de Pods | Charge de planification | Scheduler, Controller Manager, ETCD |
| Nombre de ConfigMaps | Charge de stockage | API Server, ETCD |
| Nombre de Secrets | Charge de stockage | API Server, ETCD |

## 📦 Paramètres de volumétrie

Ces paramètres permettent de faire varier la taille des objets Kubernetes sans modifier leur nombre.

| Paramètre | Impact principal | Composants principalement sollicités |
|-----------|------------------|--------------------------------------|
| Taille des ConfigMaps | Volume d'écriture | API Server, ETCD |
| Taille des Secrets | Volume d'écriture | API Server, ETCD |
| Nombre de labels | Taille des métadonnées | API Server, ETCD |
| Nombre d'annotations | Taille des métadonnées | API Server, ETCD |
| Taille des manifests | Taille des requêtes API | API Server, ETCD |

##⏱️ Paramètres temporels

Ces paramètres contrôlent la vitesse d'application de la charge.

| Paramètre | Impact principal | Composants principalement sollicités |
|-----------|------------------|--------------------------------------|
| Durée totale du test | Temps d'observation | Tous |
| Intervalle entre deux créations | Débit de création | Tous |
| Intervalle entre deux suppressions | Débit de suppression | Tous |
| Fréquence de réconciliation | Activité de l'opérateur | Operator, API Server |
| Temps d'attente entre deux incréments | Progressivité de la montée en charge | Tous |

## 🔄 Paramètres de comportement

Ces paramètres définissent la nature des opérations réalisées sur le cluster.

| Paramètre | Impact principal | Composants principalement sollicités |
|-----------|------------------|--------------------------------------|
| Création | Génération de nouvelles ressources | Tous |
| Mise à jour | Modification de ressources existantes | API Server, ETCD |
| Suppression | Suppression de ressources | API Server, Controller Manager, ETCD |
| Suppression / Recréation | Churn important | Tous |
| Scale Up | Augmentation des réplicas | Scheduler, Controller Manager |
| Scale Down | Réduction des réplicas | Controller Manager |
| Burst | Création instantanée | Tous |
| Montée progressive | Création incrémentale | Tous |

---



# 🔗 Correspondance avec la CRD


# 🖥️ Correspondance avec `auto_test.sh`

Cette section établira la correspondance entre les paramètres expérimentaux et les variables utilisées par le script de lancement des campagnes de tests.

> [!NOTE]
> Chaque scénario du document `07_scenarios.md` sera obtenu en affectant une valeur à ces paramètres.