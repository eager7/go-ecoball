#Base on Ubuntu18.04
FROM ubuntu:18.04
MAINTAINER Docker ecoball <service@ecoball.org>

#Install dependencies
RUN apt-get update
RUN apt-get install -y git gcc make wget golang-go python3 psmisc sudo vim python3-pip

#Setting environment variables
ENV GOPATH /root/go
ENV PATH /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$GOPATH/bin

#install pytoml
RUN pip3 install pytoml

#install cockroachdb
RUN wget -qO- https://binaries.cockroachdb.com/cockroach-v2.0.6.linux-amd64.tgz | tar  xvz
RUN cp -i cockroach-v2.0.6.linux-amd64/cockroach /usr/local/bin
RUN rm -fr "./cockroach-v2.0.6.linux-amd64"

#create directory
RUN mkdir -p /ecoball/ecoball/ && mkdir -p /ecoball/ecowallet/ && mkdir /ecoball/eballscan/
COPY ../build/ecoball /ecoball/ecoball/
COPY start.py /ecoball/ecoball/
COPY ../build/ecowallet /ecoball/ecowallet/
COPY ../../eballscan/build/eballscan /ecoball/eballscan/
COPY ../../eballscan/build/eballscan_service.sh /ecoball/eballscan/

EXPOSE 20678
EXPOSE 20679
EXPOSE 20680
EXPOSE 9001
EXPOSE 8081
EXPOSE 26257

STOPSIGNAL SIGTERM

CMD ["/ecoball/ecoball/ecoball", "run"]
