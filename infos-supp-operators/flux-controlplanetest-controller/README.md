| Étape | Opérateur (ControlPlaneTest Controller) | Kubernetes natif |
|---------|------------------------------------------|------------------|
| 1 | Reçoit événement CREATE CR | API Server stocke CR dans ETCD |
| 2 | GET ControlPlaneTest | |
| 3 | Détecte absence du Finalizer | |
| 4 | UPDATE CR (ajout Finalizer) | API Server → ETCD |
| 5 | Requeue immédiat | |
| 6 | GET ControlPlaneTest | |
| 7 | GET Deployment t1-deployment-0 | API Server retourne NotFound |
| 8 | CREATE Deployment | API Server → ETCD |
| 9 | Requeue immédiat | |
| 10 | | Deployment Controller détecte nouveau Deployment |
| 11 | | CREATE ReplicaSet |
| 12 | | API Server → ETCD |
| 13 | | ReplicaSet Controller détecte ReplicaSet |
| 14 | | CREATE Pod |
| 15 | | API Server → ETCD |
| 16 | | Scheduler détecte Pod Pending |
| 17 | | Choisit un nœud |
| 18 | | UPDATE Pod.spec.nodeName |
| 19 | | Kubelet démarre le conteneur |
| 20 | Event Pod Create reçu via Owns(Pod) | |
| 21 | Reconcile déclenchée | |
| 22 | GET Deployment | |
| 23 | GET Service | NotFound |
| 24 | CREATE Service | API Server → ETCD |
| 25 | Requeue immédiat | |
| 26 | Event Service Create reçu via Owns(Service) | |
| 27 | Reconcile déclenchée | |
| 28 | GET Deployment | |
| 29 | Lecture Deployment.Status.ReadyReplicas | |
| 30 | Construction Conditions | |
| 31 | UPDATE Status CR | API Server → ETCD |
| 32 | Requeue immédiat | |
| 33 | Reconcile déclenchée | |
| 34 | GET CR | |
| 35 | GET Deployment | |
| 36 | GET Service | |
| 37 | Vérification drift | Aucun drift |
| 38 | Réconciliation terminée | État convergé |

## Ressources créées au final

ControlPlaneTest
│
├── Deployment
│   │
│   └── ReplicaSet
│        │
│        └── Pod
│
└── Service

## Nombre approximatif d'opérations

CR Create                    : 1
CR Update (Finalizer)        : 1
Deployment Create            : 1
ReplicaSet Create            : 1
Pod Create                   : 1
Service Create               : 1
Status Update                : 1

Réconciliations opérateur    : ~5

API Server Writes            : ~7
ETCD Writes                  : ~7

Contrôleurs impliqués :

- ControlPlaneTest Controller (ton opérateur)
- Deployment Controller
- ReplicaSet Controller
- Scheduler
- Kubelet
- EndpointSlice Controller (création du Service)
- Garbage Collector (OwnerReferences)