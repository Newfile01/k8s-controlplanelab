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
	"time"

	"golang.org/x/time/rate"
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

	builder "sigs.k8s.io/controller-runtime/pkg/builder"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	controlplanev1alpha1 "github.com/Newfile01/k8s-controlplanelab/operator/api/v1alpha1"
)

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
	_ = logf.FromContext(ctx)

	fmt.Println("\n\n================ RECONCILIATION =================")

	// =========================================================
	// RECUPERATION DE LA CUSTOM RESOURCE
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
	// GESTION DU FINALIZER
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

				return ctrl.Result{}, err
			}
		}

		// Une fois le dernier finalizer supprimé : Kubernetes supprimera réellement la ressource.
		fmt.Println("🖥️🗑️✅ │ Finalizer supprimé, Kubernetes peut terminer la suppression")

		return ctrl.Result{}, nil
	}

	// =========================================================
	// GESTION DU DEPLOYMENT
	// =========================================================
	// L'objectif est de maintenir :
	// Etat désiré (CR) == Etat réel (Deployment)
	// Toute divergence détectée sera corrigée automatiquement par l'opérateur.

	// Noms attendus
	deploymentName := controlPlaneTest.Name + "-deployment"
	serviceName := controlPlaneTest.Name + "-service"
	configMapName := controlPlaneTest.Name + "-configmap"

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
				Replicas: &controlPlaneTest.Spec.Replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": controlPlaneTest.Name,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": controlPlaneTest.Name,
						},
					},
					Spec: corev1.PodSpec{
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

			// En cas d'erreur : controller-runtime replanifiera automatiquement une nouvelle réconciliation.
			return ctrl.Result{}, err
		}

		fmt.Println("🖥️⬆️✅ │ Deployment créé avec succès")

		// Requeue : relance immédiate pour relire l'état réel créé par Kubernetes.
		return ctrl.Result{Requeue: true}, nil
	}

	if err != nil {

		fmt.Println("🖥️🔍❌ │ Erreur lors de la récupération du Deployment")

		return ctrl.Result{}, err
	}

	// =========================================================
	// GESTION DE LA CONFIGMAP
	// =========================================================
	// L'objectif est de maintenir une ConfigMap associée à la CR.
	// Kubernetes ne la crée PAS automatiquement.
	// C'est l'opérateur qui pilote entièrement cette ressource.

	existingConfigMap := &corev1.ConfigMap{}

	fmt.Println("🖥️🔍📦 │ Recherche de la ConfigMap dans Kubernetes ...")

	err = r.Get(ctx, types.NamespacedName{
		Name:      configMapName,
		Namespace: req.Namespace,
	}, existingConfigMap)

	if err != nil && apierrors.IsNotFound(err) {

		fmt.Println("🖥️🔍🚫 │ ConfigMap introuvable")
		fmt.Println("🖥️⬆️📦 │ Création de la ConfigMap ...")

		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapName,
				Namespace: req.Namespace,
			},
			Data: map[string]string{
				"app.conf": "hello-from-operator",
			},
		}

		err = controllerutil.SetControllerReference(
			controlPlaneTest,
			configMap,
			r.Scheme,
		)
		if err != nil {

			fmt.Println("🖥️🔄❌ │ Impossible d'ajouter l'OwnerReference à la ConfigMap")

			return ctrl.Result{}, err
		}

		fmt.Println("🖥️⬆️📦 │ POST ConfigMap vers l'API Kubernetes ...")

		err = r.Create(ctx, configMap)
		if err != nil {

			fmt.Println("🖥️⬆️❌ │ Impossible de créer la ConfigMap")

			return ctrl.Result{}, err
		}

		fmt.Println("🖥️⬆️✅ │ ConfigMap créée avec succès")

		return ctrl.Result{Requeue: true}, nil
	}

	if err != nil {

		fmt.Println("🖥️🔍❌ │ Erreur lors de la récupération de la ConfigMap")

		return ctrl.Result{}, err
	}

	fmt.Println("🖥️🎯✅ │ ConfigMap déjà convergée")

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

			return ctrl.Result{}, err
		}

		fmt.Println("🖥️⬆️📦 │ POST Service vers l'API Kubernetes ...")

		err = r.Create(ctx, service)
		if err != nil {

			fmt.Println("🖥️⬆️❌ │ Impossible de créer le Service")

			return ctrl.Result{}, err
		}

		fmt.Println("🖥️⬆️✅ │ Service créé avec succès")

		return ctrl.Result{Requeue: true}, nil
	}

	if err != nil {

		fmt.Println("🖥️🔍❌ │ Erreur lors de la récupération du Service")

		return ctrl.Result{}, err
	}

	fmt.Println("🖥️🎯✅ │ Service déjà convergé")

	// =========================================================
	// VERIFICATION DE LA STRUCTURE DU DEPLOYMENT
	// =========================================================
	// Protection contre l'accès à un tableau vide.
	// Sans cela : panic possible sur containers[0]

	containers := existingDeployment.Spec.Template.Spec.Containers

	if len(containers) < 1 {

		fmt.Println("☸️📦❌ │ Aucun container trouvé dans le Deployment")

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
	desiredReplicas := controlPlaneTest.Spec.Replicas

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

			return ctrl.Result{}, err
		}

		fmt.Println("🖥️🔄✅ │ Deployment mis à jour avec succès")

		return ctrl.Result{Requeue: true}, nil
	}

	fmt.Println("🖥️🎯✅ │ Deployment déjà convergé")

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

	newDeploymentName := existingDeployment.Name
	newReadyReplicas := existingDeployment.Status.ReadyReplicas
	newAvailableReplicas := existingDeployment.Status.AvailableReplicas

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

	if newReadyReplicas == desiredReplicas {

		condition.Status = metav1.ConditionTrue
		condition.Reason = "DeploymentReady"
		condition.Message = "Deployment has enough ready replicas"

		fmt.Println("☸️📊✅ │ Condition Available=True")

	} else {

		condition.Status = metav1.ConditionFalse
		condition.Reason = "DeploymentNotReady"
		condition.Message = "Deployment does not have enough ready replicas"

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

	if controlPlaneTest.Status.DeploymentName != newDeploymentName ||
		controlPlaneTest.Status.ReadyReplicas != newReadyReplicas ||
		controlPlaneTest.Status.AvailableReplicas != newAvailableReplicas ||
		controlPlaneTest.Status.ServiceName != serviceName ||
		controlPlaneTest.Status.ConfigMapName != configMapName ||
		specChanged {

		fmt.Println("🖥️🔄📊 │ Mise à jour du Status ...")

		controlPlaneTest.Status.DeploymentName = newDeploymentName
		controlPlaneTest.Status.ServiceName = serviceName
		controlPlaneTest.Status.ConfigMapName = configMapName
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

			return ctrl.Result{}, err
		}

		fmt.Println("🖥️🔄✅ │ Status mis à jour avec succès")

		return ctrl.Result{Requeue: true}, nil
	}

	fmt.Println("🖥️🎯✅ │ Status déjà convergé")

	fmt.Println("\n================ FIN RECONCILIATION =================")

	// =========================================================
	// FIN DE RECONCILIATION
	// =========================================================
	// Etat désiré == Etat réel
	// La convergence est atteinte.

	fmt.Println("🖥️🎯✅ │ Réconciliation terminée avec succès")

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

	// Ajout d'un Predicat = Filtres pour n'effectuer de watch que sur les ressources concernées
	// Les autres Pods seront ignorés
	podPredicate := predicate.Funcs{
		// Fonciton teste et retourne un booléen pour :
		// Evènement de création d'un Pod
		CreateFunc: func(e event.CreateEvent) bool {

			pod, ok := e.Object.(*corev1.Pod)
			if !ok {
				return false
			}

			// Ignore Pods sans label app
			// Si le label "app" n'existe pas pour le Pod (!exist) on ignore
			if _, exists := pod.Labels["app"]; !exists {

				fmt.Println("☸️🔍🚫 │ Pod ignoré (pas de label app)")

				return false
			}

			// Sinon on accepte
			fmt.Println("☸️🔍✅ │ Pod accepté par predicate")

			return true
		},

		// Evènement de MàJ d'un Pod
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldPod, ok := e.ObjectOld.(*corev1.Pod)
			if !ok {
				return false
			}
			newPod, ok := e.ObjectNew.(*corev1.Pod)
			if !ok {
				return false
			}
			// Ignore Pods sans label app
			if _, exists := newPod.Labels["app"]; !exists {
				return false
			}
			// ============================================
			// CHANGEMENT DE PHASE
			// ============================================
			if oldPod.Status.Phase != newPod.Status.Phase {
				fmt.Println("☸️🔄📦 │ Changement phase Pod détecté")
				return true
			}
			// ============================================
			// POD EN COURS DE SUPPRESSION
			// ============================================
			if oldPod.DeletionTimestamp == nil &&
				newPod.DeletionTimestamp != nil {
				fmt.Println("☸️🗑️📦 │ Suppression Pod détectée")
				return true
			}

			// ============================================
			// CHANGEMENT READY
			// ============================================
			oldReady := isPodReady(oldPod)
			newReady := isPodReady(newPod)

			if oldReady != newReady {
				fmt.Println("☸️🔄📊 │ Changement Ready Pod détecté")
				return true
			}

			return false
		},

		// Evènement de MàJ d'un Pod
		DeleteFunc: func(e event.DeleteEvent) bool {

			pod, ok := e.Object.(*corev1.Pod)
			if !ok {
				return false
			}

			if _, exists := pod.Labels["app"]; !exists {
				return false
			}

			return true
		},

		// Evènement de par défaut appliqué à un Pod
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
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
		Named("controlplanetest").
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {

					fmt.Println("☸️🔍📦 │ Event Pod intercepté")

					pod := obj.(*corev1.Pod)

					// ============================================
					// RECUPERATION DU DEPLOYMENT PARENT
					// ============================================

					// Recherche du label "app" (label ajouté automatiquement dans le podTemplate du deployoment généré par notre CR)
					//  propagé automatiquement par Kubernetes depuis le PodTemplate du Deployment
					appLabel := pod.Labels["app"]

					if appLabel == "" {

						fmt.Println("☸️🔍🚫 │ Aucun label app trouvé sur le Pod")

						return nil
					}

					// ============================================
					// RECHERCHE DES DEPLOYMENTS INDEXES
					// ============================================

					deploymentList := &appsv1.DeploymentList{}

					// Vérification de correspondance Pod -> Deployment
					// Recherche des Deployments correspondant au label app du Pod
					err := r.List(
						ctx,
						deploymentList,
						client.InNamespace(pod.Namespace),
						client.MatchingLabels{
							"app": appLabel,
						},
					)
					if err != nil {

						fmt.Println("🖥️🔍❌ │ Impossible de retrouver le Deployment parent")

						return nil
					}

					// ============================================
					// MAPPING DEPLOYMENT -> CR
					// ============================================

					var requests []reconcile.Request

					// "Pour, valeur initiale non prise en compte, et index dans la liste des Deployment correspondant à 'deployment',
					// récupérer son owner, si nul ou ne correspond pas à ControlPlaneTest kind on continu, sinon on met en file d'attente
					// la requête de réconciliation avec les bons paramètres directement"
					for _, deployment := range deploymentList.Items {

						owner := metav1.GetControllerOf(&deployment)

						if owner == nil {
							continue
						}

						if owner.APIVersion != controlplanev1alpha1.GroupVersion.String() ||
							owner.Kind != "ControlPlaneTest" {
							continue
						}

						fmt.Println("☸️🔍📦 │ CR parente retrouvée :", owner.Name)

						requests = append(
							requests,
							reconcile.Request{
								NamespacedName: types.NamespacedName{
									Name:      owner.Name,
									Namespace: deployment.Namespace,
								},
							},
						)
					}

					return requests
				},
			),
			builder.WithPredicates(podPredicate),
		).
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
