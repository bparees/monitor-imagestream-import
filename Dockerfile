FROM openshift/origin-base

ADD _output/local/bin/linux/amd64/monitor /usr/bin/monitor

ENTRYPOINT [ "/usr/bin/monitor" ]
