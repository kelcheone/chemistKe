# Chemist.ke Monorepo Folder Structure

```
chemist-ke/
│
├── cmd/
│   ├── api-gateway/
│   ├── user-service/
│   ├── product-service/
│   ├── order-service/
│   ├── telehealth-service/
│   ├── cms-service/
│   └── notification-service/
│
├── internal/
│   ├── auth/
│   ├── database/
│   ├── logger/
│   ├── middleware/
│   └── models/
│
├── pkg/
│   ├── cache/
│   ├── config/
│   ├── errors/
│   └── utils/
│
├── api/
│   ├── proto/
│   │   ├── user/
│   │   ├── product/
│   │   ├── order/
│   │   ├── telehealth/
│   │   ├── cms/
│   │   └── notification/
│   │
│   └── swagger/
│
├── scripts/
│
├── deployments/
│   ├── docker/
│   └── kubernetes/
│
├── test/
│   ├── integration/
│   └── load/
│
├── docs/
│
├── web/
│   └── admin/
│
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Directory Explanations

1. `cmd/`: Contains the main applications for each service.
   - Each subdirectory represents a standalone microservice.

2. `internal/`: Houses packages used across multiple services but not meant for external use.
   - `auth/`: Authentication and authorization logic.
   - `database/`: Database connection and management.
   - `logger/`: Logging utilities.
   - `middleware/`: Shared middleware.
   - `models/`: Data models used across services.

3. `pkg/`: Shared packages that could potentially be used by external projects.
   - `cache/`: Caching mechanisms (e.g., Redis client).
   - `config/`: Configuration management.
   - `errors/`: Custom error types and handling.
   - `utils/`: General utility functions.

4. `api/`: API-related files.
   - `proto/`: Protocol Buffer definitions for gRPC services.
   - `swagger/`: Swagger/OpenAPI specifications.

5. `scripts/`: Utility scripts for development, CI/CD, etc.

6. `deployments/`: Deployment configurations.
   - `docker/`: Dockerfiles and related configs.
   - `kubernetes/`: Kubernetes manifests.

7. `test/`: Test suites beyond unit tests.
   - `integration/`: Integration tests.
   - `load/`: Load and performance tests.

8. `docs/`: Project documentation.

9. `web/`: Web-related components.
   - `admin/`: Admin panel frontend (if applicable).

10. Root files:
    - `.gitignore`: Git ignore file.
    - `go.mod` and `go.sum`: Go module files.
    - `README.md`: Project readme.
