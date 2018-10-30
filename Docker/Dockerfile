#Base on Ubuntu18.04
FROM ubuntu:18.04
MAINTAINER Docker ecoball <service@ecoball.org>

#Install dependencies
RUN apt-get update
RUN apt-get install -y git gcc make wget golang-go python3 psmisc sudo

#Setting environment variables
ENV GOPATH /root/go
ENV PATH /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$GOPATH/bin

#Download ecoball source code and build
RUN git clone http://wanghengtao:caijing321@gitlab.quachain.net/ecoball/go-ecoball/  $GOPATH/src/github.com/ecoball/go-ecoball/
RUN cd $GOPATH/src/github.com/ecoball/go-ecoball/ && make

#Download eballscan source code and build
RUN git clone https://github.com/ecoball/eballscan  $GOPATH/src/github.com/ecoball/eballscan
RUN cd $GOPATH/src/github.com/ecoball/eballscan && ./eballscan_build.sh

EXPOSE 20678
EXPOSE 20679
EXPOSE 20680
EXPOSE 9001
EXPOSE 8081
EXPOSE 26257

STOPSIGNAL SIGTERM

CMD ["/root/go/src/github.com/ecoball/go-ecoball/build/ecoball", "run"]
