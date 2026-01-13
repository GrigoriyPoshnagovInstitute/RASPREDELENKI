#!/usr/bin/env bash
set -e

echo "--- УНИЧТОЖЕНИЕ СТАРОГО КЛАСТЕРА ---"
./scripts/delete-kind-cluster.sh || true
sudo rm -rf data/mysql || true # Чистим данные базы, чтобы не было конфликтов схем

echo "--- СОЗДАНИЕ КЛАСТЕРА ---"
./scripts/create-kind-cluster.sh

echo "--- ПОДГОТОВКА ИНФРАСТРУКТУРЫ (FIX AMD64) ---"
# Скачиваем правильные архитектуры, чтобы kind не ругался на digest
docker pull --platform linux/amd64 mysql:8.3
docker pull --platform linux/amd64 rabbitmq:3.13-management
docker pull --platform linux/amd64 temporalio/auto-setup:1.29.1
# Пересобираем Temporal UI без provenance, чтобы kind не падал
echo "FROM temporalio/ui:2.34.0" | docker build --provenance=false -t temporalio/ui:2.34.0-clean -
docker tag temporalio/ui:2.34.0-clean temporalio/ui:2.34.0

echo "--- ЗАГРУЗКА ИНФРАСТРУКТУРЫ В KIND ---"
kind load docker-image mysql:8.3 --name rp-practice
kind load docker-image rabbitmq:3.13-management --name rp-practice
kind load docker-image temporalio/auto-setup:1.29.1 --name rp-practice
kind load docker-image temporalio/ui:2.34.0 --name rp-practice

echo "--- СБОРКА И ЗАГРУЗКА ПРИЛОЖЕНИЙ ---"
# Используем твой скрипт, он хороший
./scripts/load-images.sh

echo "--- ПРИМЕНЕНИЕ КОНФИГУРАЦИЙ ---"
# Сначала инфраструктура
kubectl apply -k config/infrastructure

echo "⏳ --- ОЖИДАНИЕ ГОТОВНОСТИ ИНФРАСТРУКТУРЫ (может занять 1-2 мин) ---"
kubectl wait --namespace infrastructure --for=condition=ready pod --selector=app=mysql --timeout=180s
kubectl wait --namespace infrastructure --for=condition=ready pod --selector=app=rabbitmq --timeout=180s
kubectl wait --namespace infrastructure --for=condition=ready pod --selector=app=temporal --timeout=180s

# Патчим Temporal UI (фиксы, которые мы делали руками)
echo "--- ЛЕЧЕНИЕ TEMPORAL UI ---"
kubectl set env deployment/temporal-ui -n infrastructure TEMPORAL_PORT="8080"
kubectl patch deployment temporal-ui -n infrastructure -p '{"spec": {"template": {"spec": {"enableServiceLinks": false}}}}'
# Ждем перезапуска UI
kubectl rollout status deployment/temporal-ui -n infrastructure

echo "--- ЗАПУСК ПРИЛОЖЕНИЙ ---"
# Патчим NotificationService (фикс конфликта портов)
# (Если ты уже поменял файл руками - это не повредит)
sed -i 's/name: NOTIFICATION_HTTP_ADDRESS/name: NOTIFICATION_SERVICE_HTTP_ADDRESS/g' config/application/notificationservice/deployment.yaml || true

kubectl apply -k config/application

echo "⏳ --- ОЖИДАНИЕ ГОТОВНОСТИ ПРИЛОЖЕНИЙ ---"
kubectl wait --namespace application --for=condition=ready pod --all --timeout=300s

echo "✅ --- ВСЕ ГОТОВО! ЗАПУСКАЙ ./ports.sh ---"