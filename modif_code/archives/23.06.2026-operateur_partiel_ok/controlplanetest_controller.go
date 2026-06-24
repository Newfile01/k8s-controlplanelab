/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"golang.org/x/time/rate"

	// Base pour un controller d'Operateur
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	builder "sigs.k8s.io/controller-runtime/pkg/builder"

	// Event-driven controller
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// Metrics

	// Mon API
	controlplanev1alpha1 "github.com/Newfile01/k8s-controlplanelab/operator/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

// Ajout des droits RBAC pour manipuler les Deployments, Pods, etc... avec les "Owns"
// +kubebuilder:rbac:groups=controlplane.lab.local,resources=controlplanetests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=controlplane.lab.local,resources=controlplanetests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=controlplane.lab.local,resources=controlplanetests/finalizers,verbs=update

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get

// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// ControlPlaneTestReconciler reconciles a ControlPlaneTest object
type ControlPlaneTestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Constante pour les Finalizers (permette de vérifier l'état d'un CR avant suppression, notamment pour la gestion de ressources externes)
// Exemples :
// const (
// 	finalizerName = "my.domain/finalizer"
// 	defaultImage = "nginx:latest"
// )

const (
	finalizerName      = "controlplanetest.lab.local/finalizer"
	deploymentOwnerKey = ".metadata.controller"
)

// Mes metriques personnalisées
var (
	reconciliationTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "controlplanetest_reconciliation_total",
			Help: " = Nombre total de reconciliations",
		},
	)

	erreursReconciliationTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "controlplanetest_erreurs_reconciliation_total",
			Help: " = Nombre total d erreurs de reconciliation",
		},
	)

	podsGeresGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "controlplanetest_pods_geres",
			Help: " = Nombre de Pods actuellement geres",
		},
	)

	replicasDisponiblesGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "controlplanetest_replicas_disponibles",
			Help: " = Nombre de replicas disponibles",
		},
	)

	reconciliationDuree = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "controlplanetest_duree_reconciliation_secondes",
			Help:    " = Temps d execution des reconciliations par type (creation, suppression, erreurs, etc.)",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type_reconciliation"},
	)

	// ============================================================
	// SCHEDULER STRESS METRICS
	// ============================================================

	deploymentsGeneresGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "controlplanetest_deployments_generes",
			Help: "Nombre de Deployments actuellement generes",
		},
		[]string{
			"cr",
			"namespace",
			"profile",
		},
	)

	podsDesiresGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "controlplanetest_pods_desires",
			Help: "Nombre total de Pods desires",
		},
		[]string{
			"cr",
			"namespace",
			"profile",
		},
	)

	podsPendingGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "controlplanetest_pods_pending",
			Help: "Nombre de Pods actuellement Pending",
		},
		[]string{
			"cr",
			"namespace",
			"profile",
		},
	)

	// ============================================================
	// API SERVER STRESS METRICS
	// ============================================================

	statusUpdatesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "controlplanetest_status_updates_total",
			Help: "Nombre total de mises a jour Status",
		},
		[]string{
			"cr",
			"namespace",
			"profile",
		},
	)

	requeuesForceesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "controlplanetest_requeues_forcees_total",
			Help: "Nombre total de requeues forcees",
		},
		[]string{
			"cr",
			"namespace",
			"profile",
		},
	)

	resourcesRecreatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "controlplanetest_resources_recreated_total",
			Help: "Nombre total de ressources recreees",
		},
		[]string{
			"cr",
			"namespace",
			"profile",
		},
	)

	// ============================================================
	// POD LIFECYCLE STORM METRICS
	// ============================================================

	podsSupprimesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "controlplanetest_pods_supprimes_total",
			Help: "Nombre total de Pods supprimes",
		},
		[]string{
			"cr",
			"namespace",
			"profile",
		},
	)

	// ============================================================
	// CONFIGURATION SNAPSHOT METRICS
	// ============================================================

	configurationScenarioInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "controlplanetest_configuration_info",
			Help: "Snapshot de configuration du scenario de stress",
		},
		[]string{
			"cr",
			"namespace",
			"profile",
			"scheduler_enabled",
			"topologyspread",
			"affinity",
			"antiaffinity",
			"aggressive_reconcile",
			"frequent_status_updates",
			"recreate_resources",
			"podstorm_enabled",
			"delete_pods_randomly",
		},
	)
)

func init() {
	metrics.Registry.MustRegister(
		reconciliationTotal,
		erreursReconciliationTotal,
		podsGeresGauge,
		replicasDisponiblesGauge,
		reconciliationDuree,

		deploymentsGeneresGauge,
		podsDesiresGauge,
		podsPendingGauge,

		statusUpdatesTotal,
		requeuesForceesTotal,
		resourcesRecreatedTotal,

		podsSupprimesTotal,

		configurationScenarioInfo,
	)
}

