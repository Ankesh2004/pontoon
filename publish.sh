#!/bin/bash

read -p "Enter your Docker Hub username: " DOCKER_USER

if [ -z "$DOCKER_USER" ]; then
    echo "Error: Docker Hub username cannot be empty."
    exit 1
fi

echo ""
echo "Building API image..."
docker build --target api -t $DOCKER_USER/pontoon-api:latest .
if [ $? -ne 0 ]; then
    echo "Error building API image."
    exit 1
fi

echo ""
echo "Building Worker image..."
docker build --target worker -t $DOCKER_USER/pontoon-worker:latest .
if [ $? -ne 0 ]; then
    echo "Error building Worker image."
    exit 1
fi

echo ""
echo "Pushing API image..."
docker push $DOCKER_USER/pontoon-api:latest

echo ""
echo "Pushing Worker image..."
docker push $DOCKER_USER/pontoon-worker:latest

echo ""
echo "=============================================="
echo "Success! Both images pushed to Docker Hub!"
echo "=============================================="
echo "On your GCP server, change your docker-compose.yml to:"
echo "api:"
echo "  image: $DOCKER_USER/pontoon-api:latest"
echo "  ..."
echo "worker:"
echo "  image: $DOCKER_USER/pontoon-worker:latest"
echo "  ..."
echo "=============================================="
