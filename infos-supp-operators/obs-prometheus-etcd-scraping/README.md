# Correctif Minikube pour scraping ETCD via Prometheus Operator

## 🎯 Présentation

Cette étape corrige le scraping ETCD sous Minikube.

Problème rencontré :
- ETCD écoute uniquement sur `127.0.0.1:2381`,
- Prometheus tente de scraper l'IP node,
- connexion refusée.

But :
- rendre les métriques ETCD accessibles à Prometheus Operator.

---

## 🏗️ Architecture & emplacement de travail

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

---

## ⚙️ Actions effectuées

### Diagnostic endpoint ETCD

```bash
kubectl describe pod etcd-control-plane-lab -n kube-system
```

Constat :

```text
--listen-metrics-urls=http://127.0.0.1:2381
```

### Vérification écoute locale

```bash
minikube ssh -p control-plane-lab
sudo ss -lntp | grep 2381
```

### Vérification endpoint métriques

```bash
curl http://127.0.0.1:2381/metrics
```

### Modification manifest ETCD

Fichier :

```text
/etc/kubernetes/manifests/etcd.yaml
```

Modification :

```yaml
--listen-metrics-urls=http://0.0.0.0:2381
```

### Redémarrage automatique ETCD

Le kubelet redémarre automatiquement le static pod.

---

## 🔎 Vérifications

### Vérifier écoute ETCD

```bash
ss -lntp | grep 2381
```

### Vérifier target Prometheus

Dans :

```text
Prometheus > Targets
```

Target attendue :

```text
kube-etcd UP
```

---

## ✅ Bilan

Prometheus Operator scrape désormais correctement ETCD sous Minikube.

Les dashboards ETCD sont maintenant pleinement fonctionnels :
- histogrammes,
- percentiles,
- latences disque,
- activité base clé/valeur.