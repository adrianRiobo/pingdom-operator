package pingdomcheck

import (
	"context"
	monitoringv1alpha1 "github.com/adrianRiobo/pingdom-operator/pkg/apis/monitoring/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
        //"github.com/russellcardullo/go-pingdom/pingdom"
        "os"
        //"net/url"
        "github.com/go-logr/logr"
)

const (
        pingdomCheckFinalizer = "finalizer.pingdomcheck"
        env_username = "PD_USERNAME"
        env_password = "PD_PASSWORD"
        env_apikey = "PD_APIKEY"
)

var log = logf.Log.WithName("controller_pingdomcheck")



// Add controller to Manager and Start it when the Manager is Started
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
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

        return nil
}


// Create the Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
        return &ReconcilePingdomCheck{client: mgr.GetClient(), scheme: mgr.GetScheme(), pingdomClient: createPingdomClient()}
}

// blank assignment to verify that ReconcilePingdomCheck implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcilePingdomCheck{}

// ReconcilePingdomCheck reconciles a PingdomCheck object
type ReconcilePingdomCheck struct {
	client client.Client
	scheme *runtime.Scheme
        pingdomClient PingdomClient
}

// Reconcile resources lifecycle defined by PingdomCheck CRD
func (r *ReconcilePingdomCheck) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling PingdomCheck")

	// Fetch the PingdomCheck instance
	instance := &monitoringv1alpha1.PingdomCheck{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
  			//Resource deleted after it can be managed
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
        // Manage PingdomCheck instance lifecycle
        if instance.Status.ID == 0 {
           	//New CRD create Http PingdomCheck
           	if err := r.createHttpPingdomCheck(reqLogger, instance); err != nil {
              		return reconcile.Result{}, err
   		}
        } else {
        	if instance.GetDeletionTimestamp() != nil {
           		//CRD mark for deletion check finalizer and execute to allow gc
           		if err := r.deleteHttpPingdomCheck(reqLogger, instance); err != nil {
              			return reconcile.Result{}, err
           		}
        	} else {
      			//CRD update: check external resource, compare and act if required
                        if err := r.updateHttpPingdomCheck(reqLogger, instance); err != nil {
                                return reconcile.Result{}, err
                        }
                }
        }
	return reconcile.Result{}, nil
}



// Delete http pingdom check 
func (r *ReconcilePingdomCheck) deleteHttpPingdomCheck(reqLogger logr.Logger, p *monitoringv1alpha1.PingdomCheck) error {
        if contains(p.GetFinalizers(), pingdomCheckFinalizer) {
           //Execute finalizer
           if err := r.finalizePingdomCheck(reqLogger, p); err != nil {
              return err
           }
           //Remove finalizers
           p.SetFinalizers(remove(p.GetFinalizers(), pingdomCheckFinalizer))
           err := r.client.Update(context.TODO(), p)
           if err != nil {
              return err
           }
        }
        return nil
}

func (r *ReconcilePingdomCheck) finalizePingdomCheck(reqLogger logr.Logger, p *monitoringv1alpha1.PingdomCheck) error {
	err := r.pingdomClient.DeleteHttpPingdomCheck(reqLogger, p.Status.ID)
	if err != nil {
                reqLogger.Error(err, "Error deleting Pingdomcheck")
        } else {
                // Should check if response gives a 200 if not create override err
                reqLogger.Info("PingdomCheck remove sucessfully", "PingdomCheck ID", p.Status.ID)
        }
        return err


}

// Update pingdomcheck
func (r *ReconcilePingdomCheck) updateHttpPingdomCheck(reqLogger logr.Logger, p *monitoringv1alpha1.PingdomCheck) error {
	err := r.pingdomClient.UpdateHttpPingdomCheck(reqLogger, p.Status.ID, p.Spec.Name, p.Spec.URL)
	if err != nil {
                log.Error(err, "Error updating check", "with ID", p.Status.ID)
        }
        return err
}

// Create http pingdom check 
func (r *ReconcilePingdomCheck) createHttpPingdomCheck(reqLogger logr.Logger, p *monitoringv1alpha1.PingdomCheck) error {
	ID, err := r.pingdomClient.CreateHttpPingdomCheck(reqLogger, p.Spec.Name, p.Spec.URL)
        if err != nil {
           log.Error(err, "Error creating check")
        } else {
           p.Status.ID = ID
           err := r.client.Status().Update(context.TODO(), p)
           if err != nil {
              reqLogger.Error(err, "Failed to update PingdomCheck status.")
              return err
           }
        }
        // Add finalizer for this CR
        if !contains(p.GetFinalizers(), pingdomCheckFinalizer) {
           if err := r.addFinalizer(reqLogger, p); err != nil {
              return err
           }
        }
        log.Info("Created http check:", "with ID", p.Status.ID)
        return nil
}

// Create pingdom client to manage checks within pingdom api 
func createPingdomClient() PingdomClient {
  	user := getEnv(env_username)
  	password := getEnv(env_password)
  	apikey := getEnv(env_apikey)
        pingdomClient, err := NewRCPingdomClient(user, password, apikey)
        if err != nil {
    		//Stop application as client is required.
                log.Error(err, "")
                os.Exit(1)
  	} 
  	return pingdomClient
}

//Finalizers
func (r *ReconcilePingdomCheck) addFinalizer(reqLogger logr.Logger, p *monitoringv1alpha1.PingdomCheck) error {
	reqLogger.Info("Adding Finalizer for the PingdomCheck")
	p.SetFinalizers(append(p.GetFinalizers(), pingdomCheckFinalizer))

	// Update CR
	err := r.client.Update(context.TODO(), p)
	if err != nil {
		reqLogger.Error(err, "Failed to update PingdomCheck with finalizer")
		return err
	}
	return nil
}


//Helper functions

func getEnv(env string) string {
  	value := os.Getenv(env)
  	if value == "" {
     		log.Info("Error getting ENV value", "ENV requierd:", env)
  	} else {
     		log.Info("Getting user for pingdom", "user", value)
  	}
  	return value
}


func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func remove(list []string, s string) []string {
	for i, v := range list {
		if v == s {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list
}
