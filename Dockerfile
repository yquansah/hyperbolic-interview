FROM node:18.20-alpine3.20 as ui

COPY frontend/package.json .
COPY frontend/package-lock.json .

RUN npm install

COPY frontend  .

RUN npm run build

FROM golang:1.22.4-alpine as builder

WORKDIR /app

COPY . .
COPY --from=ui /dist ./frontend/dist

RUN go build -o /hyperbolic

FROM alpine:latest

COPY --from=builder /hyperbolic /hyperbolic

ENTRYPOINT ["/hyperbolic"]
