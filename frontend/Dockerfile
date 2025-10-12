# Build Vue app
FROM node:20-alpine AS builder
WORKDIR /frontend
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Nginx
FROM nginx:alpine
WORKDIR /usr/share/nginx/html

# Clean default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy your config 
COPY ../config/nginx.conf /etc/nginx/conf.d/default.conf

# Copy built app
COPY --from=builder /frontend/dist .

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
