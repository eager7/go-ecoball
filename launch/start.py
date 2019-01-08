#!/usr/bin/env python3

import subprocess
import sys
import argparse
import json
import os
import pytoml
import platform

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

def get_config(num, host_ip, data):
    ip_index = host_ip + "_" + str(num)
    for one in data:
        if one == ip_index:
            return True, data[ip_index]
    return False, ""

def get_config_p2p(num, host_ip, data):
    ip_index = host_ip + "_" + str(num) + "_p2p"
    for one in data:
        if one == ip_index:
            return True, data[ip_index]
    return False, ""

def main():
    # Command Line Arguments
    parser = argparse.ArgumentParser()
    parser.add_argument('-o', '--host-ip', metavar='', required=True, help="IP address of host node", dest="host_ip")
    parser.add_argument('-n', '--number', type=int, required=True, metavar='', help="The index number of container instance", dest="number")
    # parser.add_argument('-a', '--all-config', metavar='', required=True, help="All configuration information", dest="all_config")
    parser.add_argument('-s', '--size', type=int, default=5, help="Number of nodes per shard")
    parser.add_argument('-c', '--config', metavar='', help="Different configuration items for ecoball.toml")

    root_dir = os.path.split(os.path.realpath(__file__))[0]
    with open(os.path.join(root_dir, 'setup.toml')) as setup_file:
        all_config = pytoml.load(setup_file)

    #parse Arguments
    args = parser.parse_args()

    #Generate the configuration json files required for sharding
    network = all_config["network"]
    for ip in network:
        host_ip = ip

    #Generate the configuration toml files required for ecoball
    ecoball_config = {}
    with open(os.path.join(root_dir, 'ecoball.toml')) as ecoball_file:
        ecoball_config = pytoml.load(ecoball_file)

    # if args.config is not None:
    with open(os.path.join(root_dir, 'ecoball.toml'), 'w') as ecoball_file:
        # config =json.loads(args.config)
        _, config = get_config(args.number, host_ip, all_config)
        for one in config:
            ecoball_config[one] = config[one]

        _, config = get_config_p2p(args.number, host_ip, all_config)
        for one in config:
            ecoball_config["p2p"][one] = config[one]

        pytoml.dump(ecoball_config, ecoball_file)

    #start ecoball
    sysstr = platform.system()
    if sysstr == "Windows":
        run("cd " + os.path.join(root_dir) + "&& start /b ecoball_" + str(args.number) + ".exe" + " run > NUL 2>&1")
    elif sysstr == "Linux":
        run("cd " + os.path.join(root_dir) + "&& ./ecoball_" + str(args.number) + " run > /dev/null 2>&1 &")

   
if __name__ == "__main__":
    main()