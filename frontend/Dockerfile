# Build Vue app
FROM node:20-alpine AS builder
WORKDIR /frontend

ARG MODE=production
# APP_VERSION is passed in from manage-stack.sh via --build-arg.
# It must be declared as ARG and then re-exported as ENV so that
# vite.config.js can read it via process.env.APP_VERSION at build time.
ARG APP_VERSION
ENV APP_VERSION=${APP_VERSION}

COPY package*.json ./
RUN npm ci
COPY . .

RUN npm run build -- --mode ${MODE}

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
