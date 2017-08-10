FROM busybox:latest

ADD .build/linux-amd64/kubernetes-usage-log /bin/kubernetes-usage-log

ENTRYPOINT ["/bin/kubernetes-usage-log"]
