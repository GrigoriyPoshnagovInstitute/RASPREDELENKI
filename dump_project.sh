#!/bin/bash

# Название выходного файла
OUTPUT_FILE="project_context.txt"

# Очищаем выходной файл, если он существует
echo "" > "$OUTPUT_FILE"

echo "Сбор данных в файл $OUTPUT_FILE..."

# Используем find для поиска файлов
# Логика:
# 1. Исключаем папки (-prune), которые нам не нужны.
# 2. Исключаем конкретные файлы по расширениям или именам.
# 3. Для всех остальных файлов записываем заголовок и содержимое.

find . \
    \( \
        -name ".git" \
        -o -name ".idea" \
        -o -name ".vscode" \
        -o -name "bin" \
        -o -name "vendor" \
        -o -name "node_modules" \
        -o -name "build" \
        -o -name "dist" \
        -o -name ".github" \
        -o -name "tmp" \
        -o -name "coverage" \
    \) -prune \
    -o -type f \
    -not -name "*.png" \
    -not -name "*.jpg" \
    -not -name "*.jpeg" \
    -not -name "*.gif" \
    -not -name "*.ico" \
    -not -name "*.svg" \
    -not -name "*.zip" \
    -not -name "*.tar.gz" \
    -not -name "*.exe" \
    -not -name "*.dll" \
    -not -name "*.so" \
    -not -name "*.test" \
    -not -name "go.sum" \
    -not -name ".DS_Store" \
    -not -name "$OUTPUT_FILE" \
    -not -name "dump_project.sh" \
    -print0 | while IFS= read -r -d '' file; do

        # Проверка на бинарный файл (на всякий случай, если расширение пропущено)
        if grep -Iq . "$file"; then
            echo "Processing: $file"
            echo "================================================================================" >> "$OUTPUT_FILE"
            echo "FILE PATH: $file" >> "$OUTPUT_FILE"
            echo "================================================================================" >> "$OUTPUT_FILE"
            cat "$file" >> "$OUTPUT_FILE"
            echo -e "\n\n" >> "$OUTPUT_FILE"
        else
            echo "Skipping binary file: $file"
        fi
done

echo "Готово! Весь код сохранен в $OUTPUT_FILE"