# Réglages cluster & prise en compte des limites Minikube + WSL2

## 🎯 Présentation

Cette étape vise à :
- ajuster les ressources allouées à Minikube
- comprendre le comportement du driver Docker sous WSL2
- adapter le cluster à la machine physique
- éviter la surallocation CPU/RAM

Le but est d'obtenir :
- un cluster cohérent avec les ressources physiques réellement disponibles
- un environnement stable pour les stress tests
- un comportement observable du control-plane sous charge

---

## 🏗️ Architecture & emplacement de travail

### Environnement utilisé

```text
Windows
→ WSL2 Ubuntu
→ Docker CLI
→ Minikube Driver Docker
→ Kubernetes
```

Important :
- avec le driver Docker, Minikube ne crée PAS de VM dédiée
- chaque noeud Kubernetes devient un conteneur Docker
- les ressources sont donc consommées directement sur l'hôte WSL2

### Cluster

```text
4 Nodes :
- 1 Control Plane
- 3 Workers
```

### ⚠️ IMPORTANT

Le fait de fonctionner avec le driver Docker ne limite PAS les quotas de ressources "PAR noeud" mais implique de raisonner globalement

```text
(FAUX)
8 CPUs demandés sur 4 noeuds = 8 CPUs / noeud

(VRAI)
8 CPUs demandés sur 4 noeuds
→ chaque noeud voit potentiellement 8 CPUs
→ soit jusqu'à 32 CPUs théoriquement visibles dans Kubernetes pour chaque noeud
```

Explication :
- les conteneurs Minikube partagent le noyau Linux WSL2
- Kubernetes voit les ressources exposées par Docker/WSL2
- il ne s'agit pas d'une vraie isolation hyperviseur comme avec VirtualBox ou KVM

Conséquence :
- Kubernetes peut annoncer davantage de ressources que réellement disponibles physiquement
- le scheduler raisonne sur les ressources visibles
- mais l'hôte physique reste la vraie limite

### Limites WSL2

WSL2 possède également ses propres limites système :
- nombre de watchers inotify
- nombre d'instances fsnotify
- nombre maximal de fichiers/socket ouverts (`ulimit`)

Exemples :
- `fs.inotify.max_user_watches`
- `fs.inotify.max_user_instances`
- `ulimit -n`

Ces limites deviennent rapidement problématiques avec :
- plusieurs noeuds Kubernetes
- Prometheus
- Grafana
- controller-runtime
- nombreux watchers Kubernetes

Cela peut provoquer :
- échec démarrage noeuds
- kubelet instable
- Prometheus instable
- erreurs containerd
- pods bloqués

Exemple typique :

```text
Noeud 1, 2 et 3 démarrent correctement
Noeud 4 impossible à lancer
→ limites système atteintes
```

---

## ⚙️ Actions effectuées

### Nettoyage complet du cluster

⚠️ Attention :
Ne pas effectuer de `docker system prune` pendant que le cluster fonctionne

Cela peut supprimer :
- réseaux Docker utilisés par Kubernetes
- volumes containerd
- images critiques
- couches nécessaires aux noeuds Minikube

Conséquences possibles :
- corruption du cluster
- kubelet cassé
- pods impossibles à redémarrer
- networking cassé

=> Impossible d'arrêter ou supprimer son cluster
=> Il faudra recréer le cluster et le supprimer d'abord via Minikube avant de le faire à travers docker

```bash
# CLEAN MINIKUBE ET DOCKER
minikube stop
minikube delete --profile=control-plane-lab
docker system prune -af
```

### Préparation WSL2 & Création du cluster

```bash
#### ERREURS DE DEMARRAGE

# ATTENTION AUX LIMITES inotify & fsnotify
cat /proc/sys/fs/inotify/max_user_watches
cat /proc/sys/fs/inotify/max_user_instances
ulimit -n

max_user_watches   = 524288
max_user_instances = 128
ulimit -n          = 10240
```

Analyse :
- `max_user_watches`
    → nombre maximal de fichiers observables simultanément

