Ecoball-Docker
========

#Depends

You need install docker and docker-compose

#Run

##docker_service.sh start | stop

If your operating system is ubuntu
```
cd $GOPATH/src/github.com/ecoball/go-ecoball/Docker
./docker_service.sh start
```
You will start 21 ecoball and ecowallet and eballscan

If you want to stop all services
```
cd $GOPATH/src/github.com/ecoball/go-ecoball/Docker
./docker_service.sh stop
```

##docker-compose

Create the map directory
```
mkdir $GOPATH/src/github.com/ecoball/go-ecoball/Docker/ecoball_log
```
start services
```
cd $GOPATH/src/github.com/ecoball/go-ecoball/Docker
sudo docker-compose up -d
```
you will start ecoball and ecowallet and eballscan
