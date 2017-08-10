FROM busybox:latest

ADD kubernetes-usage-log /bin/kubernetes-usage-log

ENTRYPOINT ["/bin/kubernetes-usage-log"]
