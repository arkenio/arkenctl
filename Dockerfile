FROM       arken/gom-base
MAINTAINER Damien Metzler <dmetzler@nuxeo.com>

RUN go get github.com/arkenio/arkenctl
WORKDIR /usr/local/go/src/github.com/arkenio/arkenctl
#RUN git checkout v0.3.0
RUN gom install
RUN gom test

ENTRYPOINT ["arkenctl", "--etcdAddress", "http://172.17.42.1:4001", "--logtostderr"]
