# E-Commerce Microservices Platform

A complete e-commerce platform built with Go microservices, featuring a React frontend and full CI/CD pipeline.

## 🏗️ Architecture

- **User Service** (Port 8081): Authentication & user management
- **Product Service** (Port 8082): Product catalog management
- **Order Service** (Port 8083): Order processing & orchestration
- **Notification Service** (Port 8084): Event-driven notifications
- **API Gateway** (Port 8080): Single entry point for all requests
- **Frontend** (Port 5173/3000): React dashboard with Tailwind CSS

## 🛠️ Tech Stack

**Backend:**
- Go 1.24
- Gin Web Framework
- PostgreSQL 15
- Redis 7
- RabbitMQ 3

**Frontend:**
- React 18
- Vite
- Tailwind CSS
- Lucide Icons

**DevOps:**
- Docker & Docker Compose
- Jenkins CI/CD
- GitHub Webhooks
- Terraform (Infrastructure as Code)
- Ansible (Configuration Management)

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.24+
- Node.js 20+

### Running the Application
```bash
# Start all backend services
docker-compose up -d

# Start frontend (in separate terminal)
cd frontend
npm install
npm run dev
```

### Access Points
- Frontend: http://localhost:5173
- API Gateway: http://localhost:8080
- Jenkins: http://localhost:8090

## 📊 Project Structure
```
ecommerce-microservices/
├── api-gateway/          # API Gateway service
├── user-service/         # User authentication service
├── product-service/      # Product catalog service
├── order-service/        # Order processing service
├── notification-service/ # Notification service
├── frontend/            # React frontend
├── shared/              # Shared Go packages
├── jenkins/             # Jenkins configuration
├── docker-compose.yml   # Service orchestration
└── Jenkinsfile         # CI/CD pipeline definition
```

## 🧪 Testing
```bash
# Run backend tests
go test ./...

# Test services with curl
curl http://localhost:8080/health
```

## 📝 License

MIT License
EOF

echo "✅ README.md created"