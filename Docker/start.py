#!/usr/bin/env python3

import subprocess
import sys
import argparse
import json
import os

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
parser.add_argument('-i', '--node-ip', metavar='', help="IP address of node", nargs='+', dest="node_ip")
parser.add_argument('-o', '--host-ip', metavar='', help="IP address of host node", dest="host_ip")
parser.add_argument('-n', '--number', type=int, metavar='', help="The index number of container instance", dest="number")
parser.add_argument('-w', '--weight', type=int, metavar='', help="The number of weights", dest="weight")

#parse Arguments
args = parser.parse_args()

#Input parameter judgment
if args.node_ip is None or args.host_ip is None or args.number is None or args.weight is None:
    print('please input iP address of node and host node and the index number of container instance and weight number. -h shows options.')
    sys.exit(1)

if args.number < 0 or args.number > 4 * args.weight - 1:
    print('The index value must be between 0 and %d' %(4 * args.weight -1))
    sys.exit(1)

Pubkey = "1109ef616830cd7b8599ae7958fbee56d4c8168ffd5421a16025a398b8a4be"
start_pubkey = 40
start_port = 2000
committee = []
shard = []

ip_index = 0
for ip in args.node_ip:
    port_index = 0
    while port_index < 4 * args.weight:
        node = {
            "Pubkey": Pubkey + str(start_pubkey + 4 * args.weight * ip_index + port_index), 
            "Address": ip, 
            "Port": str(start_port + port_index)
        }
        port_index += 1
        if port_index <= args.weight:
            committee.append(node)
        else:
            shard.append(node)
    ip_index += 1


ip_index = args.node_ip.index(args.host_ip)


data = {
    "Pubkey": Pubkey + str(start_pubkey + 4 * args.weight * ip_index + args.number),
    "Address": "127.0.0.1",
    "Port": str(start_port + args.number),
    "Committee": committee,
    "Shard": shard
}

root_dir = os.path.split(os.path.realpath(__file__))[0]

with open(os.path.join(root_dir, '../build/sharding.json'), 'w') as f:
    json.dump(data, f)

run(os.path.join(root_dir, '../build/ecoball') + " run")
