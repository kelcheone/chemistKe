FROM golang:1.23-alpine AS builder

ARG GOOSE_DBSTRING
ENV GOOSE_DBSTRING=$GOOSE_DBSTRING

ARG DB_URL
ENV DB_URL=$DB_URL

RUN apk add --no-cache  \
  git\
  make\
  build-base\
  curl\
  unzip

ENV PATH="/go/bin:${PATH}"
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

ENV PROTOC_VERSION=29.3
RUN curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
  unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /usr/local && \
  rm protoc-${PROTOC_VERSION}-linux-x86_64.zip

WORKDIR /app

COPY go.mod go.sum ./ 

RUN go mod download

COPY .env* ./

COPY . .

# Run make 
RUN make install-plugins

RUN make prepare


#build both binaries
RUN go build -o services  ./main.go
RUN go build -o gateway ./cmd/api-gateway/gateway.go


FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app

COPY --from=builder /app/services /app/
COPY --from=builder /app/gateway /app/
COPY --from=builder /app/.env* /app/
COPY entrypoint.sh /app/



RUN chmod +x /app/entrypoint.sh

RUN chown  -R appuser:appgroup /app

USER appuser


EXPOSE 9000 9000


ENTRYPOINT [ "/app/entrypoint.sh" ]

