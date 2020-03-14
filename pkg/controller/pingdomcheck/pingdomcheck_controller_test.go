package pingdomcheck

import (
	"context"
    	"testing"
	"errors"
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
	name            	= "pingdom-operator"
	namespace       	= "pingdom"
	check_name      	= "unit-test"
 	check_name_update 	= "unit-test-update"
        check_url      	  	= "https://unit.test"
        check_id        	= 5123
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

func TestPingdomCheckControllerCreateCheckOk(t *testing.T) {

	// Mock pingdomClient
  	mockPingdomClient := new(MockPingdomClient)
	mockPingdomClient.On("CreateHttpPingdomCheck", check_name, check_url).Return(check_id, nil)
	// Get reconcile with mocked pingdom client and fake k8s client
       	r := getMockedReconciler(mockPingdomClient, getPingdomCheckCR())
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
        // Get CR to check if status was updated
	p := &monitoringv1alpha1.PingdomCheck{}
	err = r.client.Get(context.TODO(), req.NamespacedName, p)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}
	if res != (reconcile.Result{}) {
                t.Error("reconcile did not return an empty Result")
        }
 	assert.Equal(t, check_id, p.Status.ID, "should be update state with same ID")
}

func TestPingdomCheckControllerCreateCheckError(t *testing.T) {

        // Mock pingdomClient
        mockPingdomClient := new(MockPingdomClient)
        mockPingdomClient.On("CreateHttpPingdomCheck", check_name, check_url).Return(0, errors.New("Error callling pingdom API"))
        // Get reconcile with mocked pingdom client and fake k8s client
        r := getMockedReconciler(mockPingdomClient, getPingdomCheckCR())
        req := reconcile.Request{
                NamespacedName: types.NamespacedName{
                        Name:      name,
                        Namespace: namespace,
                },
        }
        _, err := r.Reconcile(req)
	assert.NotEqual(t, err, nil, "should be an error")
}


func TestPingdomCheckControllerUpdateCheckOk(t *testing.T) {

        // Mock pingdomClient
        mockPingdomClient := new(MockPingdomClient)
        mockPingdomClient.On("UpdateHttpPingdomCheck", check_id, check_name_update, check_url).Return(nil)
        // Get reconcile with mocked pingdom client and fake k8s client
        r := getMockedReconciler(mockPingdomClient, getExistingPingdomCheckCR())
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
	assert.Equal(t, res, reconcile.Result{}, "Nothing to reconcile")
}

func TestPingdomCheckControllerUpdateCheckError(t *testing.T) {

        // Mock pingdomClient
        mockPingdomClient := new(MockPingdomClient)
        mockPingdomClient.On("UpdateHttpPingdomCheck", check_id, check_name_update, check_url).Return(errors.New("Error callling pingdom API"))
        // Get reconcile with mocked pingdom client and fake k8s client
        r := getMockedReconciler(mockPingdomClient, getExistingPingdomCheckCR())
        req := reconcile.Request{
                NamespacedName: types.NamespacedName{
                        Name:      name,
                        Namespace: namespace,
                },
        }
        _, err := r.Reconcile(req)
	assert.NotEqual(t, err, nil, "should be an error")
}

//TODO add Delete tests

// Get controller to test with mocked dependecies
func getMockedReconciler(mockPingdomClient PingdomClient, pingodomcheckCR *monitoringv1alpha1.PingdomCheck) (*ReconcilePingdomCheck) {
        // Objects to track in the fake client.
        objs := []runtime.Object{
                pingodomcheckCR,
        }
        // Register operator types with the runtime scheme.
        s := scheme.Scheme
        s.AddKnownTypes(monitoringv1alpha1.SchemeGroupVersion, pingodomcheckCR)
        // Create a fake client to mock API calls.
        cl := fake.NewFakeClient(objs...)
        return &ReconcilePingdomCheck{client: cl, scheme: s, pingdomClient: mockPingdomClient}
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

// Create static CRD for unit testing
func getExistingPingdomCheckCR() *monitoringv1alpha1.PingdomCheck {
        return &monitoringv1alpha1.PingdomCheck{
                ObjectMeta: metav1.ObjectMeta{
                        Name:      name,
                        Namespace: namespace,
                },
                Spec: monitoringv1alpha1.PingdomCheckSpec{
                        Name: check_name_update,
                        URL: check_url,
                },
		Status: monitoringv1alpha1.PingdomCheckStatus{
			ID: check_id,
		},
        }

}

