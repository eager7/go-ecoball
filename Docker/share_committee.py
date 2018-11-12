#!/usr/bin/env python3
##########################################################################
# Copyright 2018 The go-ecoball Authors
# This file is part of the go-ecoball.
#
# The go-ecoball is free software: you can redistribute it and/or modify
# it under the terms of the GNU Lesser General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# The go-ecoball is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU Lesser General Public License for more details.
#
# You should have received a copy of the GNU Lesser General Public License
# along with the go-ecoball. If not, see <http://www.gnu.org/licenses/>.
############################################################################
import subprocess
import sys
import argparse
import time
import os
import pytoml
import socket
import json


def run(shell_command):
    '''
    Execute shell command.
    If it fails, exit the program with an exit code of 1.
    '''

    print('share_committee.py:', shell_command)
    if subprocess.call(shell_command, shell=True):
        print('share_committee.py: exiting because of error')
        sys.exit(1)


def sleep(t):
    '''
    Sleep t seconds
    '''

    print('sleep', t, '...')
    time.sleep(t)
    print('resume')


def get_host_ip():
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(('8.8.8.8', 80))
        ip = s.getsockname()[0]
    finally:
        s.close()
    return ip


def get_config(num):
    ip_index = host_ip + "_" + str(num)
    for one in data:
        if one == ip_index:
            return True, data[ip_index]
    return False, ""


# Command Line Arguments
parser = argparse.ArgumentParser()
parser.add_argument('-d', '--deploy-browser-wallet', action='store_true', help="Whether to deploy the browser and wallet", dest="deploy")

# parse Arguments
args = parser.parse_args()

# get netwoek config
with open('shard_setup.toml') as setup_file:
    data = pytoml.load(setup_file)

network = data["network"]
network_str = json.dumps(network)
host_ip = get_host_ip()
committee_count = network[host_ip][0]
shard_count = network[host_ip][1]

#create directory
root_dir = os.path.split(os.path.realpath(__file__))[0]
log_dir = os.path.join(root_dir, 'ecoball_log/committee')
if not os.path.exists(log_dir):
     os.makedirs(log_dir)
    
start_port = 2000
PORT = 20681
image = "jatel/internal:ecoball_v1.0"

count = 0
while count < committee_count:
    # start ecoball
    command = "sudo docker run -d " + "--name=ecoball_" + str(count) + " -p "
    command += str(PORT + count) + ":20678 "
    command += "-p " + str(start_port + count) + ":" + str(start_port + count)
    command += " -v " + log_dir  + ":/var/ecoball_log "
    command += image + " /ecoball/ecoball/start.py "
    command += "-o " + host_ip + " -n " + str(count) + " -e " + "'" + network_str + "'"
    exist, config = get_config(count)
    if not exist:
        config = {"log_dir": "/var/ecoball_log/ecoball_" + str(count)}
    if exist:
        config["log_dir"] = "/var/ecoball_log/ecoball_" + str(count)
    command += " -c " + "'" + json.dumps(config) + "'"
    run(command)
    sleep(2)

    if args.deploy and count == committee_count - 1:
        # start ecowallet
        command = "sudo docker run -d --name=ecowallet -p 20679:20679 "
        command += image + " /ecoball/ecowallet/ecowallet"
        run(command)
        sleep(2)

        # start eballscan
        command = "sudo docker run -d --name=eballscan --link=ecoball_0:ecoball_alias -p 20680:20680 "
        command += image + " /ecoball/eballscan/eballscan_service.sh ecoball_0"
        run(command)
        sleep(2)

    count += 1
    
print("start all ecoball committee container on this physical machine(" + host_ip + ") successfully!!!") 
