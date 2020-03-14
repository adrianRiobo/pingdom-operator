package pingdomcheck

import (
	"context"
    	"testing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    	monitoringv1alpha1 "github.com/adrianRiobo/pingdom-operator/pkg/apis/monitoring/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
    	"k8s.io/client-go/kubernetes/scheme" 
 	"k8s.io/apimachinery/pkg/types"
    	"sigs.k8s.io/controller-runtime/pkg/client/fake" 
     	"sigs.k8s.io/controller-runtime/pkg/reconcile" 
        "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/assert"
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
	mockPingdomClient.On("CreateHttpPingdomCheck", check_name, check_url).Return(check_id, nil)

       // Objects to track in the fake client.
	objs := []runtime.Object{
		getPingdomCheckCR(),
	}
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(monitoringv1alpha1.SchemeGroupVersion, getPingdomCheckCR())
	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	r := &ReconcilePingdomCheck{client: cl, scheme: s, pingdomClient: mockPingdomClient}
	// Reconcile()
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	//res
	_, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
        // Get CR to check its status
	p := &monitoringv1alpha1.PingdomCheck{}
	err = cl.Get(context.TODO(), req.NamespacedName, p)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}
 	assert.Equal(t, check_id, p.Status.ID, "should be update state with same ID")

	
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
