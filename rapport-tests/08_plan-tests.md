# 🧪 Protocoles expérimentaux

Les expérimentations sont organisées en plusieurs protocoles, chacun visant à solliciter préférentiellement un ou plusieurs composants du plan de contrôle Kubernetes.

Chaque protocole est décliné en plusieurs scénarios correspondant à différents niveaux de charge. Les conclusions seront établies à partir de la comparaison de ces différents paliers.

## Légende

- **D** : Deployments
- **R** : Réplicas par Deployment
- **CM** : ConfigMaps
- **SEC** : Secrets
- **SZ** : Taille unitaire d'un objet
- **Δt** : Intervalle entre deux opérations (**10 s**)
- **Progressif** : La charge augmente par incréments successifs toutes les **10 s**
- **Burst** : La charge cible est appliquée immédiatement, sans montée progressive

> [!NOTE]
> Le banc d'essai est calibré sur une limite expérimentale d'environ **360 Pods** (≈20 mCPU et 30 Mi par Pod). Les scénarios progressifs utilisent un intervalle fixe de **10 secondes**, identique à l'intervalle de collecte Prometheus. Afin de conserver des campagnes de tests de durée raisonnable, le nombre d'itérations est limité à **100 modifications maximum**. Au-delà, l'incrément est automatiquement augmenté.


