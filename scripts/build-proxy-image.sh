#!/bin/bash

image_name="kayetana-proxy-image"

if ! docker build -t $image_name -f build/Dockerfile .; then
    echo "Error while building '$image_name'"
    exit 1
fi
