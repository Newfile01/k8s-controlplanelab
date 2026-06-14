# Construction d'un dashboard Grafana & compréhension des PromQL

## 🎯 Présentation

Cette étape introduit :
- Grafana
- organisation dashboards
- visualisations
- PromQL
- compréhension des métriques Kubernetes

But :
- superviser le control-plane Kubernetes
- visualiser les stress tests
- comprendre le comportement interne du cluster
- apprendre à interpréter les métriques Kubernetes et Prometheus
---

## 🏗️ Architecture & emplacement de travail

* Architecture supervision

```text
Kubernetes Components
→ Exporters / Metrics endpoints
→ Prometheus : collecte des métriques
→ Grafana : visualisation
```

* Structure Grafana

Dashboard
→ Tabs
→ Rows
→ Panels

Panels créés :
- API CPU
- API RAM
- API QPS
- ETCD latency
- Node pressure
- Network
- Disk IO

Onglets créés
- OVERVIEW
- CONTROL-PLANE

### Types de visualisations utilisées

| Métrique | Visualisation recommandée | But |
|---|---|---|
| CPU | TimeSeries | évolution consommation |
| RAM | TimeSeries | pression mémoire |
| QPS | Stat + TimeSeries | débit API |
| ETCD latency | TimeSeries | observer latence |
| Node pressure | Stat / Gauge | état cluster |
| Network | TimeSeries | trafic réseau |
| Disk IO | TimeSeries | activité disque |

---

## 💡 Compréhension PromQL

PromQL repose principalement sur :
- des métriques temporelles
- des fonctions de transformation
- des agrégations

### Fonctions étudiées

#### Rate(<metric>)[time]

- `rate()[Xm]` :
    - transforme un compteur cumulatif en débit moyen par seconde
    - calculé sur une fenêtre glissante
    Ex. Nombre de requêtes reçues par l'api-server cumulées depuis le début
        => rate() [5m] 
        => (transforme en)
        Chaque valeur de la requête présentera :
        Requêtes api-server effectuées depuis les 5 dernières minutes avant la présentation de cette valeur
        >> On observera un taux de requêtes moyen variant sur les 5 dernières minutes


Exemple :

```text
Compteur brut :
100 → 200 → 300 → 400

rate(...[5m])
→ calcule :
variation du compteur / temps écoulé
Si +100 toutes les 5 minutes => on aura rate(compteur[5m]) = 100
```

⚠️ Important

Les compteurs Prometheus :
- augmentent continuellement depuis le démarrage du composant
- ne représentent PAS directement une activité instantanée

`rate()` permet donc :
- d'obtenir une activité récente
- de lisser les variations
- de transformer un compteur brut en débit exploitable

🧪 Exemple API Server

```text
Nombre total de requêtes reçues depuis le démarrage
→ rate(...[5m])
→ nombre moyen de requêtes/seconde sur les 5 dernières minutes
```

=> Ce qu'on observe

```text
Timeline du débit moyen
→ recalculée en permanence
→ sur les 5 minutes précédentes
```

---

#### Sum(<metric>)

- `sum()` :
    Additionne l'ensemble des compteurs/métriques retournées
    Ex. Nombre de requêtes reçues par un api-server cumulées depuis le début (Pod isolé)
        => sum() 
        => (transforme en)
        Chaque valeur de la requête présentera :
        Nombre TOTAL des requêtes reçues par l'ensemble des api-server du cluster (aggregation/regroupement des Pods)
        >> On observera l'ensemble des requêtes traitées par le composant API-SERVER sur l'ensemble du cluster (vue macro)

`sum()` :
- additionne plusieurs séries temporelles
- permet une vision globale

Exemple :

```text
API Server Pod 1 → 10 req/s
API Server Pod 2 → 15 req/s

sum(...)
→ 25 req/s
```

🗃️ Utilité

Kubernetes fonctionne souvent :
- avec plusieurs Pods
- plusieurs noeuds
- plusieurs instances d'un composant

`sum()` permet donc :
- d'obtenir une vue cluster globale
- d'agréger les métriques distribuées

🔎 Ce qu'on observe

```text
Vision macro du composant Kubernetes
```

---

#### Histogram_quantile(<metric>)

- `histogram_quantile()` :
    Transforme un ensemble de données distribuées en Histogramme (plusieurs buckets regroupant des sous-ensembles de ces données)
    Quantile permettra de définir le seuil des données (toutes les données observées seront inférieur à ce seuil)

🧬 Principe

Les histogrammes Prometheus :
- répartissent les valeurs dans des "buckets"

Exemple :

```text
≤ 1ms
≤ 5ms
≤ 10ms
≤ 50ms
≤ 100ms
```

Chaque bucket contient :
- le nombre de valeurs inférieures au seuil

🍰 Fonctionnement du quantile

```promql
histogram_quantile(0.99, ...)
```

Signifie :

```text
99% des valeurs
≤ valeur affichée
```

⚠️ Important

Le quantile :
- ne calcule PAS une moyenne
- estime un percentile statistique

