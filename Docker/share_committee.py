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


def run(shell_command):
    '''
    Execute shell command.
    If it fails, exit the program with an exit code of 1.
    '''

    print('shared_start.py:', shell_command)
    if subprocess.call(shell_command, shell=True):
        print('bootstrap.py: exiting because of error')
        sys.exit(1)


def sleep(t):
    '''
    Sleep t seconds
    '''

    print('sleep', t, '...')
    time.sleep(t)
    print('resume')

    
# Command Line Arguments
parser = argparse.ArgumentParser()
parser.add_argument('-i', '--node-ip', required=True, metavar='', help="IP address of node", nargs='+', dest="node_ip")
parser.add_argument('-o', '--host-ip', required=True, metavar='', help="IP address of host node", dest="host_ip")
parser.add_argument('-w', '--weight', type=int, metavar='', help="The number of weights", default=1, dest="weight")
parser.add_argument('-d', '--deploy-browser-wallet', action='store_true', help="Whether to deploy the browser and wallet", dest="deploy")

# parse Arguments
args = parser.parse_args()

# Input parameter judgment
if args.node_ip is None or args.host_ip is None:
    print('please input iP address of nodes and ip of host node and weight number. -h shows options.')
    sys.exit(1)

#create directory
root_dir = os.path.split(os.path.realpath(__file__))[0]
log_dir = os.path.join(root_dir, 'ecoball_log/committee')
if not os.path.exists(log_dir):
     os.makedirs(log_dir)
    
start_port = 2000
PORT = 20681
image = "jatel/internal:ecoball_v1.0"
ip_index = args.node_ip.index(args.host_ip)

str_ip = " "
for ip in args.node_ip:
    str_ip += (ip + " ")

count = 0
while count < args.weight:
    # start ecoball
    command = "sudo docker run -d " + "--name=ecoball_" + str(count) + " -p "
    command += str(PORT + count) + ":20678 "
    command += "-p " + str(start_port + ip_index * 4 * args.weight + count) + ":" + str(start_port + ip_index * 4 * args.weight + count)
    command += " -v " + log_dir  + ":/var/ecoball_log "
    command += image + " /root/go/src/github.com/ecoball/go-ecoball/Docker/start.py "
    command += "-i" + str_ip + "-o " + args.host_ip + " -n " + str(count) + " -w " + str(args.weight)
    command += " --log-dir=/var/ecoball_log/ecoball_" + str(count) + "/"
    print(command)
    run(command)
    sleep(2)

    if args.deploy and count == args.weight - 1:
        # start ecowallet
        command = "sudo docker run -d --name=ecowallet -p 20679:20679 "
        command += image + " /root/go/src/github.com/ecoball/go-ecoball/build/ecowallet"
        print(command)
        run(command)
        sleep(2)

        # start eballscan
        command = "sudo docker run -d --name=eballscan --link=ecoball_0:ecoball_alias -p 20680:20680 "
        command += image + " /root/go/src/github.com/ecoball/eballscan/eballscan_service.sh ecoball_0"
        print(command)
        run(command)
        sleep(2)

    count += 1
    
print("start all ecoball committee container on this physical machine(" + args.host_ip + ") successfully!!!") 
