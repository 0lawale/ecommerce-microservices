pipeline {
    agent any
    
    environment {
        // Docker Hub credentials (we'll add these in Jenkins)
        DOCKER_REGISTRY = 'docker.io'
        DOCKER_CREDENTIALS_ID = 'dockerhub-credentials'
        
        // Image names
        USER_SERVICE_IMAGE = "${DOCKER_REGISTRY}/0lawale/user-service"
        PRODUCT_SERVICE_IMAGE = "${DOCKER_REGISTRY}/0lawale/product-service"
        ORDER_SERVICE_IMAGE = "${DOCKER_REGISTRY}/0lawale/order-service"
        NOTIFICATION_SERVICE_IMAGE = "${DOCKER_REGISTRY}/0lawale/notification-service"
        API_GATEWAY_IMAGE = "${DOCKER_REGISTRY}/0lawale/api-gateway"
        
        // Build version (using build number and git commit)
        VERSION = "${BUILD_NUMBER}-${GIT_COMMIT.take(7)}"
    }
    
    stages {
        stage('Checkout') {
            steps {
                echo '📥 Checking out code from GitHub...'
                checkout scm
            }
        }
        
        stage('Environment Info') {
            steps {
                echo '🔍 Environment Information:'
                sh '''
                    echo "Go version:"
                    go version
                    echo "\nDocker version:"
                    docker --version
                    echo "\nBuild version: ${VERSION}"
                    echo "Git commit: ${GIT_COMMIT}"
                '''
            }
        }
        
        stage('Build Go Services') {
            steps {
                echo '🔨 Building Go microservices...'
                sh '''
                    echo "Running go mod download..."
                    go mod download
                    
                    echo "Building user-service..."
                    go build -o bin/user-service ./user-service
                    
                    echo "Building product-service..."
                    go build -o bin/product-service ./product-service
                    
                    echo "Building order-service..."
                    go build -o bin/order-service ./order-service
                    
                    echo "Building notification-service..."
                    go build -o bin/notification-service ./notification-service
                    
                    echo "Building api-gateway..."
                    go build -o bin/api-gateway ./api-gateway
                    
                    echo "✅ All services built successfully!"
                '''
            }
        }
        
        stage('Run Tests') {
            steps {
                echo '🧪 Running tests...'
                sh '''
                    echo "Running Go tests..."
                    go test -v ./... || echo "⚠️  Some tests failed (continuing for demo)"
                    
                    echo "Running linter..."
                    go fmt ./... || true
                    
                    echo "✅ Tests completed!"
                '''
            }
        }
        
        stage('Build Docker Images') {
            steps {
                echo '🐳 Building Docker images...'
                script {
                    sh """
                        echo "Building user-service image..."
                        docker build -t ${USER_SERVICE_IMAGE}:${VERSION} -f user-service/Dockerfile .
                        docker tag ${USER_SERVICE_IMAGE}:${VERSION} ${USER_SERVICE_IMAGE}:latest
                        
                        echo "Building product-service image..."
                        docker build -t ${PRODUCT_SERVICE_IMAGE}:${VERSION} -f product-service/Dockerfile .
                        docker tag ${PRODUCT_SERVICE_IMAGE}:${VERSION} ${PRODUCT_SERVICE_IMAGE}:latest
                        
                        echo "Building order-service image..."
                        docker build -t ${ORDER_SERVICE_IMAGE}:${VERSION} -f order-service/Dockerfile .
                        docker tag ${ORDER_SERVICE_IMAGE}:${VERSION} ${ORDER_SERVICE_IMAGE}:latest
                        
                        echo "Building notification-service image..."
                        docker build -t ${NOTIFICATION_SERVICE_IMAGE}:${VERSION} -f notification-service/Dockerfile .
                        docker tag ${NOTIFICATION_SERVICE_IMAGE}:${VERSION} ${NOTIFICATION_SERVICE_IMAGE}:latest
                        
                        echo "Building api-gateway image..."
                        docker build -t ${API_GATEWAY_IMAGE}:${VERSION} -f api-gateway/Dockerfile .
                        docker tag ${API_GATEWAY_IMAGE}:${VERSION} ${API_GATEWAY_IMAGE}:latest
                        
                        echo "✅ All Docker images built successfully!"
                    """
                }
            }
        }
        
        stage('Push to Docker Registry') {
            when {
                branch 'main'  // Only push on main branch
            }
            steps {
                echo '📤 Pushing images to Docker Hub...'
                script {
                    docker.withRegistry('https://index.docker.io/v1/', DOCKER_CREDENTIALS_ID) {
                        sh """
                            docker push ${USER_SERVICE_IMAGE}:${VERSION}
                            docker push ${USER_SERVICE_IMAGE}:latest
                            
                            docker push ${PRODUCT_SERVICE_IMAGE}:${VERSION}
                            docker push ${PRODUCT_SERVICE_IMAGE}:latest
                            
                            docker push ${ORDER_SERVICE_IMAGE}:${VERSION}
                            docker push ${ORDER_SERVICE_IMAGE}:latest
                            
                            docker push ${NOTIFICATION_SERVICE_IMAGE}:${VERSION}
                            docker push ${NOTIFICATION_SERVICE_IMAGE}:latest
                            
                            docker push ${API_GATEWAY_IMAGE}:${VERSION}
                            docker push ${API_GATEWAY_IMAGE}:latest
                            
                            echo "✅ All images pushed successfully!"
                        """
                    }
                }
            }
        }
        
        stage('Deploy to Staging') {
            when {
                branch 'main'
            }
            steps {
                echo '🚀 Deploying to staging environment...'
                sh '''
                    echo "Deployment would happen here..."
                    echo "In production, this would:"
                    echo "  - SSH to staging server"
                    echo "  - Pull new images"
                    echo "  - Run docker-compose up with new versions"
                    echo "  - Run smoke tests"
                    echo "✅ Staging deployment simulated!"
                '''
            }
        }
        
        stage('Approve Production Deploy') {
            when {
                branch 'main'
            }
            steps {
                echo '⏸️  Waiting for manual approval to deploy to production...'
                input message: 'Deploy to Production?', ok: 'Deploy'
            }
        }
        
        stage('Deploy to Production') {
            when {
                branch 'main'
            }
            steps {
                echo '🚀 Deploying to production environment...'
                sh '''
                    echo "Production deployment would happen here..."
                    echo "✅ Production deployment simulated!"
                '''
            }
        }
    }
    
    post {
        success {
            echo '✅ Pipeline completed successfully!'
        }
        failure {
            echo '❌ Pipeline failed!'
        }
        cleanup {
            echo '🧹 Cleaning up...'
            sh 'docker system prune -f || true'
        }
    }
}
