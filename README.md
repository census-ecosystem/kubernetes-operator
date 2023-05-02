> **Warning**
>
> OpenCensus and OpenTracing have merged to form [OpenTelemetry](https://opentelemetry.io), which serves as the next major version of OpenCensus and OpenTracing.
>
> OpenTelemetry has now reached feature parity with OpenCensus, with tracing and metrics SDKs available in .NET, Golang, Java, NodeJS, and Python. **All OpenCensus Github repositories, except [census-instrumentation/opencensus-python](https://github.com/census-instrumentation/opencensus-python), will be archived on July 31st, 2023**. We encourage users to migrate to OpenTelemetry by this date.
>
> To help you gradually migrate your instrumentation to OpenTelemetry, bridges are available in Java, Go, Python, and JS. [**Read the full blog post to learn more**](https://opentelemetry.io/blog/2023/sunsetting-opencensus/).

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
namespace), run:

```bash
CLUSTER_NAME=target_cluster deploy/deploy.sh
```

(Optional) You can deploy the webhook in a specific NAMESPACE and using a specific IMAGE:
```bash
CLUSTER_NAME=target_cluster NAMESPACE=target_namespace IMAGE=image_name deploy/deploy.sh
```

## Testing (optional)
To deploy three test pods, each respectively opting in, out, and for the operator's default
configuration mode, run:

```bash
kubectl apply -f deploy/test-pods.yaml
```