Exemple :

```text
p99 = 80ms

→ 99% des opérations
≤ 80ms
```

---

### Exemple CPU API Server

```promql
sum(rate(container_cpu_usage_seconds_total{
pod=~"kube-apiserver.*"
}[5m]))
```

`container_cpu_usage_seconds_total`

Compteur cumulatif :
- temps CPU consommé depuis le démarrage du conteneur

Important :
- il ne s'agit PAS d'un pourcentage CPU
- mais d'un temps CPU cumulé

Exemple :

```text
+1 seconde CPU
→ correspond à :
1 coeur entièrement occupé pendant 1 seconde
```

📖 Lire : 

```text
"Ensemble des temps CPU du ou des pods api-server (`sum(...)`) recalculés sur une fenêtre de 5 minutes glissante (`rate(...)..[5m]`) en permanence"

OU

"Temps CPU moyen consommé par l'ensemble des API Servers sur les 5 dernières minutes"
```
=> Permet de visualiser la timeline de la consommation CPU de l'api-server sur l'ensemble du cluster (si plujsieurs instances) 

`rate(...)` transforme :
- le temps CPU cumulé → en consommation CPU moyenne récente

`sum(...)` additionne :
- tous les Pods API Server



### Exemple mémoire API Server

```promql
sum(container_memory_working_set_bytes{
pod=~"kube-apiserver.*"
}) / 1024 / 1024
```

Contrairement au CPU :
- la mémoire n'est PAS un compteur cumulatif
- c'est une mesure instantanée

`container_memory_working_set_bytes` représente :
- mémoire réellement utilisée
- mémoire active/non reclaimable

Cela ne représente PAS :
- des écritures mémoire
- des accès ETCD
- des allocations cumulées

`sum(...)` additionne :
- la mémoire utilisée par tous les Pods API Server

💡 Conversion

```text
/1024/1024
→ conversion bytes → MB
```


📖 Lire :

```text
"Mémoire utilisée par tous les pods api-server du cluster (`sum(...)` regroupe les métriques d'écriture mémoire de chaque conteneur)"

OU

"Mémoire réellement utilisée par l'ensemble des API Servers à l'instant de mesure"
```

=> Permet d'observer pression mémoire du control-plane (ici cumulée depuis le début)

---

### Exemple percentile ETCD

```promql
histogram_quantile(
0.99,
sum(rate(etcd_disk_backend_commit_duration_seconds_bucket[5m])) by (le)
)
```

`etcd_disk_backend_commit_duration_seconds_bucket`

Histogramme des durées :
- des commits ETCD
- sur le backend disque BoltDB

Un commit ETCD implique généralement :

```text
API Server
→ transaction ETCD
→ BoltDB
→ fsync disque
→ commit
```

`rate(...[5m])` transforme :
- les compteurs histogrammes cumulés
→ en distribution récente des latences

⚠️ Important :
- on ne transforme PAS la métrique en "nombre d'écritures"
- on actualise l'histogramme
- sur les 5 dernières minutes

`sum(...) by (le)` additionne :
- tous les Pods ETCD
- tout en conservant les buckets (`le`)

`histogram_quantile(0.99, ...)` reconstruit l'histogramme :
- puis estime le percentile p99


Lire : 


```text
"Histogramme des durées de commits ETCD (répartition des durées d'écritures en plusieurs buckets), convertis en 'nombre de requêtes par secondes d'écriture etcd sur les 5 dernières minutes' (toujours sous forme d'histogramme) (`rate(...)..[5m]`), on récupère ces données pour l'ensemble des pods etcd (`sum(...)`), le tout reconverti en histogramme présentant les valeurs inférieures au p99 (`histogram_quantile()`)"

OU

Durée sous laquelle 99% des commits ETCD ont été effectués sur les 5 dernières minutes
```

=> Permet d'observer le durée moyenne d'écriture de l'ETCD (un commit complet) sur une fenêtre glissante de 5 minutes précédent chaque valeur affichée

Latence disque ETCD
→ excellent indicateur santé control-plane


---

## 🔎 Vérifications

* Vérifier dans Grafana : Dashboards
* Vérifier métriques Prometheus : UP
* Vérifier targets : Prometheus > Targets

### Observations attendues

Les dashboards doivent :
- recevoir des données
- afficher des timelines cohérentes
- réagir aux stress tests

Les targets Prometheus doivent être :
- UP
- sans erreur de scrape

---

## ✅ Bilan

Le cluster dispose désormais :
- dashboards Grafana structurés
- supervision control-plane
- visualisation des stress tests
- compréhension approfondie PromQL
- compréhension des métriques Kubernetes

Les notions suivantes sont maintenant maîtrisées :
- métriques cumulatives
- séries temporelles
- agrégations
- histogrammes
- percentiles
- fenêtres glissantes
- débit moyen
- interprétation des métriques control-plane

Cette étape permet désormais :
- d'observer le comportement réel du cluster
- d'interpréter les effets des stress tests
- d'identifier les goulets d'étranglement Kubernetes