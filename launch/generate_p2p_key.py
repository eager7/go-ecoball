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
import platform

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
    goPath = os.getenv("GOPATH")
    
    gen_file = goPath + "/src/github.com/ecoball/go-ecoball/test/rsakeygen/main.go"
    sysstr = platform.system()
    if sysstr == "Windows":
        share_shard.run("cd " + tool_dir + "&& go build -o key_gen.exe " + gen_file)
        key_gen = os.path.join(tool_dir + "/key_gen.exe")
    elif sysstr == "Linux":
        share_shard.run("cd " + tool_dir + "&& go build -o key_gen " + gen_file)
        key_gen = os.path.join(tool_dir + "/key_gen")

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
                


if __name__ == "__main__":
    main()