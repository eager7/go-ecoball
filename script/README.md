Ecoball-Docker
========

# Depends

You need install docker and python3 and pip3
```
sudo apt-get install -y docker.io python3 python3-pip
```

You need to install python's pytoml module
```
pip3 install pytoml
```

You need to configure the docker environment
Docker groups already exist by default, and they need to be created manually if they do not exist
```
sudo groupadd docker
```

Add the current user to the docker group
```
sudo gpasswd -a ${USER} docker
```

Refresh docker group members
```
newgrp docker
```
After doing the above, executing the docker command does not require a sudo referral

# Run shard

### pull the mirror
Before executing the startup script, you need to get the latest image

### shard_setup.toml
Before starting shard mode, you need to configure shard start profile shard_setup.toml
```
# Configuration file for shard network startup

# Number of nodes per shard
size = 5

# Network host IP address list and the number of Committee and Shard on each physical machine
# The key string represents the host IP address 
# The first value represents the number of Committee nodes
# The second value represents the number of Shard nodes
[network]
"192.168.8.58" = [0, 5]
"192.168.8.60" = [0, 5]
"192.168.8.62" = [5, 0]


# Different configuration items for ecoball.toml
# The name from the host IP address plus a sequence number
# for example, 127.0.0.1_0 represents the first docker container on the 127.0.0.1
["192.168.8.58_0"]
output_to_terminal = true


["192.168.8.60_0"]
output_to_terminal = true
```
## key_generate.py
Once the configuration file shard_setup.toml is configured, execute key_generation.py to generate public and private keys for the startup container
```
./key_generate.py
```

## share_shard.py 
To start the sharding network, execute the share_shard.py script to start the shard container.

If the -b option is added, the eballscan container is started

If the -w option is added, the ecowallet container is started

```
cd $GOPATH/src/github.com/ecoball/go-ecoball/Docker
./share_shard.py -i ${HOST_IP} [-b] [-w]
```
Log generation for each node is under ./ecoball_log/shard/$DOCKERNAME/ 

The wallet file is generated under ./wallet

## share_committee.py

After starting the shard container, execute the share_commitment.py script to start the committee node.

If the -b option is added, the eballscan container is started

If the -w option is added, the ecowallet container is started
```
cd $GOPATH/src/github.com/ecoball/go-ecoball/Docker
./share_committee.py -i ${HOST_IP}  [-b] [-w]
```
Log generation for each node is under ./ecoball_log/committee/$DOCKERNAME/ 

The wallet file is generated under ./wallet

### docker_service.sh
You can stop and remove all docker containers with docker_service.sh before creating a new image.

Stop all running containers
```
./docker_service.sh stop
```

Delete containers that have been stopped
```
./docker_service.sh remove
```
Restart all stopped containers
```
./docker_service.sh start
```

Enter into docker container
```
docker exec -it ID /bin/bash
```