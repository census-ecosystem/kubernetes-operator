# OpenCensus Kubernetes Operator

This operator provides automated configuration of Kubernetes container resource for OpenCensus
through an admission webhook.

It is required to run the webhook before any application that uses OpenCensus library.

## Before you begin
Make sure that you have:

  * Installed golang (Recommended version 1.11.2)
  * Installed Docker
  * Installed kubectl
  * (Optional) Installed gcloud (if running on GKE)

## Build locally (optional)

1. Install make, see more instruction on how to install make [here](https://www.gnu.org/software/make/).

2. Build container
```bash
make container
```

## Deploy
To deploy the standard webhook into a cluster (by default this will be deployed in the default 
namespace), 
run:

```bash
deploy/deploy.sh
```

(Optional) You can deploy the webhook in a specific NAMESPACE and using a specific IMAGE:
```bash
NAMESPACE=target_namespace IMAGE=image_name deploy/deploy.sh
```

## Testing (optional)
To deploy three test pods, each respectively opting in, out, and for the operator's default
configuration mode, run:

```bash
kubectl apply -f deploy/test-pods.yaml
```
