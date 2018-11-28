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


def get_config(num, host_ip, data):
    ip_index = host_ip + "_" + str(num)
    for one in data:
        if one == ip_index:
            return True, data[ip_index]
    return False, ""


def main():
    # Command Line Arguments
    parser = argparse.ArgumentParser()
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
    for ip in node_ip:
        host_ip = ip
        committee_count += network[ip][0]
        shard_count += network[ip][1]

    #create directory
    shard_dir = os.path.join(root_dir, 'ecoball/shard')
    if not os.path.exists(shard_dir):
        os.makedirs(shard_dir)        

    committee_dir = os.path.join(root_dir, 'ecoball/committee')
    if not os.path.exists(committee_dir):
        os.makedirs(committee_dir)

    goPath = os.getenv("GOPATH")

    print("build ecoball with the makefile")
    run("make -C " + goPath + "/src/github.com/ecoball/go-ecoball/" + " ecoball")

    count = committee_count + shard_count - 1
    while count >= 0:
        # mkdir and copy ecoball
        if count < committee_count:
            run_dir = os.path.join(committee_dir, 'ecoball_' + str(count))
        else:
            run_dir = os.path.join(shard_dir, 'ecoball_'+ str(count))
        if not os.path.exists(run_dir):
            os.makedirs(run_dir)

        log_dir = os.path.join(run_dir, 'log')
        if not os.path.exists(log_dir):
            os.makedirs(log_dir)

        shutil.copy2(goPath + "/src/github.com/ecoball/go-ecoball/build/ecoball", os.path.join(run_dir, 'ecoball_' + str(count)))
        shutil.copy2("start.py", os.path.join(run_dir, 'start.py'))
        shutil.copy2("ecoball.toml", os.path.join(run_dir, 'ecoball.toml'))

        # start ecoball
        command = run_dir + "/start.py "
        command += "-o " + host_ip + " -n " + str(count) + " -a " + "'" + all_str + "'"
        exist, config = get_config(count, host_ip, data)
        if exist:
            command += " -c " + "'" + json.dumps(config) + "'"
        if "size" in data:
            command += " -s " + str(data["size"])

        run(command)

        count -= 1


if __name__ == "__main__":
    main()
