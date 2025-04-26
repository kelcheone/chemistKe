
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git make build-base curl unzip

ENV PROTOC_VERSION=29.3
RUN curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
    unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /usr/local && \
    rm protoc-${PROTOC_VERSION}-linux-x86_64.zip

ENV PATH="/go/bin:${PATH}"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# copy .env.docker to .env
COPY .env.docker ./.env

COPY . .

RUN make install-plugins
RUN make prepare

RUN CGO_ENABLED=0 GOOS=linux go build -a -o user-service ./cmd/user-service/main.go

# ---- FINAL STAGE ----
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/user-service /user-service
COPY --from=builder /app/.env* ./

CMD ["/user-service"]
