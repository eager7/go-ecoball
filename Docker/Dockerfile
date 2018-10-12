#基于Ubuntu18.04
FROM ubuntu:18.04
MAINTAINER Docker ecoball <service@ecoball.org>

#安装git gcc openssl
RUN apt-get update
RUN apt-get install -y git gcc make wget golang-go python3

#设置环境变量
ENV GOPATH=/root/go
ENV PATH=$PATH:$GOPATH/bin

RUN mkdir -p $GOPATH/src/github.com/ecoball/ && cd $GOPATH/src/github.com/ecoball/
RUN git clone http://wanghengtao:caijing321@gitlab.quachain.net/ecoball/go-ecoball/
RUN cd ./go-ecoball/ && make && cd bootstrap/ && ./bootstrap.py -a


STOPSIGNAL SIGTERM
