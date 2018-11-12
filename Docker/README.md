Ecoball-Docker
========

# Depends

You need install docker and docker-compose

# Run shard

### docker_build.sh
You need to use docker_build.sh first to create the image
```
./docker_build.sh
```

### shard_setup.toml
Before starting shard mode, you need to configure shard start profile shard_setup.toml
```
#Configuration file for shard network startup

#Network host IP address list and the number of Committee and Shard on each physical machine
#The key string represents the host IP address 
#The first value represents the number of Committee nodes
#The second value represents the number of Shard nodes
[network]
"192.168.8.58" = [0, 5]
"192.168.8.60" = [0, 5]
"192.168.8.62" = [5, 0]


#Different configuration items for ecoball.toml
#The name from the host IP address plus a sequence number
#for example, 127.0.0.1_0 represents the first docker container on the 127.0.0.1
["192.168.8.58_0"]
output_to_terminal = true


["192.168.8.60_0"]
output_to_terminal = true
```

## share_shard.py
Start shard node first when sharding starts
```
cd $GOPATH/src/github.com/ecoball/go-ecoball/Docker
./share_shard.py 
```
Log generation for each node is under ./ecoball_log/shard/$DOCKERNAME/ 

## share_committee.py

Start committee node second when sharding starts
```
cd $GOPATH/src/github.com/ecoball/go-ecoball/Docker
./share_committee.py 
```
Log generation for each node is under ./ecoball_log/committee/$DOCKERNAME/ 

