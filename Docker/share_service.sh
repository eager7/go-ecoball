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

SOURCE_DIR=$(cd `dirname $0` && pwd)
IMAGE="jatel/internal:ecoball_v1.0"

if [ $# -ne 5 ]; then
    echo  -e "\033[;31m You must enter the IP addresses of five servers, and the first IP address is the local IP address\033[0m"
    exit 1
fi

#pull docker images
IMAGENUM=`sudo docker images $IMAGE | wc -l`
if [ 1 -eq $IMAGENUM ]; then
    if ! sudo docker pull $IMAGE
    then
        echo  -e "\033[;31m pull $IMAGE failed!!! \033[0m"
        exit 1
    fi
fi

#create ecoball log directory
if [ ! -e "${SOURCE_DIR}/ecoball_log" ]; then
    if ! mkdir "${SOURCE_DIR}/ecoball_log"
    then
        echo  -e "\033[;31m create ecoball log directory failed!!! \033[0m"
        exit 1
    fi
fi

#get all ip
PARA_COUNT=0
IP_ARRAY=()
for ip in $*
do
    if [ 0 -ne $PARA_COUNT ]
    then
        IP_ARRAY[$(expr $PARA_COUNT - 1)]=ip
    fi
    PARA_COUNT=$(expr $PARA_COUNT + 1)
done

#Start all docker instances
INDEX=0
PORT=20678
while [ $INDEX -lt 4 ]
do
    PORT=$(expr $PORT + $INDEX)
    if ! sudo docker run -d --name=ecoball_${INDEX} -p $PORT:20678  $IMAGE  /root/go/src/github.com/ecoball/go-ecoball/build/Docker/start.py -i $IP_ARRAY[@] -o $IP_ARRAY[0] -n $INDEX
    then
        echo  -e "\033[;31m docker run start ecoball_${INDEX} failed!!! \033[0m"
        exit 1
    fi
    INDEX=$(expr $INDEX + 1)
done