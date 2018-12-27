#!/usr/bin/env python3

import subprocess
import sys
import argparse
import json
import os
import pytoml
import signal
from time import sleep

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


def main():
    # Command Line Arguments
    parser = argparse.ArgumentParser()
    parser.add_argument('-o', '--host-ip', metavar='', required=True, help="IP address of host node", dest="host_ip")
    parser.add_argument('-n', '--number', type=int, required=True, metavar='', help="The index number of container instance", dest="number")
    parser.add_argument('-a', '--all-config', metavar='', required=True, help="All configuration information", dest="all_config")
    parser.add_argument('-s', '--size', type=int, default=5, help="Number of nodes per shard")
    parser.add_argument('-c', '--config', metavar='', help="Different configuration items for ecoball.toml")

    #parse Arguments
    args = parser.parse_args()

    #Generate the configuration json files required for sharding
    all_config =json.loads(args.all_config)
    network = all_config["network"]
    node_ip = []
    for ip in network:
        node_ip.append(ip)

    start_port = 9901
    committee = []
    shard = []
    candidate = []
    list_count = []

    for ip in node_ip:
        port_index = 0
        committee_count = network[ip][0]
        shard_count = network[ip][1]
        if len(network[ip]) > 2:
            candidate_count = network[ip][2]
        while port_index < committee_count + shard_count + candidate_count:
            node_index = ip + "_" + str(port_index)
            node = {
                "Pubkey": all_config[node_index]["p2p_peer_publickey"], 
                "Address": ip, 
                "Port": str(start_port + port_index)
            }
            port_index += 1
            if port_index <= committee_count:
                committee.append(node)
            elif port_index > committee_count + shard_count:
                candidate.append(node)
            else:
                shard.append(node)
        list_count.append(port_index)

    ip_index = node_ip.index(args.host_ip)
    i = 0
    key_base = 0
    while i < ip_index:
        key_base += list_count[i]
        i += 1

    node_index = args.host_ip + "_" + str(args.number)
    data = {
        "size": str(args.size),
        "Pubkey": all_config[node_index]["p2p_peer_publickey"],
        "Address": args.host_ip,
        "Port": str(start_port + args.number),
        "Committee": committee,
        "Shard": shard,
        "Candidate": candidate
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


def signal_handler(a, b):
    print('receive SIGTERM')
    run("killall ecoball")
    print('sleep 3s for collect runtime information')
    sleep(3)
    sys.exit(0)


if __name__ == "__main__":
    signal.signal(signal.SIGTERM, signal_handler)
    signal.signal(signal.SIGINT, signal_handler)
    main()
    while 1:
        sleep(10)