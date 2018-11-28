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
Once the configuration file shard_setup.toml is configured, execute key_generation.py to generate public, private keys, http port and onlooker port for the startup
```
./key_generate.py
```

## run.py

start all node, first shard node, and then committee node

```
cd $GOPATH/src/github.com/ecoball/go-ecoball/launch
./run.py -i ${HOST_IP}  [-b] [-w]
```
Log generation for each node is under ecoball/committee/ecoball_*/log 


### stop.py
You can stop all node

Stop all ecoball
```
./stop.py
```