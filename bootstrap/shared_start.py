#!/usr/bin/env python3

import subprocess
import sys
import time
import argparse
import json

# Sharding scheme: initial startup of 5 committee, 3 Shared, each Shared 5 nodes.


def run(shell_command):
    '''
    Execute shell command.
    If it fails, exit the program with an exit code of 1.
    '''

    print('shared_start.py:', shell_command)
    if subprocess.call(shell_command, shell=True):
        print('bootstrap.py: exiting because of error')
        sys.exit(1)


def background(shell_command):
    '''
    Run script commands in the background.
    '''

    print('shared_start.py:', shell_command)
    return subprocess.Popen(shell_command, shell=True)


def sleep(t):
    '''
    Sleep t seconds
    '''

    print('sleep', t, '...')
    time.sleep(t)
    print('resume')


# Command Line Arguments
parser = argparse.ArgumentParser()
parser.add_argument('-c', '--committee-ip', metavar='', help="IP address of committee node", nargs='+', dest="committee_ip")
parser.add_argument('-s', '--shard-ip', metavar='', help="IP address of shard node", nargs='+', dest="shard_ip")
parser.add_argument('-n', '--shard-member-number', type=int, metavar='', help="Number of members in the shard", dest="number")

#parse Arguments
args = parser.parse_args()

#Input parameter judgment
if args.committee_ip is None or args.shard_ip is None:
    print('please input iP address of committee node and shard node. -h shows options.')
    sys.exit(1)


Pubkey = "1109ef616830cd7b8599ae7958fbee56d4c8168ffd5421a16025a398b8a4be"
start_pubkey = 40
start_port = 2000
committee = []
shard = []

port_index = 0
for ip in args.committee_ip:
    node = {"Pubkey": Pubkey + str(start_pubkey + port_index), "Address": "127.0.0.1", "Port": str(start_port)}
    node["Address"] = ip
    node["Port"] = str(start_port + port_index)
    port_index += 1
    committee.append(node)

for ip in args.shard_ip:
    node = {"Pubkey": Pubkey + str(start_pubkey + port_index), "Address": "127.0.0.1", "Port": str(start_port)}
    node["Address"] = ip
    node["Port"] = str(start_port + port_index)
    port_index += 1
    shard.append(node)

data = {
    "Pubkey": Pubkey,
    "Address": "127.0.0.1",
    "Port": str(start_port),
    "Committee": committee,
    "Shard": shard
}

port_index = 0
for ip in args.committee_ip:
    data["Port"] = str(start_port + port_index)
    data["Pubkey"] = Pubkey + str(start_pubkey + port_index)
    port_index += 1
    with open("sharding.json", 'w') as f:
        json.dump(data, f)
    run("scp sharding.json ecoball@" + ip + ":/home/ecoball/go/src/github.com/ecoball/go-ecoball/build/")
    run("ssh ecoball@" + ip + " \" cd /home/ecoball/go/src/github.com/ecoball/go-ecoball/build/; ./ecoball run & \"")
    sleep(1.5)

for ip in args.shard_ip:
    data["Port"] = str(start_port + port_index)
    data["Pubkey"] = Pubkey + str(start_pubkey + port_index)
    port_index += 1
    with open("sharding.json", 'w') as f:
        json.dump(data, f)
    run("scp sharding.json ecoball@" + ip + ":/home/ecoball/go/src/github.com/ecoball/go-ecoball/build/")
    run("ssh ecoball@" + ip + " \" cd /home/ecoball/go/src/github.com/ecoball/go-ecoball/build/; ./ecoball run\"")
    sleep(1.5)
