#Base on Ubuntu18.04
FROM ubuntu:18.04
MAINTAINER Docker ecoball <service@ecoball.org>

#Install dependencies
RUN apt-get update
RUN apt-get install -y python3 python3-pip

#Setting environment variables
ENV GOPATH /root/go
ENV PATH /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$GOPATH/bin

#install pytoml
RUN pip3 install pytoml

#install cockroachdb ecoball ecowallet eballscan
COPY cockroach /usr/local/bin
COPY ecoball /ecoball/ecoball/
COPY ecoball.toml /ecoball/ecoball/
COPY start.py /ecoball/ecoball/
COPY ecowallet /ecoball/ecowallet/
COPY eballscan /ecoball/eballscan/
COPY eballscan_service.sh /ecoball/eballscan/
COPY swarm.key /tmp/storage
EXPOSE 20678
EXPOSE 20679
EXPOSE 20680
EXPOSE 9001
EXPOSE 8081
EXPOSE 26257

STOPSIGNAL SIGTERM

CMD ["/ecoball/ecoball/ecoball", "run"]
