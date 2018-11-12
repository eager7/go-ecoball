#!/usr/bin/env python3

import subprocess
import sys
import argparse
import json
import os
import pytoml

# Sharding scheme: initial startup of 5 committee, 3 Shared, each Shared 5 nodes.
# Buy five servers, one server one committee docker instance and three share docker instance


def run(shell_command):
    '''
    Execute shell command.
    If it fails, exit the program with an exit code of 1.
    '''

    print('start.py:', shell_command)
    if subprocess.call(shell_command, shell=True):
        print('start.py: exiting because of error')
        sys.exit(1)


# Command Line Arguments
parser = argparse.ArgumentParser()
parser.add_argument('-o', '--host-ip', metavar='', required=True, help="IP address of host node", dest="host_ip")
parser.add_argument('-n', '--number', type=int, required=True, metavar='', help="The index number of container instance", dest="number")
parser.add_argument('-e', '--network', metavar='', required=True, help="Network host IP address list and the number of Committee and Shard on each physical machine")
parser.add_argument('-c', '--config', metavar='', help="Different configuration items for ecoball.toml")

#parse Arguments
args = parser.parse_args()

#Generate the configuration json files required for sharding
network =json.loads(args.network)
node_ip = []
for ip in network:
    node_ip.append(ip)

Pubkey = "1109ef616830cd7b8599ae7958fbee56d4c8168ffd5421a16025a398b8a4be"
start_pubkey = 40
start_port = 2000
committee = []
shard = []
list_count = []

container_count = 0
for ip in node_ip:
    port_index = 0
    committee_count = network[ip][0]
    shard_count = network[ip][1]
    while port_index < committee_count + shard_count:
        node = {
            "Pubkey": Pubkey + str(start_pubkey + container_count + port_index), 
            "Address": ip, 
            "Port": str(start_port + port_index)
        }
        port_index += 1
        if port_index <= committee_count:
            committee.append(node)
        else:
            shard.append(node)
    container_count += port_index
    list_count.append(port_index)


ip_index = node_ip.index(args.host_ip)
i = 0
key_base = 0
while i < ip_index:
    key_base += list_count[i]


data = {
    "Pubkey": Pubkey + str(start_pubkey + key_base + args.number),
    "Address": args.host_ip,
    "Port": str(start_port + args.number),
    "Committee": committee,
    "Shard": shard
}

root_dir = os.path.split(os.path.realpath(__file__))[0]

with open(os.path.join(root_dir, 'sharding.json'), 'w') as json_file:
    json.dump(data, json_file)

#Generate the configuration toml files required for ecoball
ecoball_config = {}
with open(os.path.join(root_dir, 'ecoball.toml')) as ecoball_file:
    ecoball_config = pytoml.load(ecoball_file)

if args.config is not None:
    with open(os.path.join(root_dir, 'ecoball.toml'), 'w') as ecoball_file:
        config =json.loads(args.config)
        for one in config:
            ecoball_config[one] = config[one]
        pytoml.dump(ecoball_config, ecoball_file)

#start ecoball
run("cd " + os.path.join(root_dir) + "&& ./ecoball run")

