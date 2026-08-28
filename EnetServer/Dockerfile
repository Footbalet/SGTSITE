FROM ubuntu:22.04

# Устанавливаем зависимости для Godot
RUN apt-get update && apt-get install -y \
    libgl1-mesa-glx \
    libglib2.0-0 \
    libfreetype6 \
    libx11-xcb1 \
    && rm -rf /var/lib/apt/lists/*

# Создаем директорию для приложения
WORKDIR /app

# Копируем бинарники
COPY LFLServer.x86_64 .

# Даем права на выполнение
RUN chmod +x MyGame.x86_64

# Пробрасываем порты
EXPOSE 8010

# Запускаем сервер
CMD ["./LFLServer.x86_64", "--headless"]