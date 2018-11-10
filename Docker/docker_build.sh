#!/bin/bash
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
cat <<EOF 
Before executing this script, by default you have docker installed and the docker service is enabled.

You've got the full ecoball and eballscan code, and the go environment is configured.

The script builds ecoball and eballscan with the makefile, 
making sure that the two engineering environment configurations are configured.

Eballscan relies on cockroachdb, please put the cockroach file in the current directory, 
otherwise it will be downloaded over the network, 
the network unstable script will take a long time or will fail to execute, the script needs to be reexecuted.
EOF

SOURCE_DIR=$(cd `dirname $0` && pwd)

# check cockroachdb and download
if [ ! -e ${SOURCE_DIR}/cockroach ]; then
    echo -e "\033[;34m downloaded cockroachdb over the network. \033[0m"
    wget -qO- https://binaries.cockroachdb.com/cockroach-v2.0.6.linux-amd64.tgz | tar  xvz
    if [ 0 -ne $? ]; then
        echo  -e "\033[;31m Unable to download cockroach-v2.0.6.linux-amd64.tgz at this time!!! \033[0m"
        exit 1
    fi

    cp -i cockroach-v2.0.6.linux-amd64/cockroach $SOURCE_DIR
    if [ 0 -ne $? ]; then
        echo  -e "\033[;31m install cockroach-v2.0.6.linux-amd64 failed!!! \033[0m"
        exit 1
    fi

    if ! rm -fr "./cockroach-v2.0.6.linux-amd64"
    then
        echo  -e "\033[;31m remove cockroach-v2.0.6.linux-amd64 directory failed!!! \033[0m"
        exit 1
    fi
fi

echo -e "\033[;32m get cockroach succeed. \033[0m"


# build ecoball
if ! make -C ${SOURCE_DIR}/../ ecoball
then
    echo  -e "\033[;31m compile ecoball failed!!! \033[0m"
    exit 1
fi

if ! cp ${SOURCE_DIR}/../build/ecoball ${SOURCE_DIR}
then
    echo  -e "\033[;31m copy ecoball failed!!! \033[0m"
    exit 1
fi

echo -e "\033[;32m get ecoball succeed. \033[0m"


# build ecowallet
if ! make -C ${SOURCE_DIR}/../ ecowallet
then
    echo  -e "\033[;31m compile ecowallet failed!!! \033[0m"
    exit 1
fi

if ! cp ${SOURCE_DIR}/../build/ecowallet ${SOURCE_DIR}
then
    echo  -e "\033[;31m copy ecowallet failed!!! \033[0m"
    exit 1
fi

echo -e "\033[;32m get ecowallet succeed. \033[0m"


# build eballscan
if ! make -C ${SOURCE_DIR}/../../eballscan eballscan
then
    echo  -e "\033[;31m compile eballscan failed!!! \033[0m"
    exit 1
fi

if ! cp ${SOURCE_DIR}/../../eballscan/build/eballscan ${SOURCE_DIR}
then
    echo  -e "\033[;31m copy eballscan failed!!! \033[0m"
    exit 1
fi

if ! cp ${SOURCE_DIR}/../../eballscan/eballscan_service.sh ${SOURCE_DIR}
then
    echo  -e "\033[;31m copy eballscan_service.sh failed!!! \033[0m"
    exit 1
fi

echo -e "\033[;32m get eballscan and eballscan_service.sh succeed. \033[0m"
echo -e "\033[;32m All executable files have been successful and ecoball images can now be created. \033[0m"