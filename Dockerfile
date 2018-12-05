FROM  quay.io/prometheus/busybox:latest
LABEL maintainer "Stackdriver Engineering <engineering@stackdriver.com>"

COPY opencensus-operator /bin/opencensus-operator

USER       nobody
EXPOSE     9091
ENTRYPOINT [ "/bin/opencensus-operator" ]