// +kubebuilder:rbac:groups=controlplane.lab.local,resources=controlplanetests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=controlplane.lab.local,resources=controlplanetests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=controlplane.lab.local,resources=controlplanetests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ControlPlaneTest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.3/pkg/reconcile
func (r *ControlPlaneTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	debutReconciliation := time.Now()
	typeReconciliation := "globale"
	var requeueResult *ctrl.Result

	// "Defer" Déclenche la fonction suivante à chaque "return" de la fonction dans laquelle elle se trouve
	// Ici on veut mesurer les durées pour chaque boucle de réconciliation : CREATION, UPDATE, STATUS, ERREUR, SUPPRESSION, etc...
	defer func() {
		reconciliationDuree.WithLabelValues(
			typeReconciliation,
		).Observe(
			time.Since(debutReconciliation).Seconds(),
		)
	}()

	_ = logf.FromContext(ctx)
	reconciliationTotal.Inc()

	fmt.Println("\n\n================ RECONCILIATION =================")

	// =========================================================
	// ******** RECUPERATION DE LA CUSTOM RESOURCE ********
	// =========================================================
	// Cette partie consiste à récupérer la ressource demandée par la requête envoyée au controller-runtime.
	// req.NamespacedName contient :
	// - le namespace
	// - le nom
	// de la ressource à récupérer.
	// L'opérateur effectue ici une requête GET vers l'API Server Kubernetes.
	// C'est Kubernetes qui répond avec l'état réel observé dans le cluster.

	// Structure Go vide destinée à accueillir la Custom Resource
	controlPlaneTest := &controlplanev1alpha1.ControlPlaneTest{}

	fmt.Println("🖥️🔍📦 │ Récupération des informations de la Custom Resource ...")

	// GET Kubernetes API : récupération de la CR correspondant à la requête
	err := r.Get(ctx, req.NamespacedName, controlPlaneTest)
	if err != nil {

		// IgnoreNotFound() :
		// - évite de considérer une suppression comme une erreur
		// - stoppe proprement la réconciliation
		// Cas possibles :
		// - suppression utilisateur
		// - namespace supprimé
		// - objet inexistant

		fmt.Println("🖥️🔍🚫 │ Aucune Custom Resource correspondante trouvée")
		typeReconciliation = "cr_absent"
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// =========================================================
	// OBSERVED GENERATION
	// =========================================================
	// Kubernetes incrémente automatiquement metadata.generation
	// lorsqu'un utilisateur modifie le Spec.
	// observedGeneration permet à l'opérateur de savoir
	// quelle version du Spec a déjà été traitée.

	fmt.Println("🤓🔄📦 │ Generation désirée :", controlPlaneTest.Generation)
	fmt.Println("☸️📊📦 │ Generation observée :", controlPlaneTest.Status.ObservedGeneration)

	specChanged := controlPlaneTest.Generation != controlPlaneTest.Status.ObservedGeneration

	if specChanged {
		fmt.Println("🤓🔄📦 │ Modification utilisateur détectée dans le Spec")
	}

	// =========================================================
	// ******** GESTION DU FINALIZER ********
	// =========================================================
	// Les Finalizers permettent à l'opérateur :
	// - d'intercepter une suppression
	// - d'effectuer un nettoyage custom
	// - puis seulement ensuite autoriser la suppression réelle.
	//
	// Le DeletionTimestamp est ajouté automatiquement par Kubernetes lorsque l'utilisateur effectue :
	// kubectl delete ...
	//
	// Tant que le finalizer est présent : Kubernetes refuse la suppression réelle de la ressource.

	// Ressource PAS en cours de suppression
	if controlPlaneTest.ObjectMeta.DeletionTimestamp.IsZero() {
		// Vérification présence finalizer
		if !controllerutil.ContainsFinalizer(controlPlaneTest, finalizerName) {

			fmt.Println("🖥️⬆️🔚 │ Ajout du Finalizer ...")

			// Modification locale de la structure Go
			controllerutil.AddFinalizer(controlPlaneTest, finalizerName)

			// UPDATE Kubernetes API : persistance réelle du finalizer dans le cluster
			fmt.Println("🖥️🔄📦 │ UPDATE Custom Resource avec Finalizer ...")

			err := r.Update(ctx, controlPlaneTest)
			if err != nil {

				fmt.Println("🖥️🔄❌ │ Impossible d'ajouter le Finalizer")

				erreursReconciliationTotal.Inc()
				return ctrl.Result{}, err
			}

			fmt.Println("🖥️🔄✅ │ Finalizer ajouté avec succès")

			// Nouvelle reconciliation forcée
			return ctrl.Result{Requeue: true}, nil
		}

	} else {
		// Ressource en cours de suppression
		fmt.Println("🤓🗑️📦 │ Suppression de la Custom Resource détectée ...")

		if controllerutil.ContainsFinalizer(controlPlaneTest, finalizerName) {
			fmt.Println("🖥️🔚📦 │ Exécution du nettoyage avant suppression ...")

			// Ici pourraient être exécutés :
			// - backup
			// - cleanup externe
			// - désenregistrement
			// - suppression cloud provider
			// etc...

			// Modification locale de la structure Go
			controllerutil.RemoveFinalizer(controlPlaneTest, finalizerName)

			// UPDATE Kubernetes API : suppression réelle du finalizer dans le cluster
			fmt.Println("🖥️🔄📦 │ UPDATE Custom Resource sans Finalizer ...")

			err := r.Update(ctx, controlPlaneTest)
			if err != nil {

				fmt.Println("🖥️🔄❌ │ Impossible de supprimer le Finalizer")

				erreursReconciliationTotal.Inc()
				return ctrl.Result{}, err
			}
		}

		// Une fois le dernier finalizer supprimé : Kubernetes supprimera réellement la ressource.
		fmt.Println("🖥️🗑️✅ │ Finalizer supprimé, Kubernetes peut terminer la suppression")
		typeReconciliation = "suppression_cr"
		return ctrl.Result{}, nil
	}

	// =========================================================
	// ******** GESTION DU DEPLOYMENT ********
	// =========================================================
	// L'objectif est de maintenir :
	// Etat désiré (CR) == Etat réel (Deployment)
	// Toute divergence détectée sera corrigée automatiquement par l'opérateur.

	// Noms attendus
	serviceName := controlPlaneTest.Name + "-service"

	// Liste des ConfigMaps gérées par cette CR
	var configMapNames []string

	// Liste des Deployments gérés par cette CR
	var deploymentNames []string

	// =========================================================
	// WORKLOAD DEFINITION
	// =========================================================
	//
	// Cette section définit la forme de la charge
	// générée par l'opérateur.
	//
	// DeploymentCount
	//     => nombre de Deployments
	//
	// ReplicasPerDeployment
	//     => nombre de Pods par Deployment
	//
	// Les scénarios de stress (Scheduler,
	// API Server, ETCD, Controller Manager,
	// Operator...) utilisent ensuite cette
	// charge comme base de travail.
	//==============================================

	// Compatibilité avec l'ancien mode simple.
	deploymentCount := int32(1)
	if controlPlaneTest.Spec.DeploymentCount > 0 {
		deploymentCount = controlPlaneTest.Spec.DeploymentCount
	}

	// Compatibilité avec l'ancien champ Replicas.
	replicasPerDeployment := controlPlaneTest.Spec.Replicas
	if controlPlaneTest.Spec.ReplicasPerDeployment > 0 {
		replicasPerDeployment = controlPlaneTest.Spec.ReplicasPerDeployment
	}

	// =====================================
	// ******** BOUCLE DEPLOYMENTS ********
	// =====================================

	for i := int32(0); i < deploymentCount; i++ {
		deploymentName := fmt.Sprintf(
			"%s-deployment-%d",
			controlPlaneTest.Name,
			i,
		)

		// Ajout du Deployment dans la liste des ressources gérées
		deploymentNames = append(
			deploymentNames,
			deploymentName,
		)

		fmt.Println("🖥️🔄📦 │ Gestion Deployment :", deploymentName)

		// **** DEPLOYMENT ****
		// Structure Go vide destinée à accueillir le Deployment récupéré depuis Kubernetes
		existingDeployment := &appsv1.Deployment{}

		fmt.Println("🖥️🔍📦 │ Recherche du Deployment dans Kubernetes ...")

		// GET Kubernetes API : récupération du Deployment réel existant
		err = r.Get(ctx, types.NamespacedName{
			Name:      deploymentName,
			Namespace: req.Namespace,
		}, existingDeployment)

		// =========================================================
		// CREATION DU DEPLOYMENT
		// =========================================================
		// Si Kubernetes répond "NotFound" :
		// cela signifie que le Deployment n'existe pas encore.
		//
		// L'opérateur doit donc :
		// - définir l'objet attendu
		// - puis demander sa création à Kubernetes.

		if err != nil && apierrors.IsNotFound(err) {

			fmt.Println("🖥️🔍🚫 │ Deployment introuvable")
			fmt.Println("🖥️⬆️📦 │ Création du Deployment ...")

			// =========================================================
			// AFFINITY / ANTI-AFFINITY
			// =========================================================

			var affinity *corev1.Affinity

			// =====================================
			// POD AFFINITY
			// =====================================

			if controlPlaneTest.Spec.SchedulerStress.Affinity != "" {

				fmt.Println("🖥️🧲📦 │ Activation Pod Affinity")

				podAffinityTerm := corev1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": controlPlaneTest.Name,
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				}

				// =====================================
				// SOFT AFFINITY
				// =====================================

				if controlPlaneTest.Spec.SchedulerStress.Affinity == "soft" {

					affinity = &corev1.Affinity{
						PodAffinity: &corev1.PodAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
								{
									Weight:          100,
									PodAffinityTerm: podAffinityTerm,
								},
							},
						},
					}
				}

				// =====================================
				// HARD AFFINITY
				// =====================================

				if controlPlaneTest.Spec.SchedulerStress.Affinity == "hard" {

					affinity = &corev1.Affinity{
						PodAffinity: &corev1.PodAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								podAffinityTerm,
							},
						},
					}
				}
			}

			// =====================================
			// POD ANTI-AFFINITY
			// =====================================

			if controlPlaneTest.Spec.SchedulerStress.AntiAffinity != "" {

				fmt.Println("🖥️🚫📦 │ Activation Pod AntiAffinity")

				podAntiAffinityTerm := corev1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": controlPlaneTest.Name,
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				}

				// =====================================
				// SOFT ANTI-AFFINITY
				// =====================================

				if controlPlaneTest.Spec.SchedulerStress.AntiAffinity == "soft" {

					affinity = &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
								{
									Weight:          100,
									PodAffinityTerm: podAntiAffinityTerm,
								},
							},
						},
					}
				}

				// =====================================
				// HARD ANTI-AFFINITY
				// =====================================

				if controlPlaneTest.Spec.SchedulerStress.AntiAffinity == "hard" {

					affinity = &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								podAntiAffinityTerm,
							},
						},
					}
				}
			}

			// =====================================================
			// DEFINITION LOCALE DE L'OBJET DEPLOYMENT
			// =====================================================
			// Ici : l'opérateur construit simplement une structure Go.
			// Aucune ressource Kubernetes réelle n'existe encore à ce stade.

			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName,
					Namespace: req.Namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replicasPerDeployment,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app":        controlPlaneTest.Name,
							"deployment": deploymentName,
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app":        controlPlaneTest.Name,
								"deployment": deploymentName,
							},
						},
						Spec: corev1.PodSpec{
							// appliqué un sélecteur de noeud en fonction du label "clé: valeur" associé à nodeSelector
							NodeSelector: controlPlaneTest.Spec.SchedulerStress.NodeSelector,
							// Définition du mode de fonctionnement du deployment affinité/anti-affinité
							Affinity: affinity,
							TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
								{
									// écart maximum autorisé
									MaxSkew: 1,
									// répartition par noeud (possible ici d'étendre par région, pool, rack, etc...)
									TopologyKey: "kubernetes.io/hostname",
									// préfère équilibrer mais autorise quand même le scheduling ('corev1.DoNotSchedule' pour refuser complètement)
									WhenUnsatisfiable: corev1.ScheduleAnyway,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": controlPlaneTest.Name,
										},
									},
								},
							},
							Containers: []corev1.Container{
								{
									Name:  "nginx",
									Image: controlPlaneTest.Spec.Image,
								},
							},
						},
					},
				},
			}

			// =====================================================
			// RELATION PARENT / ENFANT
			// =====================================================
			// La CR devient propriétaire du Deployment.
			//
			// Kubernetes gérera alors automatiquement :
			// - suppression cascade
			// - garbage collection
			// - relation OwnerReference

			err = controllerutil.SetControllerReference(controlPlaneTest, deployment, r.Scheme)
			if err != nil {

				fmt.Println("🖥️🔄❌ │ Impossible d'ajouter l'OwnerReference")
				erreursReconciliationTotal.Inc()
				return ctrl.Result{}, err
			}

			// =====================================================
			// CREATION REELLE DANS KUBERNETES
			// =====================================================
			// Cette fois : l'opérateur effectue réellement une requête POST vers l'API Server Kubernetes.
			//
			// Kubernetes :
			// - valide l'objet
			// - stocke l'objet
			// - crée ReplicaSet
			// - crée Pods
			// - déclenche scheduler
			// etc...

			fmt.Println("🖥️⬆️📦 │ POST Deployment vers l'API Kubernetes ...")

			err = r.Create(ctx, deployment)
			if err != nil {

				fmt.Println("🖥️⬆️❌ │ Impossible de créer le Deployment")
				typeReconciliation = "err_creation_deployment"
				// En cas d'erreur : controller-runtime replanifiera automatiquement une nouvelle réconciliation.
				erreursReconciliationTotal.Inc()
				return ctrl.Result{}, err
			}

			fmt.Println("🖥️⬆️✅ │ Deployment créé avec succès")
			typeReconciliation = "creation_deployment"
			// Requeue : relance immédiate pour relire l'état réel créé par Kubernetes.
			return ctrl.Result{Requeue: true}, nil
		}

		if err != nil {

			fmt.Println("🖥️🔍❌ │ Erreur lors de la récupération du Deployment")
			typeReconciliation = "err_api_deployment"
			erreursReconciliationTotal.Inc()
			return ctrl.Result{}, err
		}

		// =========================================================
		// VERIFICATION DE LA STRUCTURE DU DEPLOYMENT
		// =========================================================
		// Protection contre l'accès à un tableau vide.
		// Sans cela : panic possible sur containers[0]

		containers := existingDeployment.Spec.Template.Spec.Containers

		if len(containers) < 1 {
			fmt.Println("☸️📦❌ │ Aucun container trouvé dans le Deployment")
			typeReconciliation = "err_conteneur_dans_deployment"
			return ctrl.Result{}, fmt.Errorf(
				"deployment %s ne contient aucun container",
				existingDeployment.Name,
			)
		}

		// =========================================================
		// DETECTION DE DRIFT
		// =========================================================
		// Comparaison : Etat désiré (CR) VS Etat réel (Deployment)
		// Toute différence déclenche une correction automatique.

		currentImage := containers[0].Image
		desiredImage := controlPlaneTest.Spec.Image
		currentReplicas := *existingDeployment.Spec.Replicas
		desiredReplicas := replicasPerDeployment

		if currentImage != desiredImage || currentReplicas != desiredReplicas {
			fmt.Println("🖥️🔄📦 │ Drift détecté sur le Deployment")

			// =====================================================
			// MODIFICATION LOCALE DE LA STRUCTURE GO
			// =====================================================

			existingDeployment.Spec.Template.Spec.Containers[0].Image = desiredImage
			existingDeployment.Spec.Replicas = &desiredReplicas

			if currentImage == desiredImage && currentReplicas != desiredReplicas {
				fmt.Println("🤓🔄📦 │ Replicas :", currentReplicas, "→", desiredReplicas)
			}
			if currentReplicas == desiredReplicas && currentImage != desiredImage {
				fmt.Println("🤓🔄📦 │ Image :", currentImage, "→", desiredImage)
			}

			// =====================================================
			// UPDATE REEL DANS KUBERNETES
			// =====================================================
			// L'opérateur effectue une requête UPDATE vers l'API Server Kubernetes.
			//
			// Kubernetes appliquera ensuite :
			// - rolling update
			// - nouveau ReplicaSet
			// - remplacement Pods
			// etc...

			fmt.Println("🖥️🔄📦 │ UPDATE Deployment dans Kubernetes ...")

			err := r.Update(ctx, existingDeployment)
			if err != nil {

				fmt.Println("🖥️🔄❌ │ Impossible de mettre à jour le Deployment")
				typeReconciliation = "err_maj_deployment"
				erreursReconciliationTotal.Inc()
				return ctrl.Result{}, err
			}

			fmt.Println("🖥️🔄✅ │ Deployment mis à jour avec succès")
			typeReconciliation = "maj_deployment"
			return ctrl.Result{Requeue: true}, nil
		}

		fmt.Println("🖥️🎯✅ │ ", deploymentName, " convergé")
	}

	// =========================================================
	// CLEANUP DEPLOYMENTS EXCEDENTAIRES
	// =========================================================
	// Suppression des Deployments qui existent encore dans Kubernetes
	// mais qui ne sont plus présents dans l'état désiré de la CR.

	fmt.Println("🖥️🧹📦 │ Vérification des Deployments excédentaires ...")

	deploymentList := &appsv1.DeploymentList{}

	err = r.List(
		ctx,
		deploymentList,
		client.InNamespace(req.Namespace),
		client.MatchingLabels{
			"app": controlPlaneTest.Name,
		},
	)

	if err != nil {
		fmt.Println("🖥️🔍❌ │ Impossible de lister les Deployments")
		typeReconciliation = "err_list_deployments"
		erreursReconciliationTotal.Inc()
		return ctrl.Result{}, err
	}

	// Parcours des Deployments existants dans Kubernetes
	for _, deployment := range deploymentList.Items {
		// Vérifie si le Deployment fait encore partie
		// des ressources attendues par la CR
		expected := false

		for _, expectedDeploymentName := range deploymentNames {
			if deployment.Name == expectedDeploymentName {
				expected = true
				break
			}
		}
		// Deployment encore attendu
		if expected {
			continue
		}

		// =====================================
		// DEPLOYMENT EXCEDENTAIRE
		// =====================================
		fmt.Println(
			"🖥️🗑️📦 │ Suppression Deployment excédentaire :",
			deployment.Name,
		)

		err = r.Delete(
			ctx,
			&deployment,
		)

		if err != nil {
			fmt.Println("🖥️🗑️❌ │ Impossible de supprimer le Deployment")
			typeReconciliation = "err_auto-delete_deployment"
			erreursReconciliationTotal.Inc()
			return ctrl.Result{}, err
		}
	}

	// =========================================================
	// GESTION DES CONFIGMAPS
	// =========================================================
	// L'objectif est de maintenir une ConfigMap associée à la CR.
	// Kubernetes ne la crée PAS automatiquement.
	// C'est l'opérateur qui pilote entièrement cette ressource.

	configMapCount := controlPlaneTest.Spec.EtcdStress.ConfigMapCount

	if configMapCount <= 0 {
		configMapCount = 1
	}

	for i := int32(0); i < configMapCount; i++ {

		configMapName := fmt.Sprintf(
			"%s-configmap-%d",
			controlPlaneTest.Name,
			i,
		)

		configMapNames = append(
			configMapNames,
			configMapName,
		)

		existingConfigMap := &corev1.ConfigMap{}

		err := r.Get(
			ctx,
			types.NamespacedName{
				Name:      configMapName,
				Namespace: req.Namespace,
			},
			existingConfigMap,
		)

		if err != nil && apierrors.IsNotFound(err) {

			fmt.Println("🖥️⬆️📦 │ Création ConfigMap ...")

			configMapData := map[string]string{}
			configMapSizeKB := controlPlaneTest.Spec.EtcdStress.ConfigMapSizeKB

			if configMapSizeKB <= 0 {
				configMapSizeKB = 1
			}

			configMapData["payload"] = strings.Repeat(
				"A",
				int(configMapSizeKB*1024),
			)

			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      configMapName,
					Namespace: req.Namespace,
				},
				Data: configMapData,
			}

			err = ctrl.SetControllerReference(
				controlPlaneTest,
				configMap,
				r.Scheme,
			)

			if err != nil {

				fmt.Println("🖥️🔗❌ │ Impossible d'ajouter l'OwnerReference à la ConfigMap")
				erreursReconciliationTotal.Inc()
				return ctrl.Result{}, err
			}

			err = r.Create(ctx, configMap)

			if err != nil {

				fmt.Println("🖥️⬆️❌ │ Impossible de créer la ConfigMap")
				erreursReconciliationTotal.Inc()
				return ctrl.Result{}, err
			}

			fmt.Println("🖥️⬆️✅ │ ConfigMap créée :", configMapName)
			continue
		}

		if err != nil {

			fmt.Println("🖥️🔍❌ │ Impossible de récupérer la ConfigMap")
			erreursReconciliationTotal.Inc()
			return ctrl.Result{}, err
		}

		fmt.Println("🖥️🎯✅ │ ", configMapName, " convergée")
	}

	// =========================================================
	// CLEANUP CONFIGMAPS EXCEDENTAIRES
	// =========================================================
	// Suppression des ConfigMaps qui existent encore dans Kubernetes
	// mais qui ne sont plus présentes dans l'état désiré de la CR.

	fmt.Println("🖥️🧹📦 │ Vérification des ConfigMaps excédentaires ...")

	configMapList := &corev1.ConfigMapList{}

	err = r.List(
		ctx,
		configMapList,
		client.InNamespace(req.Namespace),
		client.MatchingLabels{
			"app": controlPlaneTest.Name,
		},
	)

	if err != nil {

		fmt.Println("🖥️🔍❌ │ Impossible de lister les ConfigMaps")
		typeReconciliation = "err_list_configmaps"
		erreursReconciliationTotal.Inc()
		return ctrl.Result{}, err
	}

	// Parcours des ConfigMaps existantes dans Kubernetes
	for _, configMap := range configMapList.Items {

		// Vérifie si la ConfigMap fait encore partie
		// des ressources attendues par la CR
		expected := false

		for _, expectedConfigMapName := range configMapNames {

			if configMap.Name == expectedConfigMapName {
				expected = true
				break
			}
		}

		// ConfigMap encore attendue
		if expected {
			continue
		}

		// =====================================
		// CONFIGMAP EXCEDENTAIRE
		// =====================================

		fmt.Println(
			"🖥️🗑️📦 │ Suppression ConfigMap excédentaire :",
			configMap.Name,
		)

		err = r.Delete(
			ctx,
			&configMap,
		)

		if err != nil {
			fmt.Println("🖥️🗑️❌ │ Impossible de supprimer la ConfigMap")
			typeReconciliation = "err_auto-delete_configmap"
			erreursReconciliationTotal.Inc()
			return ctrl.Result{}, err
		}
	}

	// =========================================================
	// GESTION DU SERVICE
	// =========================================================
	// L'objectif est de maintenir un Service associé au Deployment.
	// Kubernetes ne crée PAS automatiquement de Service.
	// C'est l'opérateur qui doit gérer cette ressource.

	existingService := &corev1.Service{}

	fmt.Println("🖥️🔍📦 │ Recherche du Service dans Kubernetes ...")

	err = r.Get(ctx, types.NamespacedName{
		Name:      serviceName,
		Namespace: req.Namespace,
	}, existingService)

	if err != nil && apierrors.IsNotFound(err) {

		fmt.Println("🖥️🔍🚫 │ Service introuvable")
		fmt.Println("🖥️⬆️📦 │ Création du Service ...")

		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName,
				Namespace: req.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": controlPlaneTest.Name,
				},
				Ports: []corev1.ServicePort{
					{
						Port:       80,
						TargetPort: intstr.FromInt(80),
					},
				},
			},
		}

		err = controllerutil.SetControllerReference(
			controlPlaneTest,
			service,
			r.Scheme,
		)

		if err != nil {
			fmt.Println("🖥️🔄❌ │ Impossible d'ajouter l'OwnerReference au Service")
			typeReconciliation = "err_owner-ref_svc"
			erreursReconciliationTotal.Inc()
			return ctrl.Result{}, err
		}

		fmt.Println("🖥️⬆️📦 │ POST Service vers l'API Kubernetes ...")

		err = r.Create(ctx, service)

		if err != nil {
			fmt.Println("🖥️⬆️❌ │ Impossible de créer le Service")
			typeReconciliation = "err_creation_svc"
			erreursReconciliationTotal.Inc()
			return ctrl.Result{}, err
		}

		fmt.Println("🖥️⬆️✅ │ Service créé avec succès")
		typeReconciliation = "creation_svc"
		return ctrl.Result{Requeue: true}, nil
	}

	if err != nil {
		fmt.Println("🖥️🔍❌ │ Erreur lors de la récupération du Service")
		typeReconciliation = "err_api_svc"
		erreursReconciliationTotal.Inc()
		return ctrl.Result{}, err
	}

	fmt.Println("🖥️🎯✅ │ Service convergé")

	// $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$
	// $$$$$$$$$$$$$$$$ STATUS $$$$$$$$$$$$$$$
	// $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$

	// =========================================================
	// RECUPERATION DU STATUS REEL
	// =========================================================
	// Ces informations proviennent directement du status calculé par Kubernetes lui-même.
	//
	// Ce n'est PAS l'opérateur qui calcule :
	// - ReadyReplicas
	// - AvailableReplicas
	//
	// Kubernetes met automatiquement à jour ces valeurs selon l'état réel du cluster.

	var totalReadyReplicas int32
	var totalAvailableReplicas int32

	for i := int32(0); i < deploymentCount; i++ {
		deploymentName := fmt.Sprintf(
			"%s-deployment-%d",
			controlPlaneTest.Name,
			i,
		)

		fmt.Println("🖥️📊📦 │ Récupération Status Deployment :", deploymentName)

		existingDeployment := &appsv1.Deployment{}

		err = r.Get(ctx, types.NamespacedName{
			Name:      deploymentName,
			Namespace: req.Namespace,
		}, existingDeployment)

		if err != nil {
			fmt.Println("🖥️🔍❌ │ Impossible de récupérer le Deployment pour le Status")
			typeReconciliation = "err_status_deployment"
			erreursReconciliationTotal.Inc()
			return ctrl.Result{}, err
		}

		// Agrégation globale des replicas Ready
		totalReadyReplicas += existingDeployment.Status.ReadyReplicas
		// Agrégation globale des replicas Available
		totalAvailableReplicas += existingDeployment.Status.AvailableReplicas
	}

	newReadyReplicas := totalReadyReplicas
	newAvailableReplicas := totalAvailableReplicas

	// METRIQUES
	// Nombre total de Deployments générés
	deploymentsGeneresGauge.WithLabelValues(
		controlPlaneTest.Name,
		controlPlaneTest.Namespace,
		controlPlaneTest.Spec.OperatorStress.Profile,
	).Set(
		float64(deploymentCount),
	)

	// Nombre total de Pods désirés
	podsDesiresGauge.WithLabelValues(
		controlPlaneTest.Name,
		controlPlaneTest.Namespace,
		controlPlaneTest.Spec.OperatorStress.Profile,
	).Set(
		float64(
			deploymentCount * replicasPerDeployment,
		),
	)

	// METRIQUE
	// Permet de définir la CR à afficher dans le dashboard pour en découler toutes les autres métriques reliées
	// ============================================================
	// CONFIGURATION SNAPSHOT METRICS
	// ============================================================

	topologySpread := "disabled"

	if controlPlaneTest.Spec.SchedulerStress.TopologySpread {
		topologySpread = "enabled"
	}

	aggressiveReconcile := "disabled"

	if controlPlaneTest.Spec.APIServerStress.AggressiveReconcile {
		aggressiveReconcile = "enabled"
	}

	frequentStatusUpdates := "disabled"

	if controlPlaneTest.Spec.APIServerStress.FrequentStatusUpdates {
		frequentStatusUpdates = "enabled"
	}

	recreateResources := "disabled"

	if controlPlaneTest.Spec.APIServerStress.RecreateResources {
		recreateResources = "enabled"
	}

	podStormEnabled := "disabled"

	if controlPlaneTest.Spec.PodLifecycleStorm.Enabled {
		podStormEnabled = "enabled"
	}

	deletePodsRandomly := "disabled"

	if controlPlaneTest.Spec.PodLifecycleStorm.DeletePodsRandomly {
		deletePodsRandomly = "enabled"
	}

	configurationScenarioInfo.WithLabelValues(
		controlPlaneTest.Name,
		controlPlaneTest.Namespace,
		controlPlaneTest.Spec.OperatorStress.Profile,
		fmt.Sprintf(
			"%t",
			controlPlaneTest.Spec.SchedulerStress.Enabled,
		),
		topologySpread,
		controlPlaneTest.Spec.SchedulerStress.Affinity,
		controlPlaneTest.Spec.SchedulerStress.AntiAffinity,
		aggressiveReconcile,
		frequentStatusUpdates,
		recreateResources,
		podStormEnabled,
		deletePodsRandomly,
	).Set(1)

	// ============================================
	// METRICS PROMETHEUS
	// ============================================

	// Mise à jour des métriques custom exposées sur /metrics
	podsGeresGauge.Set(float64(newReadyReplicas))
	replicasDisponiblesGauge.Set(float64(newAvailableReplicas))

	fmt.Println("☸️📊📦 │ ReadyReplicas :", newReadyReplicas)
	fmt.Println("☸️📊📦 │ AvailableReplicas :", newAvailableReplicas)

	// =========================================================
	// CONDITIONS KUBERNETES
	// =========================================================
	// Les Conditions représentent un état standardisé Kubernetes-native.
	//
	// Elles sont utilisées par :
	// - kubectl wait
	// - ArgoCD
	// - Prometheus
	// - dashboards
	// - GitOps
	// etc...

	condition := metav1.Condition{
		Type:               "Available",
		LastTransitionTime: metav1.Now(),
	}

	expectedReplicas := deploymentCount * replicasPerDeployment

	if newReadyReplicas == expectedReplicas {
		condition.Status = metav1.ConditionTrue
		condition.Reason = "DeploymentReady"
		condition.Message = "Deployments have enough ready replicas"

		fmt.Println("☸️📊✅ │ Condition Available=True")
	} else {
		condition.Status = metav1.ConditionFalse
		condition.Reason = "DeploymentNotReady"
		condition.Message = "Deployments do not have enough ready replicas"

		fmt.Println("☸️📊❌ │ Condition Available=False")
	}

	// Injection de la condition dans le status
	meta.SetStatusCondition(&controlPlaneTest.Status.Conditions, condition)

	// =========================================================
	// DETECTION DE DRIFT SUR LE STATUS
	// =========================================================
	// Evite :
	// - updates inutiles
	// - boucles infinies
	// - reconciliations permanentes

	statusChanged :=
		!reflect.DeepEqual(
			controlPlaneTest.Status.DeploymentNames,
			deploymentNames,
		) ||
			!reflect.DeepEqual(
				controlPlaneTest.Status.ConfigMapNames,
				configMapNames,
			) ||
			controlPlaneTest.Status.ServiceName != serviceName ||
			controlPlaneTest.Status.ReadyReplicas != newReadyReplicas ||
			controlPlaneTest.Status.AvailableReplicas != newAvailableReplicas ||
			controlPlaneTest.Status.ObservedGeneration != controlPlaneTest.Generation

	if statusChanged {

		fmt.Println("🖥️🔄📊 │ Mise à jour du Status ...")

		controlPlaneTest.Status.DeploymentNames = deploymentNames
		controlPlaneTest.Status.ServiceName = serviceName
		controlPlaneTest.Status.ConfigMapNames = configMapNames
		controlPlaneTest.Status.ReadyReplicas = newReadyReplicas
		controlPlaneTest.Status.AvailableReplicas = newAvailableReplicas

		// L'opérateur confirme qu'il a traité cette version du Spec utilisateur
		controlPlaneTest.Status.ObservedGeneration = controlPlaneTest.Generation

		fmt.Println("🖥️🔄📊 │ observedGeneration =", controlPlaneTest.Generation)

		// =====================================================
		// UPDATE DU SOUS-ENDPOINT /status
		// =====================================================
		// IMPORTANT :
		// r.Status().Update(...) ne modifie QUE :
		// status:
		// et PAS spec:
		//
		// Cela correspond au sous-endpoint Kubernetes : /status

		err = r.Status().Update(ctx, controlPlaneTest)
		if err != nil {
			fmt.Println("🖥️🔄❌ │ Impossible de mettre à jour le Status")
			typeReconciliation = "err_maj_status"
			erreursReconciliationTotal.Inc()
			return ctrl.Result{}, err
		}

		fmt.Println("🖥️🔄✅ │ Status mis à jour avec succès")
		typeReconciliation = "maj_status"
		return ctrl.Result{Requeue: true}, nil
	}

	// =========================================================
	// ************ API SERVER STRESS ***********
	// =========================================================
	// Cette section permet de générer davantage d'activité sur :
	// - kube-apiserver
	// - etcd
	// - informer cache
	// - controller-runtime
	//
	// Le stress est provoqué via :
	// - updates status fréquents
	// - suppressions/recréations ressources
	// - requeues agressifs
	// - multiplication appels API Kubernetes

	// =====================================
	// FREQUENT STATUS UPDATES
	// =====================================
	// Force des mises à jour Status très fréquentes
	// afin de générer davantage de PATCH/UPDATE
	// sur l'API Server et etcd.

	if controlPlaneTest.Spec.APIServerStress.Enabled &&
		controlPlaneTest.Spec.APIServerStress.FrequentStatusUpdates {

		fmt.Println("🖥️🔥📊 │ Frequent Status Updates activé")

		controlPlaneTest.Status.ObservedGeneration = controlPlaneTest.Generation
		controlPlaneTest.Status.ReadyReplicas = newReadyReplicas
		controlPlaneTest.Status.AvailableReplicas = newAvailableReplicas

		err = r.Status().Update(
			ctx,
			controlPlaneTest,
		)

		// METRIQUE
		// le nombre de mises à jour Status envoyées à l'API Server
		statusUpdatesTotal.WithLabelValues(
			controlPlaneTest.Name,
			controlPlaneTest.Namespace,
			controlPlaneTest.Spec.OperatorStress.Profile,
		).Inc()

		if err != nil {
			fmt.Println("🖥️🔥❌ │ Impossible de mettre à jour le Status")
			typeReconciliation = "err_frequent_status_update"
			erreursReconciliationTotal.Inc()
			return ctrl.Result{}, err
		}
	}

	// =====================================
	// AGGRESSIVE RECONCILE
	// =====================================
	// Force des réconciliations très fréquentes
	// via requeue automatique.

	if controlPlaneTest.Spec.APIServerStress.Enabled &&
		controlPlaneTest.Spec.APIServerStress.AggressiveReconcile {

		fmt.Println("🖥️🔥🔄 │ Aggressive Reconcile activé")

		// METRIQUE
		// le nombre de requeues forcés par l'opérateur
		requeuesForceesTotal.WithLabelValues(
			controlPlaneTest.Name,
			controlPlaneTest.Namespace,
			controlPlaneTest.Spec.OperatorStress.Profile,
		).Inc()

		requeueResult = &ctrl.Result{
			RequeueAfter: 1 * time.Second,
		}
	}

	// =====================================
	// RECREATE RESOURCES
	// =====================================
	// Force suppression/recréation systématique
	// des Deployments afin de provoquer :
	// - DELETE
	// - CREATE
	// - ReplicaSet churn
	// - Pod churn
	// - Events Kubernetes massifs

	if controlPlaneTest.Spec.APIServerStress.Enabled &&
		controlPlaneTest.Spec.APIServerStress.RecreateResources {
		deploymentList := &appsv1.DeploymentList{}

		err = r.List(
			ctx,
			deploymentList,
			client.InNamespace(req.Namespace),
			client.MatchingLabels{
				"app": controlPlaneTest.Name,
			},
		)

		if err != nil {
			fmt.Println("🖥️🔥❌ │ Impossible de lister les Deployments")
			typeReconciliation = "err_list_recreate"
			erreursReconciliationTotal.Inc()
			return ctrl.Result{}, err
		}

		fmt.Println("🖥️🔥🗑️ │ Recreate Resources activé")
		for _, deployment := range deploymentList.Items {
			fmt.Println(
				"🖥️🔥🗑️📦 │ Suppression forcée Deployment :",
				deployment.Name,
			)

			err = r.Delete(
				ctx,
				&deployment,
			)

			// METRIQUE
			// Mesure le nombre de ressources supprimées/recréées
			resourcesRecreatedTotal.WithLabelValues(
				controlPlaneTest.Name,
				controlPlaneTest.Namespace,
				controlPlaneTest.Spec.OperatorStress.Profile,
			).Inc()

			if err != nil {
				fmt.Println("🖥️🔥❌ │ Impossible de supprimer le Deployment")
				typeReconciliation = "err_recreate_delete"
				erreursReconciliationTotal.Inc()
				return ctrl.Result{}, err
			}
		}

		if requeueResult == nil {
			requeueResult = &ctrl.Result{
				RequeueAfter: 2 * time.Second,
			}
		}
	}

	// =========================================================
	// ******** POD LIFECYCLE STORM ********
	// =========================================================
	// Génère volontairement des tempêtes de cycle de vie Pods :
	// - suppressions aléatoires
	// - recréations ReplicaSets
	// - rescheduling
	// - events Kubernetes massifs

	if controlPlaneTest.Spec.PodLifecycleStorm.Enabled {
		fmt.Println("🖥️🌪️📦 │ Pod Lifecycle Storm activé")

		// =====================================
		// RECUPERATION PODS
		// =====================================

		podList := &corev1.PodList{}
		var pendingPods int

		err = r.List(
			ctx,
			podList,
			client.InNamespace(req.Namespace),
			client.MatchingLabels{
				"app": controlPlaneTest.Name,
			},
		)

		if err != nil {
			fmt.Println("🖥️🌪️❌ │ Impossible de récupérer les Pods")
			typeReconciliation = "err_podstorm_list"
			erreursReconciliationTotal.Inc()
			return ctrl.Result{}, err
		}
		// METRIQUE
		// Comptage Pods Pending
		for _, pod := range podList.Items {

			if pod.Status.Phase == corev1.PodPending {
				pendingPods++
			}
		}

		// Mise à jour métrique Pending Pods
		podsPendingGauge.WithLabelValues(
			controlPlaneTest.Name,
			controlPlaneTest.Namespace,
			controlPlaneTest.Spec.OperatorStress.Profile,
		).Set(
			float64(pendingPods),
		)

		// =====================================
		// DELETE RANDOM POD
		// =====================================
		if controlPlaneTest.Spec.PodLifecycleStorm.DeletePodsRandomly {
			if len(podList.Items) > 0 {
				randomIndex := rand.Intn(len(podList.Items))
				randomPod := podList.Items[randomIndex]

				fmt.Println(
					"🖥️🌪️🗑️📦 │ Suppression aléatoire Pod :",
					randomPod.Name,
				)

				err = r.Delete(
					ctx,
					&randomPod,
				)

				// METRIQUE
				// Mesure le nombre total de Pods supprimés par le PodLifecycleStorm
				podsSupprimesTotal.WithLabelValues(
					controlPlaneTest.Name,
					controlPlaneTest.Namespace,
					controlPlaneTest.Spec.OperatorStress.Profile,
				).Inc()

				if err != nil {
					fmt.Println("🖥️🌪️❌ │ Impossible de supprimer le Pod")
					typeReconciliation = "err_podstorm_delete"
					erreursReconciliationTotal.Inc()
					return ctrl.Result{}, err
				}
			}
		}

		// =====================================
		// REQUEUE PERIODIQUE
		// =====================================

		if controlPlaneTest.Spec.PodLifecycleStorm.RestartPodsEverySeconds > 0 {
			restartDelay := time.Duration(
				controlPlaneTest.Spec.PodLifecycleStorm.RestartPodsEverySeconds,
			) * time.Second

			fmt.Println(
				"🖥️🌪️🔄 │ Requeue Pod Storm :",
				restartDelay,
			)

			if requeueResult == nil {
				requeueResult = &ctrl.Result{
					RequeueAfter: restartDelay,
				}
			}
		}
	}

	fmt.Println("🖥️🎯✅ │ Status déjà convergé")
	fmt.Println("\n================ FIN RECONCILIATION =================")
	// =========================================================
	// FIN DE RECONCILIATION
	// =========================================================
	// Etat désiré == Etat réel
	// La convergence est atteinte.

	fmt.Println("🖥️🎯✅ │ Réconciliation terminée avec succès")
	typeReconciliation = "reconciliation_complete"
	if requeueResult != nil {
		return *requeueResult, nil
	}

	return ctrl.Result{}, nil
}

