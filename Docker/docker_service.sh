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
IMAGE="zhongxh/internal:ecoball_v1.0"

#pull docker images
IMAGENUM=`docker images $IMAGE | wc -l`
if [ 1 -eq $IMAGENUM ]; then
    if ! docker pull $IMAGE
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

case $1 in
    "start")
    #Start all stopped containers
    for i in $(docker ps -a --filter 'exited=137' | sed '1d' | awk '$2=="'"$IMAGE"'"{print $NF | "sort -r -n"}')
    do
        docker start $i
    done

    echo  -e "\033[47;34m start all ecoball and wallet and eballscan success!!! \033[0m"
    ;;

    "stop")
    #stop container
    for i in $(docker ps | sed '1d' | awk '$2=="'"$IMAGE"'"{print $1}')
    do
        docker stop $i
    done
    echo  -e "\033[47;34m stop all container success!!! \033[0m"
    ;;

    "remove")
    #remove container
    for i in $(docker ps -a | sed '1d' | awk '$2=="'"$IMAGE"'"{print $1}')
    do
        docker rm $i
    done

    echo  -e "\033[47;34m remove all container success!!! \033[0m"
    ;;

    *)
    echo "please input docker_service start|stop|remove"
    ;;
    
esac
