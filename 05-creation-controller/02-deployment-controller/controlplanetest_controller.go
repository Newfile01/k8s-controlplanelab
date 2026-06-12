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

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	controlplanev1alpha1 "github.com/Newfile01/k8s-controlplanelab/operator/api/v1alpha1"
)

// ControlPlaneTestReconciler reconciles a ControlPlaneTest object
type ControlPlaneTestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
	_ = logf.FromContext(ctx)

	fmt.Println("\n\n======= NOUVELLE BOUCLE DE RECONCILIATION =========")

	// On créé une structure Go vide qui accueillera la Custom Resource
	// récupérée depuis l'API Kubernetes
	controlPlaneTest := &controlplanev1alpha1.ControlPlaneTest{}

	// On récupère la ressource correspondant à la requête reçue
	// req.NamespacedName contient :
	// - le nom
	// - le namespace
	// de la ressource à traiter
	//
	// Le controller fait une requête GET vers l'API Server
	// et récupère la réponse dans controlPlaneTest
	fmt.Println("Récupération des informations de la Custom Resource ...")

	err := r.Get(ctx, req.NamespacedName, controlPlaneTest)
	if err != nil {

		// Si la ressource n'existe plus :
		// - suppression utilisateur
		// - namespace supprimé
		// - objet inexistant
		//
		// IgnoreNotFound() évite de considérer cela (suppression, etc.) comme une erreur réelle
		// et stoppe simplement cette réconciliation
		fmt.Println("Aucune ressource correspondante trouvée")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Nom du Deployment que l'on souhaite associer à la CR
	deploymentName := controlPlaneTest.Name + "-deployment"

	// Structure vide destinée à accueillir le Deployment si il existe déjà
	existingDeployment := &appsv1.Deployment{}

	fmt.Println("Vérification de l'existence du Deployment dans le cluster ...")

	// Vérification de l'existence du Deployment
	err = r.Get(ctx, types.NamespacedName{
		Name:      deploymentName,
		Namespace: req.Namespace,
	}, existingDeployment)

	// Si le Deployment n'existe pas :
	// création du Deployment
	if err != nil && apierrors.IsNotFound(err) {

		fmt.Println("Création du Deployment ...")

		// Définition du Deployment à créer
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

		// Ajout d'une OwnerReference :
		// la CR devient propriétaire du Deployment
		//
		// Si la CR est supprimée :
		// Kubernetes supprimera automatiquement le Pod
		err = controllerutil.SetControllerReference(
			controlPlaneTest,
			deployment,
			r.Scheme,
		)
		if err != nil {
			return ctrl.Result{}, err
		}

		// Création du Deployment via l'API Server
		//
		// Le controller effectue ici une requête POST Kubernetes
		err = r.Create(ctx, deployment)
		if err != nil {

			fmt.Println("Impossible de créer le Deployment")

			// En cas d'erreur :
			// controller-runtime replanifiera automatiquement
			// une nouvelle réconciliation
			return ctrl.Result{}, err
		}

		fmt.Println("Deployment créé avec succès")

		// Requeue :
		// on relance immédiatement une nouvelle réconciliation
		// afin de relire l'état réel du Deployment nouvellement créé
		return ctrl.Result{Requeue: true}, nil
	}

	// Si une autre erreur que "NotFound" survient :
	// erreur RBAC, timeout API Server, problème réseau...
	if err != nil {
		fmt.Println("Erreur lors de la récupération du Deployment")
		return ctrl.Result{}, err
	}

	// Vérification de la convergence entre état désiré (CR)
	// et état réel (Deployment)

	// Protection contre l'accès au tableau vide
	containers := existingDeployment.Spec.Template.Spec.Containers

	if len(containers) < 1 {

		fmt.Println("Aucun container trouvé dans le Deployment")

		return ctrl.Result{}, fmt.Errorf(
			"deployment %s ne contient aucun container",
			existingDeployment.Name,
		)
	}

	currentImage := containers[0].Image
	desiredImage := controlPlaneTest.Spec.Image
	currentReplicas := *existingDeployment.Spec.Replicas
	desiredReplicas := controlPlaneTest.Spec.Replicas

	if currentImage != desiredImage || currentReplicas != desiredReplicas {

		existingDeployment.Spec.Template.Spec.Containers[0].Image = desiredImage
		existingDeployment.Spec.Replicas = &desiredReplicas

		fmt.Println("MàJ du Deployment ...")
		if currentImage == desiredImage && currentReplicas != desiredReplicas {
			fmt.Println("Nombre de Replicas passe de ", currentReplicas, "à ", desiredReplicas)
		}
		if currentReplicas == desiredReplicas && currentImage != desiredImage {
			fmt.Println("Image passe de ", currentImage, "à ", desiredImage)
		}
		err := r.Update(ctx, existingDeployment)

		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	fmt.Println("Deployment déjà existant")

	// Etat réel observé
	newDeploymentName := existingDeployment.Name
	newReadyReplicas := existingDeployment.Status.ReadyReplicas
	newAvailableReplicas := existingDeployment.Status.AvailableReplicas

	// Mise à jour Status si changement
	if controlPlaneTest.Status.DeploymentName != newDeploymentName ||
		controlPlaneTest.Status.ReadyReplicas != newReadyReplicas ||
		controlPlaneTest.Status.AvailableReplicas != newAvailableReplicas {

		fmt.Println("Mise à jour du Status ...")

		controlPlaneTest.Status.DeploymentName =
			newDeploymentName

		controlPlaneTest.Status.ReadyReplicas =
			newReadyReplicas

		controlPlaneTest.Status.AvailableReplicas =
			newAvailableReplicas

		// UPDATE /status
		err = r.Status().Update(ctx, controlPlaneTest)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	fmt.Println("Status déjà convergé")

	fmt.Println("Deployment mis à jour avec succès")

	fmt.Println("\n======= FIN BOUCLE DE RECONCILIATION =========")

	// Réconciliation terminée :
	// l'état réel correspond maintenant à l'état désiré
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControlPlaneTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&controlplanev1alpha1.ControlPlaneTest{}).
		Named("controlplanetest").
		Complete(r)
}
