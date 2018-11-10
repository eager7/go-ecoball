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

    print('shared_start.py:', shell_command)
    if subprocess.call(shell_command, shell=True):
        print('shared_start.py: exiting because of error')
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


# get netwoek config
with open('shard_setup.toml') as setup_file:
    data = pytoml.load(setup_file)

network = data["network"]
network_str = json.dumps(network)
ip_str = ""
for ip in network:
    ip_str += ip

host_ip = get_host_ip()
committee_count = network[host_ip][0]
shard_count = network[host_ip][1]

#create directory
root_dir = os.path.split(os.path.realpath(__file__))[0]
log_dir = os.path.join(root_dir, 'ecoball_log/shard')
if not os.path.exists(log_dir):
     os.makedirs(log_dir)

start_port = 2000
PORT = 20681
image = "jatel/internal:ecoball_v1.0"

count = committee_count - 1
while count < committee_count + shard_count:
    # start ecoball
    command = "sudo docker run -d " + "--name=ecoball_" + str(count) + " -p "
    command += str(PORT + count) + ":20678 "
    command += "-p " + str(start_port + count) + ":" + str(start_port + count)
    command += " -v " + log_dir  + ":/var/ecoball_log "
    command += image + " /ecoball/ecoball/start.py "
    command += "-i" + ip_str + "-o " + host_ip + " -n " + str(count) + " -e " + network_str
    exist, config = get_config(count)
    if exist:
        command += " -c " + json.dumps(config)
    run(command)
    sleep(2)
    count += 1
    

print("start all ecoball shard container on this physical machine(" + host_ip + ") successfully!!!") 
