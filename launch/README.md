Ecoball
========

# Depends

You need install python3 and pip3
```
sudo apt-get install -y python3 python3-pip
```

You need to install python's pytoml module
```
pip3 install pytoml
```

After doing the above, executing the docker command does not require a sudo referral
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
size = 2

# Network host IP address list and the number of Committee and Shard on each physical machine
# The key string represents the host IP address 
# The first value represents the number of Committee nodes
# The second value represents the number of Shard nodes
[network]
"192.168.9.43" = [2, 4]

```
## key_generate.py
Once the configuration file shard_setup.toml is configured, execute key_generation.py to generate public and private keys for the startup
```
./key_generate.py
```

## run.py

After starting the shard container, execute the share_commitment.py script to start the committee node.

If the -b option is added, the eballscan container is started

If the -w option is added, the ecowallet container is started
```
cd $GOPATH/src/github.com/ecoball/go-ecoball/Docker
./share_committee.py -i ${HOST_IP}  [-b] [-w]
```
Log generation for each node is under ./ecoball_log/committee/$DOCKERNAME/ 

The wallet file is generated under ./wallet

### stop.py
You can stop and remove all docker containers with docker_service.sh before creating a new image.

Stop all ecoball
```
./stop.py
```