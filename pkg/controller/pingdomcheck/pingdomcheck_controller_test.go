package pingdomcheck

import (
    "context"
    "testing"

    monitoringv1alpha1 "github.com/adrianriobo/pingdom-operator/pkg/apis/monitoring/v1alpha1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/client-go/kubernetes/scheme"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/client/fake" 
)

var (
	name            = "pingdom-operator"
	namespace       = "pingdom"
	check_name      = "unit-test"
        check_url      	= "https://unit.test"
)


func TestMemcachedController(t *testing.T) {
	pingdomcheck := getPingdomCheckCR()

       // Objects to track in the fake client.
	objs := []runtime.Object{
		pingdomcheck,
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(monitoringv1alpha1.SchemeGroupVersion, pingdomcheck)
	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileMemcached{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
}

// Create static CRD for unit testing
func getPingdomCheckCR() *monitoringv1alpha1.PingdomCheck {
	return &monitoringv1alpha1.PingdomCheck{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: cachev1alpha1.MemcachedSpec{
			Name: check_name, 
			URL: check_url,
		},
	}

}
