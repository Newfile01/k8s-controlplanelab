# Prise en charge du monitoring par Prometheus Operator

## 🎯 Présentation

L'objectif de cette étape est de rendre l'opérateur observable par Prometheus Operator afin de superviser :
- son état,
- ses métriques internes,
- ses performances,
- ses réconciliations.

Cette étape introduit :
- Service Kubernetes,
- ServiceMonitor Prometheus Operator,
- exposition endpoint `/metrics`.

Ressources concernées :
- Deployment opérateur,
- Service,
- ServiceMonitor,
- Prometheus Operator.

---

## 🏗️ Architecture & emplacement de travail

### Répertoire principal

```text
operator/config/default/
operator/config/prometheus/
operator/config/rbac/
```

### Ressources générées

```text
Service
ServiceMonitor
ClusterRole metrics
ClusterRoleBinding metrics
```

### Chaîne de supervision

```text
Operator
→ Service
→ ServiceMonitor
→ Prometheus Operator
→ Prometheus
→ Grafana
```

---

## ⚙️ Actions effectuées

### Exposition du port métriques

Ajout des arguments dans le manager :

```yaml
--metrics-bind-address=:8443
```

Fichier :

```text
operator/config/default/manager_metrics_patch.yaml
```

### Création du Service métriques

Fichier :

```text
operator/config/default/metrics_service.yaml
```

Service créé :

```yaml
kind: Service
port: 8443
targetPort: 8443
```

### Création du ServiceMonitor

Fichier :

```text
operator/config/prometheus/monitor.yaml
```

Ajout :
- endpoint `/metrics`,
- port `https`,
- `tlsConfig`,
- `bearerTokenFile`.

### RBAC métriques

Fichiers :

```text
operator/config/rbac/metrics_reader_role.yaml
operator/config/rbac/metrics_auth_role.yaml
operator/config/rbac/metrics_auth_role_binding.yaml
```

---

## 🔎 Vérifications

### Vérifier le Service

```bash
kubectl get svc -n operator-system
```

### Vérifier le ServiceMonitor

```bash
kubectl get servicemonitor -n operator-system
```

### Vérifier la target Prometheus

Dans Grafana/Prometheus :

```text
Status > Targets
```

### Vérifier les métriques

```bash
kubectl port-forward svc/operator-controller-manager-metrics-service 8443:8443 -n operator-system
```

Puis :

```text
https://localhost:8443/metrics
```

---

## ✅ Bilan

L'opérateur expose désormais ses métriques Prometheus :
- temps de réconciliation,
- nombre de réconciliations,
- erreurs,
- métriques runtime Go,
- métriques controller-runtime.

Le monitoring est automatiquement intégré à Prometheus Operator via ServiceMonitor.