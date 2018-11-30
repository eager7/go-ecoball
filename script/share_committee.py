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
import argparse
import os
import pytoml
import json
import share_shard


def main():
    # Command Line Arguments
    parser = argparse.ArgumentParser()
    parser.add_argument('-i', '--host-ip', metavar='', required=True, help="IP address of host node", dest="host_ip")
    parser.add_argument('-b', '--deploy-browser', action='store_true', help="Whether to deploy the browsert", dest="browser")
    parser.add_argument('-w', '--deploy-wallet', action='store_true', help="Whether to deploy the wallet", dest="wallet")

    # parse Arguments
    args = parser.parse_args()

    # get netwoek config
    root_dir = os.path.split(os.path.realpath(__file__))[0]
    with open(os.path.join(root_dir, 'shard_setup.toml')) as setup_file:
        data = pytoml.load(setup_file)

    network = data["network"]
    all_str = json.dumps(data)
    host_ip = args.host_ip
    committee_count = network[host_ip][0]

    #create directory
    log_dir = os.path.join(root_dir, 'ecoball_log/committee')
    if not os.path.exists(log_dir):
        os.makedirs(log_dir)
        
 
    p2p_start = 9901
    ipfs_start = 5000
    ipfs_gateway = 7000
    PORT = 20681
    image = "registry.quachain.net:5000/ecoball:1.0.0"

    count = 0
    while count < committee_count:
        # start ecoball
        command = "docker run -d " + "--name=ecoball_" + str(count) + " -p "
        command += str(PORT + count) + ":20678 "
        command += "-p " + str(ipfs_start + count) + ":5011 " 
        command += "-p " + str(ipfs_gateway + count) + ":7011 " 
        command += " -p " + str(p2p_start + count) + ":" + str(p2p_start + count)
        command += " -v " + log_dir  + ":/var/ecoball_log "
        command += image + " /ecoball/ecoball/start.py "
        command += "-o " + host_ip + " -n " + str(count) + " -a " + "'" + all_str + "'"
        exist, config = share_shard.get_config(count, host_ip, data)
        if not exist:
            config = {"log_dir": "/var/ecoball_log/ecoball_" + str(count) + "/",
            "root_dir": "/var/ecoball_log/ecoball_" + str(count) + "/"}
        if exist:
            config["log_dir"] = "/var/ecoball_log/ecoball_" + str(count) + "/"
            config["root_dir"] = "/var/ecoball_log/ecoball_" + str(count) + "/"
        command += " -c " + "'" + json.dumps(config) + "'"
        if "size" in data:
            command += " -s " + str(data["size"])
        share_shard.run(command)
        # share_shard.sleep(2)

        if args.browser and count == committee_count - 1:
            # start eballscan
            command = "docker run -d --name=eballscan --link=ecoball_0:ecoball_alias -p 20680:20680 "
            command += image + " /ecoball/eballscan/eballscan_service.sh ecoball_0"
            share_shard.run(command)
            share_shard.sleep(2)

        if args.wallet and count == committee_count - 1:
            # start ecowallet
            command = "docker run -d --name=ecowallet -p 20679:20679 "
            command += "-v " + root_dir + ":/var "
            command += image + " /ecoball/ecowallet/ecowallet start -p /var"
            share_shard.run(command)
            share_shard.sleep(2)

        count += 1
        
    print("start all ecoball committee container on this physical machine(" + host_ip + ") successfully!!!") 


if __name__ == "__main__":
    main()
