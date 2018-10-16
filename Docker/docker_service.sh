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

SERVICE=`ps -ef | grep /usr/bin/dockerd | wc -l`
IMAGE="jatel/internal:ecoball_v1.0"

#install docker
if [ ! -e /usr/bin/docker ]; then
    sudo apt-get update
    sudo apt-get install docker
fi

#start docker service
if [ 2 -ne $SERVICE ]; then
    if ! sudo service docker start
    then
        echo  -e "\033[;31m docker service start failed!!! \033[0m"
        exit 1
    fi
fi

#pull docker images
IMAGENUM=`sudo docker images jatel/internal:ecoball_v1.0 | wc -l`
if [ 1 -eq $IMAGENUM ]; then
    if ! sudo docker pull $IMAGE
    then
        echo  -e "\033[;31m pull $IMAGE failed!!! \033[0m"
        exit 1
    fi
fi

#run ecoball docker images
NUM=21
PORT=20677
for((i=1;i<=$NUM;i++))
do   
    PORT=`expr $PORT + 1`
    if [ 20679 -eq $PORT ]; then
        PORT=`expr $PORT + 1`
    fi
    if ! sudo docker run -d -p $PORT:20678 $IMAGE
    then
        echo  -e "\033[;31m docker run failed!!! \033[0m"
        exit 1
    fi
done

#run ecowallet docker images
if ! sudo docker run -d -p 20679:20679 jatel/internal:ecoball_v1.0 /root/go/src/github.com/ecoball/go-ecoball/build/ecowallet
then
    echo  -e "\033[;31m docker run start ecowallet failed!!! \033[0m"
    exit 1
fi