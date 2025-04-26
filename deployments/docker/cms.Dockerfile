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
# cat the .env file
RUN cat ./.env

COPY . .

RUN make install-plugins
RUN make prepare

RUN CGO_ENABLED=0 GOOS=linux go build -a -o cms-service ./cmd/cms-service/main.go

# ---- FINAL STAGE ----
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/cms-service /cms-service
COPY --from=builder /app/.env* ./

CMD ["/cms-service"]
