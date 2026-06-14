# Correctif Minikube pour scraping ETCD via Prometheus Operator

## 🎯 Présentation

Cette étape corrige le scraping ETCD sous Minikube.

Problème rencontré :
- ETCD écoute uniquement sur `127.0.0.1:2381`
- Prometheus tente de scraper l'IP du noeud
- connexion refusée

Conséquence :
- target `monitoring-stack-kube-etcd` reste DOWN
- absence de métriques ETCD dans Prometheus/Grafana

But de cette étape :
- rendre les métriques ETCD accessibles à Prometheus Operator
- permettre le scraping ETCD depuis le cluster

---

## 🏗️ Architecture & emplacement de travail

Dans le cluster Minikube démarré il suffit d'observer puis corriger les éléments configurés en dur dans le manifest static pod ETCD.

### Architecture initiale

```text
Prometheus
→ NodeIP:2381
→ refused
```

### Architecture corrigée

```text
Prometheus
→ ETCD metrics endpoint
→ scrape OK
```

### Fonctionnement important à comprendre

ETCD est déployé sous forme de :
- static pod Kubernetes

Le manifest est directement stocké sur le noeud :

```text
/etc/kubernetes/manifests/etcd.yaml
```

Le kubelet surveille automatiquement ce répertoire :
- toute modification du fichier
→ redémarre automatiquement le Pod ETCD

---

## 🔎 Diagnostic

* Diagnostic endpoint ETCD : `kubectl describe pod etcd-control-plane-lab -n kube-system`
* Constat : `--listen-metrics-urls=http://127.0.0.1:2381`

```bash
# Vérification écoute locale
minikube ssh -p control-plane-lab
sudo ss -lntp | grep 2381

# Vérification endpoint métriques
curl http://127.0.0.1:2381/metrics
```

Ce que montrent ces vérifications :
- ETCD écoute uniquement sur `127.0.0.1`
- donc uniquement accessible depuis le noeud lui-même
- Prometheus scrape les NodeIP Kubernetes
- aucune route réseau possible vers localhost ETCD

Observation typique :

```text
LISTEN 127.0.0.1:2381
```

On remarque donc que :
- ETCD expose le port 2381 uniquement au localhost du noeud
- aucune exposition réseau externe
- Prometheus scrape les NodeIP du cluster

Conséquence :

```text
Pas de liaison Prometheus ↔ ETCD
```

---

## ⚙️ Modification du manifest ETCD

### Correction rapide

```bash
## CORRECTION ETCD TARGET POUR PROMETHEUS

# Entrer dans le noeud Minikube
minikube ssh -p control-plane-lab

# Modifier le listen-metrics ETCD
sudo sed -i \
's#127.0.0.1:2381#0.0.0.0:2381#g' \
/etc/kubernetes/manifests/etcd.yaml
```

### Explication

Modification effectuée :

```text
127.0.0.1:2381
→
0.0.0.0:2381
```

Cela signifie :
- ETCD écoute désormais sur toutes les interfaces réseau
- le port devient accessible depuis le cluster
- Prometheus peut maintenant joindre ETCD

Important :
- cette modification est spécifique au lab Minikube
- en production ETCD ne doit généralement PAS exposer ses métriques publiquement sans sécurisation adaptée

---

## 🔎 Vérifications

```bash
# Vérifier la modification
sudo cat /etc/kubernetes/manifests/etcd.yaml | grep listen-metrics

# Vérifier nouvelle écoute réseau
sudo ss -lntp | grep 2381
```

### Observation attendue

```yaml
--listen-metrics-urls=http://0.0.0.0:2381
```

### Observation socket attendue

```text
LISTEN 0.0.0.0:2381
```

### Redémarrage automatique ETCD

Le kubelet redémarre automatiquement le static pod ETCD après modification du manifest.

Aucune commande Kubernetes supplémentaire nécessaire.

### Vérification Prometheus

Dans :

```text
Prometheus > Status > Targets
```

Target attendue :

```text
kube-etcd UP
```

### Vérification Grafana

Les dashboards ETCD doivent maintenant afficher :
- métriques temps réel
- histogrammes
- percentiles
- latences
- activité base clé/valeur

---

## ⚠️ Limites de cette approche

Cette correction :
- modifie directement le noeud Minikube
- n'est pas persistante après recréation cluster
- dépend fortement de l'implémentation Minikube

Sur un vrai cluster Kubernetes :
- ETCD est généralement mieux sécurisé
- les métriques passent souvent par :
    - certificats TLS
    - auth mTLS
    - endpoints dédiés
    - configuration kube-prometheus adaptée

---

## ✅ Bilan

Prometheus Operator scrape désormais correctement ETCD sous Minikube.

Les dashboards ETCD sont maintenant pleinement fonctionnels :
- histogrammes
- percentiles
- latences disque
- activité base clé/valeur
- commit durations
- métriques backend BoltDB

Cette étape permet désormais :
- l'observation réelle du comportement ETCD
- l'analyse des stress tests
- l'étude des performances du control-plane Kubernetes