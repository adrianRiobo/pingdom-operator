package pingdomcheck

import (
	"context"

	monitoringv1alpha1 "github.com/adrianRiobo/pingdom-operator/pkg/apis/monitoring/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
        "github.com/russellcardullo/go-pingdom/pingdom"
        "os"
        "net/url"
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
        /*
        pingdomChecks, err := pingdomClient.Checks.List()
        if err != nil {
          log.Error(err, "Error listing checks")
        }
        log.Info("All checks intial:", "all checks", pingdomChecks)
        */
	return &ReconcilePingdomCheck{client: mgr.GetClient(), scheme: mgr.GetScheme(), pingdomClient: pingdomClient}
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
        pingdomClient *pingdom.Client
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
        parsedUrl, err := url.Parse(instance.Spec.URL)
        //Create the http check
        newCheck := pingdom.HttpCheck{Name: instance.Spec.Name, Hostname: parsedUrl.Host, Resolution: 5}
	check, err := r.pingdomClient.Checks.Create(&newCheck)
        if err != nil {
          log.Error(err, "Error creating check")
        } else {
          instance.Status.ID = check.ID
          err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update Memcached status.")
			return reconcile.Result{}, err
		}
          log.Info("Created http check:", "with ID", instance.Status.ID)
        }
        //instance.Status.ID = check.ID
        // Update state with check.id
        
        /*
        for i, configMap := range configMapList.Items {
                reqLogger.Info("Getting Configmap", "Configmap,Name", i, "BAD", configMap.Name)
	}
        */

        pingdomChecks, err := r.pingdomClient.Checks.List()
        if err != nil {
          log.Error(err, "Error listing checks")
        }
        log.Info("All checks intial:", "all checks", pingdomChecks)
        
        /*
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
        */
	return reconcile.Result{}, nil
}

