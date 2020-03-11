package pingdomcheck

import (
	"context"

	monitoringv1alpha1 "github.com/adrianRiobo/pingdom-operator/pkg/apis/monitoring/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
        "github.com/russellcardullo/go-pingdom/pingdom"
        "os"
)

var log = logf.Log.WithName("controller_pingdomcheck")

// Add creates a new PingdomCheck Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
        //add here pingdom client initialization
        user := os.Getenv("PD_USERNAME") 
        if user == "" {
	  log.Info("PD_USERNAME should be defined as ENV")
        } else {
          log.Info("Getting user for pingdom", "user", user)
        }
        password := os.Getenv("PD_PASSWORD")
        if password == "" {
          log.Info("PD_PASSWORD should be defined as ENV")
        } else {
          log.Info("Getting user for pingdom", "password", password)
        }
        apikey := os.Getenv("PD_APIKEY")
        if apikey == "" {
          log.Info("PD_APIKEY should be defined as ENV")
        } else {
          log.Info("Getting user for pingdom", "apikey", apikey)
        }
        pingdomClient, err := pingdom.NewClientWithConfig(pingdom.ClientConfig{
              User:     user,
              Password: password,
              APIKey:   apikey,
        })
        if err != nil {
          log.Info("Error creating pingdom client")
        } else {
          log.Info("Info de client", "", pingdomClient)
        }
        pingdomChecks, err := pingdomClient.Checks.List()
        if err != nil {
          log.Error(err, "Error listing checks")
        }
        log.Info("All checks intial:", "all checks", pingdomChecks)
        
	return &ReconcilePingdomCheck{client: mgr.GetClient(), scheme: mgr.GetScheme(), pingdomClient: *pingdomClient}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("pingdomcheck-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PingdomCheck
	err = c.Watch(&source.Kind{Type: &monitoringv1alpha1.PingdomCheck{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner PingdomCheck
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &monitoringv1alpha1.PingdomCheck{},
	})
	if err != nil {
		return err
	}
     
	return nil
}

// blank assignment to verify that ReconcilePingdomCheck implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcilePingdomCheck{}

// ReconcilePingdomCheck reconciles a PingdomCheck object
type ReconcilePingdomCheck struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
        pingdomClient pingdom.Client
}

// Reconcile reads that state of the cluster for a PingdomCheck object and makes changes based on the state read
// and what is in the PingdomCheck.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePingdomCheck) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling PingdomCheck")

	// Fetch the PingdomCheck instance
	instance := &monitoringv1alpha1.PingdomCheck{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
        // Get Secret
        founds := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "pingdom-credentials", Namespace: request.Namespace}, founds)
	if err != nil {
	 	if errors.IsNotFound(err) {
			return reconcile.Result{Requeue: true}, nil
		} else {
                    
			return reconcile.Result{}, err
		}
	}
        reqLogger.Info("Getting Secret", "Secret.Name", founds.Name, "Secret.Data.username", founds.Data["username"])
 
        if e := os.Getenv("PD_USERNAME"); e != "" {
		reqLogger.Info("Getting Secret", "Secret.Data.username", e)
	}
        

        // Get configmap to configure pingdomclient
        // Update the App status with the pod names.
	// List the pods for this app's deployment.
	configMapList := &corev1.ConfigMapList{}
	listOpts := []client.ListOption{
		client.InNamespace(request.Namespace),
	}
	if err = r.client.List(context.TODO(), configMapList, listOpts...); err != nil {
		reqLogger.Error(err, "error")
	}
        for i, configMap := range configMapList.Items {
                reqLogger.Info("Getting Configmap", "Configmap,Name", i, "BAD", configMap.Name)
	}


        //var pingdomConfig corev1.ConfigMapList
        //if err := r.client.List(context.TODO(), "", &pingdomConfig); err != nil {
        //  reqLogger.Error(err, "error")
        //}
        //for _, pingdomConfig := range pingdomConfig.Items {
        //  reqLogger.Info("name=%q", *pingdomConfig.Metadata.Name)
        //}
        
        // Added client to pingdom
        //pingdomClient, err := pingdom.NewClientWithConfig(pingdom.ClientConfig{
	//	User:     "usermail",
	//	Password: "pass",
	//	APIKey:   "apikey",
	//})
        pingdomChecks, _ := r.pingdomClient.Checks.List()
        //for _, pingdomCheck := range pingdomChecks.Items {
        //  reqLogger.Info("Pingdom CHeck names:", "pingdom check name", pingdomCheck.Name)
        //}

        reqLogger.Info("All checks:", "all checks", pingdomChecks)

	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set PingdomCheck instance as the owner and controller
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
	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *monitoringv1alpha1.PingdomCheck) *corev1.Pod {
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
