#!/usr/bin/env bash
set -e

echo "--- УНИЧТОЖЕНИЕ СТАРОГО КЛАСТЕРА ---"
./scripts/delete-kind-cluster.sh || true
# Используем sudo, так как Docker создает файлы от root, и обычный rm падает с Permission denied
sudo rm -rf data || true

echo "--- СОЗДАНИЕ КЛАСТЕРА ---"
./scripts/create-kind-cluster.sh

echo "--- ПОДГОТОВКА ИНФРАСТРУКТУРЫ (FIX DIGEST & PROVENANCE) ---"

# Функция: скачивает, пересобирает образ БЕЗ метаданных (чтобы Kind не падал) и загружает
function clean_and_load {
    IMAGE=$1
    echo ">>> Обработка $IMAGE..."

    # 1. Скачиваем оригинал
    docker pull $IMAGE

    # 2. Пересобираем локально с флагом --provenance=false (убирает attestations, от которых падает kind)
    # Создаем временный тег -clean
    echo "FROM $IMAGE" | docker build --provenance=false -t "$IMAGE-clean" -

    # 3. Вешаем оригинальный тег на наш чистый образ
    docker tag "$IMAGE-clean" $IMAGE

    # 4. Загружаем в Kind
    kind load docker-image $IMAGE --name rp-practice
}

# Применяем лечение ко всей инфраструктуре
clean_and_load "mysql:8.3"
clean_and_load "rabbitmq:3.13-management"
clean_and_load "temporalio/auto-setup:1.29.1"
clean_and_load "temporalio/ui:2.34.0"

echo "--- СБОРКА И ЗАГРУЗКА ПРИЛОЖЕНИЙ ---"
# Твои сервисы собираются локально, у них этой проблемы нет
./scripts/load-images.sh

echo "--- ПРИМЕНЕНИЕ КОНФИГУРАЦИЙ ---"
# Сначала инфраструктура
kubectl apply -k config/infrastructure

echo "⏳ --- ОЖИДАНИЕ ГОТОВНОСТИ ИНФРАСТРУКТУРЫ (может занять 1-2 мин) ---"
kubectl wait --namespace infrastructure --for=condition=ready pod --selector=app=mysql --timeout=180s
kubectl wait --namespace infrastructure --for=condition=ready pod --selector=app=rabbitmq --timeout=180s
kubectl wait --namespace infrastructure --for=condition=ready pod --selector=app=temporal --timeout=180s

# Патчи для Temporal UI (фикс порта и ссылок)
echo "--- ЛЕЧЕНИЕ TEMPORAL UI ---"
kubectl set env deployment/temporal-ui -n infrastructure TEMPORAL_PORT="8080"
kubectl patch deployment temporal-ui -n infrastructure -p '{"spec": {"template": {"spec": {"enableServiceLinks": false}}}}'
# Ждем перезапуска UI
kubectl rollout status deployment/temporal-ui -n infrastructure

echo "--- ЗАПУСК ПРИЛОЖЕНИЙ ---"
# Применяем твои новые конфиги (где мы разбили deployment на 3 части)
kubectl apply -k config/application

echo "⏳ --- ОЖИДАНИЕ ГОТОВНОСТИ ПРИЛОЖЕНИЙ ---"
# Ждем пока все поды поднимутся
kubectl wait --namespace application --for=condition=ready pod --all --timeout=300s

echo "✅ --- ВСЕ ГОТОВО! ЗАПУСКАЙ ./ports.sh ---"