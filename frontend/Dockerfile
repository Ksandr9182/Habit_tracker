# Dockerfile для фронтенда (React)
# Используем официальный образ Node.js для сборки
FROM node:18 AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем package.json и package-lock.json
COPY package.json package-lock.json ./
RUN npm install

# Копируем исходный код фронтенда
COPY . .

# Собираем фронтенд
RUN npm run build

# Используем Nginx для раздачи статических файлов
FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
EXPOSE 80

# Запускаем Nginx
CMD ["nginx", "-g", "daemon off;"]