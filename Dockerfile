FROM  gcr.io/distroless/static:latest
LABEL maintainer "Stackdriver Engineering <engineering@stackdriver.com>"

COPY opencensus-operator /bin/opencensus-operator

EXPOSE     9091
ENTRYPOINT [ "/bin/opencensus-operator" ]
