# ── Build stage ────────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o blog . && ./blog build

# ── Serve stage ────────────────────────────────────────────────────────────────
FROM nginx:alpine AS runner

RUN rm -rf /usr/share/nginx/html/*

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/templates/default.conf.template

RUN apk add --no-cache gettext

EXPOSE 80
CMD ["/bin/sh", "-c", "envsubst '${ANALYTICS_ID} ${ANALYTICS_SCRIPT_URL}' < /etc/nginx/templates/default.conf.template > /etc/nginx/conf.d/default.conf && nginx -g 'daemon off;'"]
