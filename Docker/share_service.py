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


def run(shell_command):
    '''
    Execute shell command.
    If it fails, exit the program with an exit code of 1.
    '''

    print('shared_start.py:', shell_command)
    if subprocess.call(shell_command, shell=True):
        print('bootstrap.py: exiting because of error')
        sys.exit(1)


# Command Line Arguments
parser = argparse.ArgumentParser()
parser.add_argument('-i', '--node-ip', metavar='', help="IP address of node", nargs='+', dest="node_ip")
parser.add_argument('-o', '--host-ip', metavar='', help="IP address of host node", dest="host_ip")
parser.add_argument('-w', '--weight', type=int, metavar='', help="The number of weights", dest="weight")

#parse Arguments
args = parser.parse_args()

#Input parameter judgment
if args.node_ip is None or args.host_ip is None:
    print('please input iP address of nodes and ip of host node and weight number. -h shows options.')
    sys.exit(1)

start_port = 2000
PORT = 20678
image = "jatel/internal:ecoball_v1.0"
ip_index = args.node_ip.index(args.host_ip)

str_ip = " "
for ip in args.node_ip:
    str_ip += (ip + " ")

count = 0
while count < 4 * args.weight:
    command = "sudo docker run -d " + "--name=ecoball_" + str(count) + " -p "
    command += str(PORT + count) + ":20678 "
    command += str(start_port + ip_index * 4 * args.weight + count) + ":" + str(start_port + ip_index * 4 * args.weight + count)
    command += " " + image + " /root/go/src/github.com/ecoball/go-ecoball/build/Docker/start.py "
    command += "-i" + str_ip + "-o " + args.host_ip + " -n " + str(count) + " -w " + str(args.weight)
    run(command)
    
