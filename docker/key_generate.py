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
import os
import sys
import pytoml
import share_shard


def run_shell_output(command, print_output=True, universal_newlines=True):
    p = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True, universal_newlines=universal_newlines)
    if print_output:
        output_array = []
        while True:
            line = p.stdout.readline()
            if not line:
                break
            print(line.strip("\n"))
            output_array.append(line)
        output ="".join(output_array)
    else:
        output = p.stdout.read()
    p.wait()
    errout = p.stderr.read()
    if print_output and errout:
        sys.stdout.write(errout)
    p.stdout.close()
    p.stderr.close()
    return output, p.returncode


def main():
    # Determine if the tool directory exists
    root_dir = os.path.split(os.path.realpath(__file__))[0]
    tool_dir = os.path.join(root_dir, 'tools')
    if not os.path.exists(tool_dir):
        os.makedirs(tool_dir)

    # Generate the latest tools
    gen_file = os.path.join(root_dir, "../test/rsakeygen/main.go")
    share_shard.run("cd " + tool_dir + "&& go build -o key_gen " + gen_file)
    key_gen = os.path.join(tool_dir + "/key_gen")

    #get config
    data = {}
    with open(os.path.join(root_dir, 'shard_setup.toml')) as setup_file:
        data = pytoml.load(setup_file)

    start_port = 9901
    network = data["network"]
    for one_ip in network:
        count_list = network[one_ip]
        for i in range(3):
            count = 0
            while count < count_list[i]:
                result_str, result_code = run_shell_output(key_gen)
                if result_code != 0:
                    print('key_generate.py: exiting because of error')
                    sys.exit(1)
                result_list = result_str.split("\n")
                private_str = ""
                public_str = ""
                for one_str in result_list:
                    index = one_str.find("Private Key:")
                    if -1 != index:
                        private_str = one_str[index + len("Private Key:"):].strip()
                        continue
                    index = one_str.find("Public  Key:") 
                    if -1 != index:
                        public_str = one_str[index + len("Public  Key:"):].strip()
                if 1 == i:
                    tail = count + count_list[0]
                elif 2 == i:
                    tail = count + count_list[0] + count_list[1]                    
                else:
                    tail = count
                one_config = one_ip + "_" + str(tail)
                if one_config not in data:
                    data[one_config] = {}
                data[one_config]["p2p_peer_privatekey"] = private_str
                data[one_config]["p2p_peer_publickey"] = public_str
                port = start_port + tail
                data[one_config]["p2p_listen_address"] = ["/ip4/0.0.0.0/tcp/" + str(port), "/ip6/::/tcp/4013"]
                count += 1

    #new config
    with open(os.path.join(root_dir, 'shard_setup.toml'), 'w') as setup_file:
        pytoml.dump(data, setup_file)


if __name__ == "__main__":
    main()