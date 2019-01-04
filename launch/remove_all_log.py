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
import argparse
import shutil
import platform


def run(shell_command):
    '''
    Execute shell command.
    If it fails, exit the program with an exit code of 1.
    '''

    print('run.py:', shell_command)
    if subprocess.call(shell_command, shell=True):
        print('run.py: exiting because of error')
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
    # get netwoek config
    root_dir = os.path.split(os.path.realpath(__file__))[0]
    with open(os.path.join(root_dir, 'setup.toml')) as setup_file:
        data = pytoml.load(setup_file)

    network = data["network"]

    node_ip = []
    for ip in network:
        node_ip.append(ip)

    producer_count = 0
    candidate_count = 0
    for ip in node_ip:
        host_ip = ip
        producer_count += network[ip][0]
        candidate_count += network[ip][1]

    ecoball_log_dir = os.path.join(root_dir, 'ecoball_log')
    if not os.path.exists(ecoball_log_dir):
        return        

    goPath = os.getenv("GOPATH")

    count = producer_count + candidate_count - 1
    while count >= 0:
        # mkdir and copy ecoball
        run_dir = os.path.join(ecoball_log_dir, 'ecoball_' + str(count))
        if not os.path.exists(run_dir):
            os.makedirs(run_dir)

        log_dir = os.path.join(run_dir, 'log')
        if not os.path.exists(log_dir):
            os.makedirs(log_dir)

        shutil.rmtree(os.path.join(run_dir, 'log'))

        count -= 1        


if __name__ == "__main__":
    main()
