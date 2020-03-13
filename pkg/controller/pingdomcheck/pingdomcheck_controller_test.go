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
       	//"github.com/stretchr/testify/assert"
 	"github.com/go-logr/logr"
)

const (
	name            = "pingdom-operator"
	namespace       = "pingdom"
	check_name      = "unit-test"
        check_url      	= "https://unit.test"
        check_id        = 5123
)

type MockPingdomClient struct{
  mock.Mock
}

func (m *MockPingdomClient) CreateHttpPingdomCheck(reqLogger logr.Logger, name string, url string) (int, error) {
	args := m.Called(name, url)
	return args.Int(0), args.Error(1)
}

func (m *MockPingdomClient) UpdateHttpPingdomCheck(reqLogger logr.Logger, ID int, name string, url string) error {
 	args := m.Called(ID, name, url)
        return args.Error(0)
}

func (m *MockPingdomClient) DeleteHttpPingdomCheck(reqLogger logr.Logger, ID int) error {
	args := m.Called(ID)
        return args.Error(0)
}

func TestPingdomCheckControllerNewCheckOk(t *testing.T) {

	// create an instance of our test object
  	mockPingdomClient := new(MockPingdomClient)

  	// setup expectations
  	mockPingdomClient.On("CreateHttpPingdomCheck", mock.Anything, check_name, check_url).Return(check_id, nil)

	//Assertion should check status from CR agaisnt check_id
	//mockPingdomClient.AssertExpectations(t)
	//expect.Equal(t, mockThing, actual, "should return a Thing")



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
