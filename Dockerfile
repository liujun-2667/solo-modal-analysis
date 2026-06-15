FROM golang:1.21-alpine AS backend-build

WORKDIR /app

COPY backend/ .
RUN go mod tidy && go mod download
RUN go build -o modal-analysis .

FROM node:20-alpine AS frontend-build

WORKDIR /app

COPY frontend/package.json ./
COPY .npmrc ./
RUN npm install

COPY frontend/ .
RUN npm run build

FROM nginx:alpine

COPY --from=backend-build /app/modal-analysis /usr/local/bin/
COPY --from=frontend-build /app/dist /usr/share/nginx/html

COPY nginx.conf /etc/nginx/nginx.conf
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

EXPOSE 80
EXPOSE 8080

CMD ["/entrypoint.sh"]
