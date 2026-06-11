# 📊 Metrics Kubernetes Operator — HTTP / HTTPS / AuthN & AuthZ

# 🎯 Objectif

Controller-runtime (Kubebuilder) expose automatiquement un endpoint :

```text
/metrics
```

Ce endpoint permet :
- Prometheus scraping
- monitoring opérateur
- debugging
- observabilité
- mesure performances reconcile/workqueue

---

# 🧠 Fonctionnement général des flags Go

Dans `main.go`, Kubebuilder utilise le package standard Go :

```go
import "flag"
```

Les flags permettent :
- de modifier le comportement du binaire au lancement
- sans recompiler l’opérateur

## 📦 Exemple décomposé

```go
flag.StringVar(
	&metricsAddr,
	"metrics-bind-address",
	":8443",
	"Adresse d'écoute du serveur metrics",
)
```

| Élément | Rôle |
|---|---|
| `flag` | package standard Go de gestion CLI |
| `StringVar` | création d’un flag string |
| `&metricsAddr` | variable Go alimentée par le flag |
| `"metrics-bind-address"` | nom du flag CLI |
| `":8443"` | valeur par défaut |
| `"..."` | description affichée avec `--help` |

---

# 🔄 Fonctionnement runtime

Au lancement :

```bash
./operator --metrics-bind-address=:8443
```

Go :
1. parse les flags (`flag.Parse()`)
2. remplit `metricsAddr`
3. controller-runtime démarre le serveur metrics
4. le serveur écoute sur l’adresse définie

---

# 🧩 Architecture générale metrics

```text
curl / Prometheus
        ↓
metrics server controller-runtime
        ↓
/metrics
        ↓
Prometheus metrics Go + Kubernetes + controller-runtime
```

---

# ============================================================
# 1️⃣ HTTP SIMPLE (AUCUNE SECURITE)
# ============================================================

## 🎯 Usage

- DEV local
- LAB
- tests rapides

⚠️ NON recommandé en production.


## 📍 Configuration main.go

* Activer HTTP :
Remplacer `flag.StringVar(&metricsAddr, "metrics-bind-address", "0",...)` par `flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080",...)`

* Désactiver HTTPS :
`flag.BoolVar(&secureMetrics, "metrics-secure", false,...)`

* Supprimer AuthN/AuthZ :
➡️ commenter/supprimer ce bloc :
```go
metricsServerOptions.FilterProvider =
	filters.WithAuthenticationAndAuthorization
```

## 🔄 Fonctionnement schématique

```text
curl HTTP
    ↓
controller-runtime metrics server
    ↓
/metrics
    ↓
réponse immédiate
```

## 🔍 Vérification

## CURL

```bash
curl http://localhost:8080/metrics
```

Doit afficher :

```text
controller_runtime_reconcile_total
```


## OpenSSL

Ne fonctionne PAS :

```bash
openssl s_client -connect localhost:8080
```

➡️ erreur normale car :
- pas TLS
- pas HTTPS

---

# ============================================================
# 2️⃣ HTTPS SIMPLE (TLS UNIQUEMENT)
# ============================================================

## 🎯 Usage

- DEV avancé
- LAB sécurisé
- chiffrement réseau local

➡️ HTTPS sans contrôle utilisateur.

## 📍 Configuration main.go

* Activer HTTPS :
Remplacer `flag.StringVar(&metricsAddr, "metrics-bind-address", "0",...)` par `flag.StringVar(&metricsAddr, "metrics-bind-address", ":8443",...)`

* Activer TLS :
`flag.BoolVar(&secureMetrics, "metrics-secure", true,...)`

* Désactiver AuthN/AuthZ :
➡️ commenter/supprimer :
```go
metricsServerOptions.FilterProvider =
	filters.WithAuthenticationAndAuthorization
```

## 🔐 Certificats

Controller-runtime génère automatiquement :
- certificat auto-signé
- CA locale
- clé privée

➡️ générés automatiquement au runtime.

## 🔄 Fonctionnement schématique

```text
curl HTTPS
    ↓
TLS Handshake
    ↓
controller-runtime metrics server
    ↓
/metrics
    ↓
réponse chiffrée HTTPS
```

## 🔍 Vérification

### CURL

```bash
curl -k https://localhost:8443/metrics
```

➡️ `-k` ignore le certificat auto-signé.

Doit afficher :

```text
controller_runtime_reconcile_total
```

### OpenSSL

```bash
openssl s_client -connect localhost:8443
```

## ✅ Informations importantes à observer

| Information | Résultat attendu |
|---|---|
| Version TLS | `TLSv1.3` |
| Cipher | `TLS_AES_128_GCM_SHA256` |
| Certificat | `CN = localhost-ca` |
| HTTP2 | `No ALPN negotiated` |

---

# ============================================================
# 3️⃣ HTTPS + AUTHN + AUTHZ
# ============================================================

## 🎯 Usage

- production
- Prometheus Kubernetes sécurisé
- RBAC metrics
- multi-tenant clusters

## 📍 Configuration main.go

* Activer HTTPS :
Remplacer `flag.StringVar(&metricsAddr, "metrics-bind-address", "0",...)` par `flag.StringVar(&metricsAddr, "metrics-bind-address", ":8443",...)`

* Activer TLS :
`flag.BoolVar(&secureMetrics, "metrics-secure", true,...)`

* Activer AuthN/AuthZ :
```go
metricsServerOptions.FilterProvider =
	filters.WithAuthenticationAndAuthorization
```

## 🔐 Fonctionnement sécurité

Controller-runtime délègue :
- authentification
- autorisation

au cluster Kubernetes via RBAC.

## 🔄 Fonctionnement schématique

```text
curl / Prometheus
        ↓
TLS Handshake
        ↓
Authentication Kubernetes
        ↓
Authorization RBAC
        ↓
/metrics autorisé/refusé
```

## 🔍 Vérification

### CURL

```bash
curl -k https://localhost:8443/metrics
```

Résultat attendu :

```text
Unauthorized
```

➡️ indique :
- TLS fonctionnel
- endpoint protégé
- AuthN/AuthZ actif

### OpenSSL

```bash
openssl s_client -connect localhost:8443
```

## ✅ Informations importantes à observer

| Information | Résultat attendu |
|---|---|
| Version TLS | `TLSv1.3` |
| Cipher | `TLS_AES_128_GCM_SHA256` |
| Certificat | `CN = localhost-ca` |
| HTTP2 | `No ALPN negotiated` |
| Réponse HTTP | `Unauthorized` |

---

# 🔐 Certificats personnalisés

## 📍 Variables utilisées

```go
metricsCertPath
metricsCertName
metricsCertKey
```

## 📦 Exemple configuration

```go
metricsServerOptions.CertDir = "/certs"
metricsServerOptions.CertName = "tls.crt"
metricsServerOptions.KeyName = "tls.key"
```

## 📁 Structure attendue

```text
/certs/
 ├── tls.crt
 └── tls.key
```

---

# 🔥 Vérification rapide de la situation actuelle

| Situation | curl | openssl |
|---|---|---|
| HTTP simple | `curl http://...` | échoue |
| HTTPS simple | `curl -k https://...` | TLS OK |
| HTTPS + Auth | `Unauthorized` | TLS OK |