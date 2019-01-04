#Base on Ubuntu18.04
FROM ubuntu:18.04
MAINTAINER Docker ecoball <service@ecoball.org>

#Install dependencies
RUN apt-get update
RUN apt-get install -y python3 python3-pip vim
RUN apt-get install psmisc

#Setting environment variables
ENV GOPATH /root/go
ENV PATH /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$GOPATH/bin

#install pytoml
RUN pip3 install pytoml

#install cockroachdb ecoball ecowallet eballscan
#ecoball
COPY ecoball /ecoball/ecoball/
COPY ecoball.toml /ecoball/ecoball/
COPY start.py /ecoball/ecoball/

#ipfs
# COPY swarm.key /tmp/storage/

#ecowallet
COPY ecowallet /ecoball/ecowallet/

#eballscan
# COPY cockroach /usr/local/bin
# COPY eballscan /ecoball/eballscan/build/
# COPY eballscan_service.sh /ecoball/eballscan/

EXPOSE 20678
EXPOSE 20679
EXPOSE 20680

#eballscan
EXPOSE 9001

#cockroachdb
EXPOSE 8081
EXPOSE 26257

STOPSIGNAL SIGTERM

CMD ["/ecoball/ecoball/ecoball", "run"]
