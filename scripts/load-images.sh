#!/usr/bin/env bash
set -euo pipefail

echo "Building images..."
docker build -t userservice:latest -f rp-userservice-main/Dockerfile rp-userservice-main/
docker build -t productservice:latest -f rp-productservice/Dockerfile rp-productservice/
docker build -t orderservice:latest -f rp-orderservice/Dockerfile rp-orderservice/
docker build -t paymentservice:latest -f rp-paymentservice/Dockerfile rp-paymentservice/
docker build -t notificationservice:latest -f rp-notificationservice/Dockerfile rp-notificationservice/

echo "Loading images into Kind..."
kind load docker-image userservice:latest --name rp-practice
kind load docker-image productservice:latest --name rp-practice
kind load docker-image orderservice:latest --name rp-practice
kind load docker-image paymentservice:latest --name rp-practice
kind load docker-image notificationservice:latest --name rp-practice