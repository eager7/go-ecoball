Ecoball-Docker
========

# Depends

You need install docker and docker-compose

# Run

## docker_service.sh start | stop

If your operating system is ubuntu
```
cd $GOPATH/src/github.com/ecoball/go-ecoball/Docker
./docker_service.sh start
```
You will start twenty-one ecoball and one ecowallet and one eballscan

If you want to stop all services
```
cd $GOPATH/src/github.com/ecoball/go-ecoball/Docker
./docker_service.sh stop
```

## docker-compose

Create the map directory
```
mkdir $GOPATH/src/github.com/ecoball/go-ecoball/Docker/ecoball_log
```
start services
```
cd $GOPATH/src/github.com/ecoball/go-ecoball/Docker
sudo docker-compose up -d
```
you will start one ecoball and one ecowallet and one eballscan


## share_shard.py
Start shard node first when sharding starts
```
cd $GOPATH/src/github.com/ecoball/go-ecoball/Docker
./share_shard.py -i $IPOFALLNODES -o $HOSTIP -w $WEIGHT
```
It will start 3 * $WEIGHT ecoball shard node.

Log generation for each node is under ./ecoball_log/shard/$DOCKERNAME/ directory

## share_committee.py

Start committee node second when sharding starts
```
cd $GOPATH/src/github.com/ecoball/go-ecoball/Docker
./share_committee.py -i $IPOFALLNODES -o $HOSTIP -w $WEIGHT
```
It will start 3 * $WEIGHT ecoball committee node.

Log generation for each node is under ./ecoball_log/committee/$DOCKERNAME/ directory

