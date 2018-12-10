# OpenCensus Kubernetes Operator

This operator provides automated configuration of Kubernetes pods for OpenCensus
through an admission webhook.

## Usage

To deploy the webhook into a cluster, run:

```
NAMESPACE=<target_namespace> IMAGE=<image_name> deploy/deploy.sh
```

To deploy three test pods, each respectively opting in, out, and for the operator's default configuration mode, run:

```
kubeactl apply -f deploy/test-pods.yaml
```