| Protocole                                   | Paramètre étudié                                   | Paramètre impacté                                                                                    | Composants principalement sollicités            | Paliers                                                  | Mode d'application                                 | Durée max / palier                                   | Durée totale                | Conclusion recherchée                                                         |
| ------------------------------------------- | -------------------------------------------------- | ---------------------------------------------------------------------------------------------------- | ----------------------------------------------- | -------------------------------------------------------- | -------------------------------------------------- | ---------------------------------------------------- | --------------------------- | ----------------------------------------------------------------------------- |
| **P1 - Variation du nombre de Deployments** | Nombre de Deployments (R=5)                        | `deploymentCount`                                                                                    | API Server, ETCD, Controller Manager, Scheduler | **D=8 → D=36 → D=72** (40 → 180 → 360 Pods)              | **Progressif** (+1 D / 10 s)                       | **1m30 → 6m40 → 13m20** (90 s → 400 s → 800 s)       | **≈22 min** (1320 s)        | Influence du nombre de Deployments sur le Control Plane                       |
| **P2 - Variation du nombre de Pods**        | Nombre de réplicas (D=10)                          | `replicasPerDeployment`                                                                              | Scheduler, Controller Manager, API Server, ETCD | **R=4 → R=18 → R=36** (40 → 180 → 360 Pods)              | **Progressif** (+1 / +2 / +4 R / 10 s)             | **50 s → 5 min → 10 min** (50 s → 300 s → 600 s)     | **≈16 min** (950 s)         | Influence de la montée en charge des Pods                                     |
| **P3 - Variation de la charge statique**    | Nombre de ConfigMaps et Secrets                    | `configMapCount`, `secretCount`                                                                      | API Server, ETCD                                | **50+50 → 250+250 → 1000+1000**                          | **Progressif** (+1 / +5 / +50 CM+SEC / 10 s)       | **10 min → 10 min → 16m40** (600 s → 600 s → 1000 s) | **≈37 min** (2200 s)        | Influence du nombre d'objets stockés                                          |
| **P4 - Variation du volume des objets**     | Taille des ConfigMaps et Secrets                   | `configMapSizeKB`, `secretSizeKB`                                                                    | API Server, ETCD                                | **10×500 Ko → 10×5 Mo → 10×50 Mo**                       | **Burst (80 % de charge)**                         | **10 min** (600 s)                                   | **30 min** (1800 s)         | Impact du volume de données stockées                                          |
| **P5 - Type d'opérations**                  | Création, mise à jour, suppression                 | `recreateResources`                                                                                  | API Server, ETCD, Controller Manager            | **Create → Update → Delete/Recreate**                    | **Burst (80 % de charge)**                         | **10 min** (600 s)                                   | **30 min** (1800 s)         | Comparer le coût des différents types d'opérations                            |
| **P6 - Fréquence de réconciliation**        | Intervalle entre deux réconciliations              | `reconcileRequeueDelay`                                                                              | Opérateur, API Server, ETCD                     | **Δt = 30 s → 10 s → 1 s**                               | **Variation de Δt**                                | **10 min** (600 s)                                   | **30 min** (1800 s)         | Influence de la fréquence des écritures                                       |
| **P7 - Complexité du Scheduler**            | Stratégie de placement                             | `nodeSelector`, `affinityMode`, `antiAffinityMode`, `topologySpread`                                 | Scheduler                                       | **NodeSelector → Affinity → TopologySpread** (≈288 Pods) | **Burst (80 % de charge)**                         | **10 min** (600 s)                                   | **30 min** (1800 s)         | Impact de la complexité de planification                                      |
| **P8 - Self-Healing**                       | Suppression aléatoire de Pods                      | `deletePodsRandomly`, `deletePodsDelay`                                                              | Controller Manager, Scheduler                   | **1 → 10 → 50 Pods/min**                                 | **Progressif** (+1 / +10 / +50 Pods/min)           | **10 min** (600 s)                                   | **30 min** (1800 s)         | Évaluer les mécanismes de convergence                                         |
| **P9 - Crash d'un Worker**                  | Arrêt d'un nœud                                    | Arrêt manuel du `kubelet`                                                                            | Scheduler, Controller Manager                   | **288 Pods (≈80 % de charge)**                           | **Burst (80 % de charge)**                         | **20 min** (1200 s)                                  | **20 min** (1200 s)         | Temps de convergence après perte d'un nœud                                    |
| **P10 - Pression API**                      | Fréquence des Status Update                        | `frequentStatusUpdates`                                                                              | API Server, ETCD                                | **Δt = 30 s → 10 s → 1 s**                               | **Variation de Δt**                                | **10 min** (600 s)                                   | **30 min** (1800 s)         | Influence des écritures fréquentes sur la latence                             |
| **P11 - Charge combinée**                   | Charge dynamique + charge statique + Status Update | `deploymentCount`, `replicasPerDeployment`, `configMapCount`, `secretCount`, `frequentStatusUpdates` | Ensemble du Control Plane                       | **288 Pods + 4000 CM + 4000 SEC**                        | **Progressif** (+3 Pods / +40 CM / +40 SEC / 10 s) | **20 min** (1200 s)                                  | **20 min** (1200 s)         | Identifier le premier composant limitant et les interactions entre composants |
| **Total**                                   | **11 protocoles**                                  | **11 paramètres étudiés**                                                                            | **5 composants observés**                       | **29 scénarios**                                         | **Mixte (Progressif + Burst)**                     | **≈ 3 h 37 min**                                     | **≈ 4 h 49 min (17 340 s)** |                                                                               |




## Organisation des scénarios

Chaque protocole est décliné en plusieurs scénarios correspondant à des niveaux de charge croissants.

À titre d'exemple, un protocole comportant cinq paliers pourra être organisé de la manière suivante :


| Scénario | Niveau de charge     |
| -------- | -------------------- |
| S1       | Faible charge        |
| S2       | Charge modérée       |
| S3       | Charge intermédiaire |
| S4       | Forte charge         |
| S5       | Limite du cluster    |


Cette progression permet d'identifier l'apparition des premiers phénomènes de dégradation, leur évolution ainsi que le seuil à partir duquel le plan de contrôle ne parvient plus à maintenir un fonctionnement nominal.

trois grands axes scientifiques qui se dégagent naturellement :

1. Faire varier la quantité de travail

- nombre de Pods
- nombre de ConfigMaps
- nombre de Secrets
- nombre de Deployments

1. Faire varier la nature du travail

- Create
- Update
- Delete
- Delete/Recreate
- Garbage Collection
- Status Update

1. Faire varier le contexte d'exécution

- Crash d'un nœud
- Scheduler complexe (Affinity, TopologySpread)
- Sous-charge
- Surcharge
- Charges combinées

