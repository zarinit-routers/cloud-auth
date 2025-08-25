FROM python:3.12-slim

WORKDIR /app

# Установка системных зависимостей
# RUN apt-get update && apt-get install -y \
#     gcc \
#     libpq-dev \
#     && rm -rf /var/lib/apt/lists/*

# Копирование requirements.txt и установка зависимостей
COPY . .
RUN pip install --no-cache-dir -r requirements.txt

# Создание пользователя для безопасности
RUN useradd -m -u 1000 appuser
USER appuser

# Порт для приложения
EXPOSE 5000

# Команда запуска
CMD ["python", "app.py"]
