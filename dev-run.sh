#!/bin/bash
# Script to run the application in Docker for development/testing

# Build the Docker image
docker build -t cfwg-zt:dev .

# Run the container with mounted configuration
# Create a config.yaml in the current directory before running this
docker run -it --rm \
  --name cfwg-zt \
  --cap-add=NET_ADMIN \
  --device=/dev/net/tun \
  -v "$(pwd)/config.yaml:/etc/cfwg-zt/config.yaml" \
  -v "$(pwd)/logs:/var/log/cfwg-zt" \
  cfwg-zt:dev $@
