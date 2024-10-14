FROM golang:1.23-alpine3.20 as goBuilder

WORKDIR /app/
COPY go.mod /app/
COPY go.sum /app/
RUN go mod download

COPY . /app/
RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /app/  \
    github.com/sergeleger/powermeter/cmd/powermeter


FROM node:22-alpine3.20 as jsBuilder

WORKDIR /app/
COPY SvelteUI/ /app/

RUN npm install
RUN npm run build

FROM alpine:3.20

WORKDIR /app/
COPY --from=goBuilder /app/powermeter /app/
COPY --from=jsBuilder /app/public/ /app/public/

EXPOSE 80

ENV POWERMETER_METER=
ENV POWERMETER_DATABASE=/db/powermeter.db

CMD ["/app/powermeter", "serve", "-http", ":80", "-html", "/app/public"]
