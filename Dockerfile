FROM python:3.12-slim

WORKDIR /app

# Установка системных зависимостей
RUN apt-get update && apt-get install -y \
    gcc \
    libpq-dev \
    nodejs npm \
    && rm -rf /var/lib/apt/lists/*

# Копирование requirements.txt и установка зависимостей
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# Порт для приложения
EXPOSE 5001

# Создание пользователя для безопасности
RUN useradd -m -u 1000 appuser
USER appuser

# Команда запуска
CMD ["python", "app.py"]
