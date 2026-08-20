#!/usr/bin/env bash

set -e

cleanup() {
  echo "Stopping port-forwards..."
  kill $(jobs -p) 2>/dev/null || true
}

trap cleanup EXIT INT TERM

echo "Starting Grafana..."
kubectl port-forward \
  -n monitoring \
  svc/monitoring-grafana \
  3000:80 &

echo "Starting Prometheus..."
kubectl port-forward \
  -n monitoring \
  svc/monitoring-kube-prometheus-prometheus \
  9090:9090 &

echo "Starting Platform PostgreSQL..."
kubectl port-forward \
  -n postgres-demo \
  svc/cloud3-platform-db-rw \
  5432:5432 &

echo
echo "Grafana:    http://localhost:3000"
echo "Prometheus: http://localhost:9090"
echo "Postgres:   localhost:5432"
echo
echo "Press Ctrl+C to stop all port-forwards."

wait