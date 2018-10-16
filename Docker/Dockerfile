#基于Ubuntu18.04
FROM ubuntu:18.04
MAINTAINER Docker ecoball <service@ecoball.org>

#安装git gcc openssl
RUN apt-get update
RUN apt-get install -y git gcc make wget golang-go

#设置环境变量
ENV GOPATH /root/go
ENV PATH /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$GOPATH/bin

#下载代码打包生成执行文件
RUN git clone http://wanghengtao:caijing321@gitlab.quachain.net/ecoball/go-ecoball/  $GOPATH/src/github.com/ecoball/go-ecoball/
RUN cd $GOPATH/src/github.com/ecoball/go-ecoball/ && make

EXPOSE 80

STOPSIGNAL SIGTERM

CMD ["./build/ecoball", "run"]
