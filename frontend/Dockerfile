FROM node:18-alpine as build

WORKDIR /usr/src/app

# Копируем package.json и package-lock.json
COPY package*.json ./

# Устанавливаем зависимости
RUN npm ci --only=production

# Копируем исходный код
COPY ./public ./public
COPY ./src ./src

# Собираем приложение
RUN npm run build

# Продакшн этап
FROM node:18-alpine

WORKDIR /usr/src/app

# Устанавливаем serve глобально
RUN npm install -g serve

# Копируем собранное приложение из build этапа
COPY --from=build /usr/src/app/build ./build

# Экспортируем порт
EXPOSE 3000

# Команда запуска
CMD ["serve", "-s", "build", "-l", "3000"]