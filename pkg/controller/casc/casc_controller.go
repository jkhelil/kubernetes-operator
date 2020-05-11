package casc

import (
	"context"
	"fmt"

	"github.com/jenkinsci/kubernetes-operator/pkg/apis/jenkins/v1alpha2"
	jenkinsclient "github.com/jenkinsci/kubernetes-operator/pkg/controller/jenkins/client"
	"github.com/jenkinsci/kubernetes-operator/pkg/controller/jenkins/configuration"
	"github.com/jenkinsci/kubernetes-operator/pkg/controller/jenkins/configuration/user"
	"github.com/jenkinsci/kubernetes-operator/pkg/controller/jenkins/notifications/event"
	"github.com/jenkinsci/kubernetes-operator/pkg/controller/jenkins/notifications/reason"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	jenkinsv1alpha3 "github.com/jenkinsci/kubernetes-operator/pkg/apis/jenkins/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_casc")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Casc Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
/*func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}*/

func Add(mgr manager.Manager, jenkinsAPIConnectionSettings jenkinsclient.JenkinsAPIConnectionSettings, clientSet kubernetes.Clientset, config rest.Config, notificationEvents *chan event.Event) error {
	reconciler := newReconciler(mgr, jenkinsAPIConnectionSettings, clientSet, config, notificationEvents)
	return add(mgr, reconciler)
}

// newReconciler returns a new reconcile.Reconciler.
func newReconciler(mgr manager.Manager, jenkinsAPIConnectionSettings jenkinsclient.JenkinsAPIConnectionSettings, clientSet kubernetes.Clientset, config rest.Config, notificationEvents *chan event.Event) reconcile.Reconciler {
	return &ReconcileCasc{
		client:                       mgr.GetClient(),
		scheme:                       mgr.GetScheme(),
		jenkinsAPIConnectionSettings: jenkinsAPIConnectionSettings,
		clientSet:                    clientSet,
		config:                       config,
		notificationEvents:           notificationEvents,
	}
}

// newReconciler returns a new reconcile.Reconciler
/*func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCasc{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}*/

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("casc-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Casc
	err = c.Watch(&source.Kind{Type: &jenkinsv1alpha3.Casc{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	/* TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Casc
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &jenkinsv1alpha3.Casc{},
	})
	if err != nil {
		return err
	}*/

	return nil
}

// blank assignment to verify that ReconcileCasc implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCasc{}

// ReconcileCasc reconciles a Casc object
type ReconcileCasc struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client                       client.Client
	scheme                       *runtime.Scheme
	jenkinsAPIConnectionSettings jenkinsclient.JenkinsAPIConnectionSettings
	clientSet                    kubernetes.Clientset
	config                       rest.Config
	notificationEvents           *chan event.Event
}

// Reconcile reads that state of the cluster for a Casc object and makes changes based on the state read
// and what is in the Casc.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCasc) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Casc")

	// Fetch the Casc instance
	casc := &jenkinsv1alpha3.Casc{}
	err := r.client.Get(context.TODO(), request.NamespacedName, casc)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// fetch the jenkins CR
	jenkins := &v1alpha2.Jenkins{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: casc.Spec.JenkinsRef.Name, Namespace: casc.Namespace}, jenkins)
	if err != nil {
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	config := configuration.Configuration{
		Client:                       r.client,
		ClientSet:                    r.clientSet,
		Notifications:                r.notificationEvents,
		Jenkins:                      jenkins,
		Scheme:                       r.scheme,
		Config:                       &r.config,
		JenkinsAPIConnectionSettings: r.jenkinsAPIConnectionSettings,
	}

	jenkinsClient, err := config.GetJenkinsClient()
	if err != nil {
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	reqLogger.Info("Jenkins API client set")

	// Reconcile user configuration
	userConfiguration := user.New(config, jenkinsClient, reqLogger)
	var messages []string
	messages, err = userConfiguration.Validate(jenkins)
	if err != nil {
		return reconcile.Result{}, err
	}
	if len(messages) > 0 {
		message := "Validation of user configuration failed, please correct Jenkins CR"
		*r.notificationEvents <- event.Event{
			Jenkins: *jenkins,
			Phase:   event.PhaseUser,
			Level:   v1alpha2.NotificationLevelWarning,
			Reason:  reason.NewUserConfigurationFailed(reason.HumanSource, []string{message}, append([]string{message}, messages...)...),
		}

		reqLogger.Info(message)
		for _, msg := range messages {
			reqLogger.Info(msg)
		}
		return reconcile.Result{}, nil // don't requeue
	}

	result, err := userConfiguration.Reconcile()
	if err != nil {
		return reconcile.Result{}, err
	}
	if result.Requeue {
		return result, nil
	}

	if jenkins.Status.UserConfigurationCompletedTime == nil {
		now := metav1.Now()
		jenkins.Status.UserConfigurationCompletedTime = &now
		err = r.client.Update(context.TODO(), jenkins)
		if err != nil {
			return reconcile.Result{}, err
		}
		message := fmt.Sprintf("User configuration phase is complete, took %s",
			jenkins.Status.UserConfigurationCompletedTime.Sub(jenkins.Status.ProvisionStartTime.Time))
		*r.notificationEvents <- event.Event{
			Jenkins: *jenkins,
			Phase:   event.PhaseUser,
			Level:   v1alpha2.NotificationLevelInfo,
			Reason:  reason.NewUserConfigurationComplete(reason.OperatorSource, []string{message}),
		}
		reqLogger.Info(message)
	}

	// Define a new Pod object
	/*pod := newPodForCR(instance)

	// Set Casc instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	*/
	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *jenkinsv1alpha3.Casc) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
