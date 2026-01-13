#!/usr/bin/env bash

# Функция для убийства всех фоновых процессов при выходе
trap 'kill $(jobs -p)' EXIT

echo "Открываем туннели..."
echo "   - Temporal UI: http://localhost:8080"
echo "   - Services: :8081 (User), :8083 (Product), :8085 (Payment), :8087 (Order)"

kubectl port-forward svc/temporal-ui -n infrastructure 8080:8080 > /dev/null 2>&1 &
kubectl port-forward svc/userservice -n application 8081:8081 > /dev/null 2>&1 &
kubectl port-forward svc/productservice -n application 8083:8081 > /dev/null 2>&1 &
kubectl port-forward svc/paymentservice -n application 8085:8081 > /dev/null 2>&1 &
kubectl port-forward svc/orderservice -n application 8087:8081 > /dev/null 2>&1 &

echo "Туннели активны. Нажми Ctrl+C, чтобы остановить всё."
wait