// Fonction permettant de détecter si un Pod passe en état "Ready" (prêt à être déployé)
func isPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControlPlaneTestReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// Indexation locale au controleur de notre ressource (permet de retrouver rapidement
	// les parents sans tester de multiples ressources)
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&appsv1.Deployment{},
		deploymentOwnerKey,
		func(rawObj client.Object) []string {

			deployment := rawObj.(*appsv1.Deployment)

			// Récupération du propriétaire de l'objet Deployment
			owner := metav1.GetControllerOf(deployment)

			if owner == nil {
				return nil
			}

			// Vérifie que le propriétaire est bien une CR du bon type
			if owner.APIVersion != controlplanev1alpha1.GroupVersion.String() ||
				owner.Kind != "ControlPlaneTest" {
				return nil
			}

			// Valeur indexée
			return []string{owner.Name}
		},
	); err != nil {
		return err
	}

	// On ajoute un controller au manager
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&controlplanev1alpha1.ControlPlaneTest{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		// Ajout d'un watch sur les ressources secondaires (ici le deployment dépendant de la CR)
		// permet de réagir en cas de modification du deployment directement sans passer par la CR
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Pod{}).
		Named("controlplanetest").
		// On ajoute des limmitations :
		// - MaxConcurrentReconciles : nbr max de réconciliation simultanées
		// - RateLimiter : Attente de 1s puis 2s ... jusqu'à 30s avant chaque nouvelle réconciliation
		// - BucketRateLimiter : Limite le nombre d'évènement générés par le controleur à la seconde de 10 à 100/sec
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
			RateLimiter: workqueue.NewTypedMaxOfRateLimiter(
				workqueue.NewTypedItemExponentialFailureRateLimiter[reconcile.Request](
					1*time.Second,
					30*time.Second,
				),
				&workqueue.TypedBucketRateLimiter[reconcile.Request]{
					Limiter: rate.NewLimiter(
						rate.Limit(10),
						100,
					),
				},
			),
		}).
		Complete(r)
}
