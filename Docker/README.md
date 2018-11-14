Ecoball-Docker
========

# Depends

You need install docker and python3 and pip3

# Run shard

### ecoball.toml
The ecoball.toml profile will be mirrored. Please configure the configuration items before mirroring.

Add a new configuration item to the project's ecoball.toml, and be sure to copy the latest code-generated file ecoball.toml to the Docker directory.

If the configuration items for a container require special customization, do the configuration in the shard_setup.toml file(Refer to the shard_setup.toml configuration file for details).

### docker_build.sh
You need to use docker_build.sh first to create the image

This script will call the Makefile of go-ecoball to generate the latest executable file ecoball, ecowallet, call the Makefile of eballscan to generate the executable file eballscan, and copy it under the Docker directory to do the image(Refer to the script header for details).

```
./docker_build.sh
```

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

```
cd $GOPATH/src/github.com/ecoball/go-ecoball/Docker
./share_shard.py 
```
Log generation for each node is under ./ecoball_log/shard/$DOCKERNAME/ 

## share_committee.py

After starting the shard container, execute the share_commitment.py script to start the committee node.

If the -b option is added, the eballscan container is started

If the -w option is added, the ecowallet container is started
```
cd $GOPATH/src/github.com/ecoball/go-ecoball/Docker
./share_committee.py [-d]
```
Log generation for each node is under ./ecoball_log/committee/$DOCKERNAME/ 

### docker_service.sh
You can stop all docker containers with docker_service.sh before creating a new image.
```
./docker_service.sh stop
```

Enter into docker container
```
docker exec -it ID /bin/bash
```