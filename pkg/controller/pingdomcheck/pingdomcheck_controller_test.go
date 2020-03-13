package pingdomcheck

import (
    	//"context"
    	"testing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    	monitoringv1alpha1 "github.com/adrianRiobo/pingdom-operator/pkg/apis/monitoring/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
    	"k8s.io/client-go/kubernetes/scheme" 
 	"k8s.io/apimachinery/pkg/types"
    	//"sigs.k8s.io/controller-runtime/pkg/client"
    	"sigs.k8s.io/controller-runtime/pkg/client/fake" 
     	"sigs.k8s.io/controller-runtime/pkg/reconcile"
        "github.com/russellcardullo/go-pingdom/pingdom" 
        "github.com/stretchr/testify/mock"
)

var (
	name            = "pingdom-operator"
	namespace       = "pingdom"
	check_name      = "unit-test"
        check_url      	= "https://unit.test"
)

// Mocked service
type MockedCheckService struct{
  client *Client
  mock.Mock
}

func (mcs *MockedCheckService) Create(check Check) (*CheckResponse, error) { 
	args := m.Called(check)
	return nil, args.Error(1)
}

func TestPingdomCheckController(t *testing.T) {
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
	// Create a PingdomCheck object with the scheme and fake client.
        pingdomClient, err := pingdom.NewClientWithConfig(pingdom.ClientConfig{
                User:     "user",
                Password: "password",
                APIKey:   "apikey",
        })

  	// create an instance of our test object
  	testObj := new(MockedCheckService)

  	// setup expectations
  	testObj.On("Create", mock.Anything).Return(nil, nil)

        pingdomClient.Checks = testObj
	
	r := &ReconcilePingdomCheck{client: cl, scheme: s, pingdomClient: pingdomClient}

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
		Spec: monitoringv1alpha1.PingdomCheckSpec{
			Name: check_name, 
			URL: check_url,
		},
	}

}
