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
import json
import argparse
import shutil
import platform

def run(shell_command):
    '''
    Execute shell command.
    If it fails, exit the program with an exit code of 1.
    '''

    print('stop.py:', shell_command)
    if subprocess.call(shell_command, shell=True):
        print('stop.py: exiting because of error')
        sys.exit(1)


def sleep(t):
    '''
    Sleep t seconds
    '''

    print('sleep', t, '...')
    time.sleep(t)
    print('resume')


def main():
    # Command Line Arguments
    parser = argparse.ArgumentParser()
    parser.add_argument('-i', '--host-ip', metavar='', help="IP address of host node", dest="host_ip")
    parser.add_argument('-b', '--deploy-browser', action='store_true', help="Whether to deploy the browsert", dest="browser")
    parser.add_argument('-w', '--deploy-wallet', action='store_true', help="Whether to deploy the wallet", dest="wallet")
    args = parser.parse_args()

    # get netwoek config
    root_dir = os.path.split(os.path.realpath(__file__))[0]
    with open(os.path.join(root_dir, 'shard_setup.toml')) as setup_file:
        data = pytoml.load(setup_file)

    network = data["network"]
    all_str = json.dumps(data)

    node_ip = []
    for ip in network:
        node_ip.append(ip)

    committee_count = 0
    shard_count = 0
    candidate_count = 0
    for ip in node_ip:
        committee_count += network[ip][0]
        shard_count += network[ip][1]
        candidate_count += network[ip][2]

    sysstr = platform.system()
    count = 0
    while count < committee_count + shard_count + candidate_count:
        # stop ecoball
        if sysstr == "Windows":
            command = "taskkill /im " + "ecoball_" + str(count) + ".exe /F"
        elif sysstr == "Linux":
            command = "killall " + "ecoball_" + str(count)        
        run(command)
        sleep(1)
        count += 1


if __name__ == "__main__":
    main()
