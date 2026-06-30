# Récupération d'un cluster Minikube bloqué par un grand nombre de ConfigMaps

## Contexte

Suite à un test de charge, plus de **1000 ConfigMaps** étaient encore enregistrées dans le snapshot **etcd**. Lors du redémarrage du cluster Minikube, l'API Server tentait de recharger l'ensemble de ces objets, provoquant une saturation mémoire du control-plane (`OOMKilled`) et empêchant le cluster de démarrer correctement.

L'objectif est donc de supprimer directement les ConfigMaps de la base **etcd** sans recréer le cluster.

```bash
###############################################################################
# 1. Démarrer le conteneur du control-plane
###############################################################################

docker start control-plane-lab

###############################################################################
# 2. Ouvrir un shell dans le conteneur
###############################################################################

docker exec -it control-plane-lab bash

###############################################################################
# 3. Vérifier la présence des binaires etcd / etcdctl
###############################################################################

find / -name etcd 2>/dev/null
find / -name etcdctl 2>/dev/null

###############################################################################
# 4. Vérifier le manifeste du pod statique etcd
###############################################################################

cat /etc/kubernetes/manifests/etcd.yaml

###############################################################################
# 5. Démarrer manuellement etcd à partir de la base existante
###############################################################################

/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/31/fs/usr/local/bin/etcd \
  --name=control-plane-lab \
  --data-dir=/var/lib/minikube/etcd \
  --listen-client-urls=https://127.0.0.1:2379 \
  --advertise-client-urls=https://127.0.0.1:2379 \
  --listen-peer-urls=https://192.168.49.2:2380 \
  --initial-advertise-peer-urls=https://192.168.49.2:2380 \
  --initial-cluster=control-plane-lab=https://192.168.49.2:2380 \
  --cert-file=/var/lib/minikube/certs/etcd/server.crt \
  --key-file=/var/lib/minikube/certs/etcd/server.key \
  --trusted-ca-file=/var/lib/minikube/certs/etcd/ca.crt \
  --client-cert-auth=true \
  --peer-cert-file=/var/lib/minikube/certs/etcd/peer.crt \
  --peer-key-file=/var/lib/minikube/certs/etcd/peer.key \
  --peer-trusted-ca-file=/var/lib/minikube/certs/etcd/ca.crt \
  --peer-client-cert-auth=true

###############################################################################
# 6. Dans un second terminal, ouvrir un nouveau shell dans le conteneur
###############################################################################

docker exec -it control-plane-lab bash

###############################################################################
# 7. Vérifier qu'etcd écoute bien sur le port 2379
###############################################################################

ss -lnt | grep 2379

###############################################################################
# 8. Lister les ConfigMaps stockées dans etcd
###############################################################################

/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/32/fs/usr/local/bin/etcdctl \
  --endpoints=https://127.0.0.1:2379 \
  --cacert=/var/lib/minikube/certs/etcd/ca.crt \
  --cert=/var/lib/minikube/certs/etcd/server.crt \
  --key=/var/lib/minikube/certs/etcd/server.key \
  get /registry/configmaps/operator-system \
  --prefix --keys-only

###############################################################################
# 9. Supprimer toutes les ConfigMaps du namespace operator-system
###############################################################################

/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/32/fs/usr/local/bin/etcdctl \
  --endpoints=https://127.0.0.1:2379 \
  --cacert=/var/lib/minikube/certs/etcd/ca.crt \
  --cert=/var/lib/minikube/certs/etcd/server.crt \
  --key=/var/lib/minikube/certs/etcd/server.key \
  del /registry/configmaps/operator-system \
  --prefix

###############################################################################
# 10. Vérifier que toutes les ConfigMaps ont bien été supprimées
###############################################################################

/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/32/fs/usr/local/bin/etcdctl \
  --endpoints=https://127.0.0.1:2379 \
  --cacert=/var/lib/minikube/certs/etcd/ca.crt \
  --cert=/var/lib/minikube/certs/etcd/server.crt \
  --key=/var/lib/minikube/certs/etcd/server.key \
  get /registry/configmaps/operator-system \
  --prefix --keys-only

###############################################################################
# 11. Arrêter le serveur etcd lancé manuellement
###############################################################################

# Ctrl+C dans le terminal où etcd est exécuté.

###############################################################################
# 12. Redémarrer normalement le cluster Minikube
###############################################################################

minikube start -p control-plane-lab
```
