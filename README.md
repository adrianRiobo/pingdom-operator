# pingdom-operator
k8s sample operator

# Overview

[best practices](https://blog.openshift.com/kubernetes-operators-best-practices/)  

# About [pingdom](https://my.pingdom.com/)

[api 3.1 announcement](https://royal.pingdom.com/announcing-the-pingdom-api-3-1/)
[api 2.1 with apikey](https://my.pingdom.com/app/account/appkeys) vs [api 3.1 with api-tokens](https://my.pingdom.com/app/api-tokens)

# Updating CRD

```
# Update go generated resoruces
operator-sdk generate k8s
# Generate CRD manifests
operator-sdk generate crds
```

# Build

```
# Build operator
operator-sdk build adrianriobo/pingdom-operator:0.1
# Login docker hub
docker login --username adrianriobo
# Push operator
docker push adrianriobo/pingdom-operator:0.1
```
# Testing 

Unit testing will be implemented using [testify framework](https://github.com/stretchr/testify)  

```
# Unit testing 
go test pkg/controller/pingdomcheck/*
```

# Deploy

Important due to used version within Pingdom: 2.1 APIKEY should be created [here](https://my.pingdom.com/app/account/appkeys)

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
kubectl create -f deploy/crds/crd.yaml  
# Deploy the app-operator
kubectl create -f deploy/operator.yaml
# Deploy pingdom check
kubectl create -f deploy/crds/pdc_google.yaml
kubectl create -f deploy/crds/pdc_sport.yaml
kubectl edit ...
kubectl delete ..
```
