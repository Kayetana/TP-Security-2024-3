#!/bin/bash

container_name="kayetana-proxy"
image_name="kayetana-proxy-image"

if [[ $(docker ps -aq --filter="name=$container_name") ]]; then
  docker rm "$(docker stop "$(docker ps -aq --filter="name=$container_name")")" > /dev/null
  echo "Container '$container_name' removed"
else
  echo "Container '$container_name' not found"
fi

if docker images | grep $image_name > /dev/null; then
  docker rmi $image_name > /dev/null
  echo "Image '$image_name' removed"
else
  echo "Image '$image_name' not found"
fi
