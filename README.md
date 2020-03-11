# pingdom-operator
k8s sample operator

# Build

```
# Build operator
operator-sdk build adrianriobo/pingdom-operator:0.1
# Login docker hub
docker login --username adrianriobo
# Push operator
docker push adrianriobo/pingdom-operator:0.1
```

# Deploy

```
# Create pingdom credentials
deploy/secrets/create_secret.sh username password apikey
kubectl create -f pingdomsecret.yaml
rm pingdomsecret.yaml
# Setup Service Account
kubectl create -f deploy/service_account.yaml  
# Setup RBAC  
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
# Setup the CRD
kubectl create -f deploy/crds/monitoring.adrianriobo.com_pingdomchecks_crd.yaml
# Deploy the app-operator
kubectl create -f deploy/operator.yaml
# Deploy pingdom check
kubectl create -f deploy/crds/monitoring.adrianriobo.com_v1alpha1_pingdomcheck_cr.yaml
```
