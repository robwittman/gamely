#!/usr/bin/env bash

helm upgrade \
  --install \
  minio \
  oci://registry-1.docker.io/bitnamicharts/minio \
  -n default \
  --set "auth.rootPassword=password" \
  --set "defaultBuckets=backups"

kubectl create secret generic backups \
  -n default \
  --from-literal=AWS_ACCESS_KEY_ID=admin \
  --from-literal=AWS_SECRET_ACCESS_KEY=password \
  --dry-run=client \
  --save-config \
  -o yaml | \
  kubectl apply -f -