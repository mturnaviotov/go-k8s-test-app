FROM node:22-alpine AS build
ARG BACKEND_URL
ENV BACKEND_URL=$BACKEND_URL
WORKDIR /app
COPY package*.json ./
COPY public/* ./public/
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]