- `max_user_instances`
    → nombre maximal d'instances fsnotify/inotify ouvertes

- `ulimit -n`
    → nombre maximal de descripteurs ouverts
    → sockets
    → fichiers
    → connexions réseau

Dans Kubernetes ces limites sont fortement sollicitées :
- kubelet observe énormément de fichiers
- controller-runtime ouvre des watchers
- Prometheus maintient beaucoup de connexions
- containerd utilise de nombreux file descriptors


### Correctifs

```bash
#  immédiats
sudo sysctl fs.inotify.max_user_instances=8192
ulimit -n 65535

# permanents
sudo vi /etc/sysctl.conf

# Ajouter
fs.inotify.max_user_watches=524288
fs.inotify.max_user_instances=8192

# Puis
sudo sysctl -p

# Et enfin redémarrage complet WSL2 dans powershell
wsl --shutdown
```

Important :
- WSL2 conserve certains états mémoire/kernel
- un simple redémarrage shell ne suffit pas toujours


### Configuration cluster correcte

```bash
minikube start \
--profile=control-plane-lab \
--driver=docker \
--container-runtime=containerd \
--kubernetes-version=v1.35.1 \
--nodes=4 \
--memory=4096 \
--cpus=2 \
--bootstrapper=kubeadm \
--extra-config=kubelet.authentication-token-webhook=true \
--extra-config=kubelet.authorization-mode=Webhook \
--extra-config=scheduler.bind-address=0.0.0.0 \
--extra-config=controller-manager.bind-address=0.0.0.0

# Activation metrics-server
minikube addons enable metrics-server

# Configurer ce cluster comme contexte par défaut 
# (pratique pour toute commande minikube ultiéreur ainsi que pour kubebuilder)
minikube profile control-plane-lab
```

Explication paramètres importants :

- `--nodes=4`
    → cluster multi-noeuds

- `--memory=4096`
    → mémoire visible par noeud

- `--cpus=2`
    → vCPU visibles par noeud

- `scheduler.bind-address=0.0.0.0`
    → expose métriques scheduler

- `controller-manager.bind-address=0.0.0.0`
    → expose métriques controller-manager

- `addons enable metrics-server`
    Permet :
    - `kubectl top`
    - métriques CPU/RAM live
    - HPA
    - supervision consommation cluster

### Vérification ressources visibles

```bash
kubectl get nodes \
-o custom-columns=NAME:.metadata.name,CPU:.status.capacity.cpu,MEM:.status.capacity.memory
```

Permet de visualiser :
- ressources annoncées à Kubernetes
- ressources visibles par noeud
- cohérence avec la configuration Minikube

---

## ⚙️ Modification des ressources (si cluster déjà en fonction)

Configuration retenue :

```text
2 vCPU / node
4GB RAM / node
```

But :
- éviter la saturation machine hôte
- conserver suffisamment de pression control-plane
- rendre les dashboards intéressants visuellement

### Modification configuration Minikube

```bash
minikube stop -p control-plane-lab

# PAR NOEUD
minikube config set cpus 2
minikube config set memory 4096
```

Important :
- ces valeurs s'appliquent à chaque noeud
- Kubernetes démultiplie ensuite les ressources globales visibles

Donc :

```text
4 nodes
× 2 CPUs
= 8 vCPU visibles cluster

4 nodes
× 4GB RAM
= 16GB RAM visibles cluster
```

---

## 🔎 Vérifications

```bash
# statut cluster
minikube status -p control-plane-lab
# consommation du cluster
kubectl top nodes
# config des noeuds
kubectl describe node
```

---

## ✅ Bilan

Le cluster Minikube est désormais :
- cohérent avec la machine physique
- calibré pour les stress tests
- adapté à WSL2 + Docker Driver
- suffisamment contraint pour rendre les métriques control-plane visibles

Les limites système Linux/WSL2 sont maintenant correctement prises en compte :
- inotify
- file descriptors
- watchers Kubernetes
- ressources Docker/WSL2

Le cluster devient ainsi :
- plus stable
- plus réaliste
- plus exploitable pour les scénarios de supervision et de stress tests