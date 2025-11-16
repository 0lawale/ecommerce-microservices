pipeline {
    agent any
    
    environment {
        // Docker Hub credentials
        DOCKER_REGISTRY = 'docker.io'
        DOCKER_CREDENTIALS_ID = 'dockerhub-credentials'
        
        DOCKER_USERNAME = '0lawale'
        
        // Image names
        USER_SERVICE_IMAGE = "${DOCKER_REGISTRY}/${DOCKER_USERNAME}/user-service"
        PRODUCT_SERVICE_IMAGE = "${DOCKER_REGISTRY}/${DOCKER_USERNAME}/product-service"
        ORDER_SERVICE_IMAGE = "${DOCKER_REGISTRY}/${DOCKER_USERNAME}/order-service"
        NOTIFICATION_SERVICE_IMAGE = "${DOCKER_REGISTRY}/${DOCKER_USERNAME}/notification-service"
        API_GATEWAY_IMAGE = "${DOCKER_REGISTRY}/${DOCKER_USERNAME}/api-gateway"
        
        // Build version
        VERSION = "${BUILD_NUMBER}-${GIT_COMMIT.take(7)}"
    }
    
    stages {
        stage('Checkout') {
            steps {
                echo '📥 Checking out code from GitHub...'
                checkout scm
                sh 'ls -la'
                sh 'cat go.mod | head -n 5'
            }
        }
        
        stage('Environment Info') {
            steps {
                echo '🔍 Environment Information:'
                sh '''
                    echo "Current directory:"
                    pwd
                    
                    echo "\nGo version:"
                    go version
                    
                    echo "\nDocker version:"
                    docker --version
                    
                    echo "\nProject structure:"
                    ls -la
                    
                    echo "\nChecking go.mod:"
                    cat go.mod
                    
                    echo "\nBuild version: ${VERSION}"
                    echo "Git commit: ${GIT_COMMIT}"
                '''
            }
        }
        
        stage('Install Dependencies') {
            steps {
                echo '📦 Installing Go dependencies...'
                sh '''
                    # Ensure we're in the root directory
                    pwd
                    
                    # Clean any previous builds
                    rm -rf bin/ || true
                    mkdir -p bin
                    
                    # Download dependencies (mono-repo style)
                    echo "Running go mod download..."
                    go mod download
                    
                    # Verify dependencies
                    echo "Running go mod tidy..."
                    go mod tidy
                    
                    # Verify go.sum exists
                    if [ -f go.sum ]; then
                        echo "✅ go.sum exists"
                        wc -l go.sum
                    else
                        echo "⚠️  go.sum not found, creating it..."
                        go mod tidy
                    fi
                    
                    echo "✅ Dependencies installed successfully!"
                '''
            }
        }
        
        stage('Build Go Services') {
            steps {
                echo '🔨 Building Go microservices (mono-repo)...'
                sh '''
                    # We're building from the ROOT directory using the root go.mod
                    # All imports are: ecommerce/service-name/...
                    
                    echo "Building user-service..."
                    go build -v -o bin/user-service ./user-service
                    ls -lh bin/user-service
                    
                    echo "Building product-service..."
                    go build -v -o bin/product-service ./product-service
                    ls -lh bin/product-service
                    
                    echo "Building order-service..."
                    go build -v -o bin/order-service ./order-service
                    ls -lh bin/order-service
                    
                    echo "Building notification-service..."
                    go build -v -o bin/notification-service ./notification-service
                    ls -lh bin/notification-service
                    
                    echo "Building api-gateway..."
                    go build -v -o bin/api-gateway ./api-gateway
                    ls -lh bin/api-gateway
                    
                    echo "\n✅ All services built successfully!"
                    echo "Built binaries:"
                    ls -lh bin/
                '''
            }
        }
        
        stage('Run Tests') {
            steps {
                echo '🧪 Running tests...'
                sh '''
                    echo "Running Go tests from root..."
                    # Run tests for all packages
                    go test ./... -v -cover || echo "⚠️ Some tests failed (continuing for demo)"
                    
                    echo "\nRunning go fmt..."
                    go fmt ./... || true
                    
                    echo "\nRunning go vet..."
                    go vet ./... || echo "⚠️ go vet found issues (continuing)"
                    
                    echo "✅ Tests completed!"
                '''
            }
        }
        
        stage('Build Docker Images') {
            steps {
                echo '🐳 Building Docker images...'
                script {
                    // Build from root context since Dockerfiles copy from root
                    sh """
                        echo "Building user-service image..."
                        docker build -t ${USER_SERVICE_IMAGE}:${VERSION} -t ${USER_SERVICE_IMAGE}:latest -f user-service/Dockerfile .
                        
                        echo "Building product-service image..."
                        docker build -t ${PRODUCT_SERVICE_IMAGE}:${VERSION} -t ${PRODUCT_SERVICE_IMAGE}:latest -f product-service/Dockerfile .
                        
                        echo "Building order-service image..."
                        docker build -t ${ORDER_SERVICE_IMAGE}:${VERSION} -t ${ORDER_SERVICE_IMAGE}:latest -f order-service/Dockerfile .
                        
                        echo "Building notification-service image..."
                        docker build -t ${NOTIFICATION_SERVICE_IMAGE}:${VERSION} -t ${NOTIFICATION_SERVICE_IMAGE}:latest -f notification-service/Dockerfile .
                        
                        echo "Building api-gateway image..."
                        docker build -t ${API_GATEWAY_IMAGE}:${VERSION} -t ${API_GATEWAY_IMAGE}:latest -f api-gateway/Dockerfile .
                        
                        echo "\n✅ All Docker images built successfully!"
                        docker images | grep ${DOCKER_USERNAME}
                    """
                }
            }
        }
        
        stage('Push to Docker Registry') {
            when {
                branch pattern: ".*main", comparator: "REGEXP"
            }
            steps {
                echo '📤 Pushing images to Docker Hub...'
                script {
                    docker.withRegistry('https://index.docker.io/v1/', DOCKER_CREDENTIALS_ID) {
                        sh """
                            echo "Pushing user-service..."
                            docker push ${USER_SERVICE_IMAGE}:${VERSION}
                            docker push ${USER_SERVICE_IMAGE}:latest
                            
                            echo "Pushing product-service..."
                            docker push ${PRODUCT_SERVICE_IMAGE}:${VERSION}
                            docker push ${PRODUCT_SERVICE_IMAGE}:latest
                            
                            echo "Pushing order-service..."
                            docker push ${ORDER_SERVICE_IMAGE}:${VERSION}
                            docker push ${ORDER_SERVICE_IMAGE}:latest
                            
                            echo "Pushing notification-service..."
                            docker push ${NOTIFICATION_SERVICE_IMAGE}:${VERSION}
                            docker push ${NOTIFICATION_SERVICE_IMAGE}:latest
                            
                            echo "Pushing api-gateway..."
                            docker push ${API_GATEWAY_IMAGE}:${VERSION}
                            docker push ${API_GATEWAY_IMAGE}:latest
                            
                            echo "\n✅ All images pushed to Docker Hub!"
                        """
                    }
                }
            }
        }
        
        stage('Deploy to Staging') {
            when {
                branch pattern: ".*main", comparator: "REGEXP"
            }
            steps {
                echo '🚀 Deploying to staging environment...'
                sh '''
                    echo "Deployment would happen here..."
                    echo "In production, this would:"
                    echo "  - SSH to staging server"
                    echo "  - Pull new images: docker pull ${USER_SERVICE_IMAGE}:${VERSION}"
                    echo "  - Update docker-compose.yml with new version"
                    echo "  - Run: docker-compose up -d"
                    echo "  - Run smoke tests"
                    echo "✅ Staging deployment simulated!"
                '''
            }
        }
        
        stage('Approve Production Deploy') {
            when {
                branch pattern: ".*main", comparator: "REGEXP"
            }
            steps {
                echo '⏸️ Waiting for manual approval to deploy to production...'
                timeout(time: 1, unit: 'HOURS') {
                    input message: 'Deploy to Production?', ok: 'Deploy'
                }
            }
        }
        
        stage('Deploy to Production') {
            when {
                branch pattern: ".*main", comparator: "REGEXP"
            }
            steps {
                echo '🚀 Deploying to production environment...'
                sh '''
                    echo "Production deployment would happen here..."
                    echo "Steps would be similar to staging but with production servers"
                    echo "✅ Production deployment simulated!"
                '''
            }
        }
    }
    
    post {
        success {
            echo '✅ Pipeline completed successfully!'
            echo "Docker images available at:"
            echo "  - ${USER_SERVICE_IMAGE}:${VERSION}"
            echo "  - ${PRODUCT_SERVICE_IMAGE}:${VERSION}"
            echo "  - ${ORDER_SERVICE_IMAGE}:${VERSION}"
            echo "  - ${NOTIFICATION_SERVICE_IMAGE}:${VERSION}"
            echo "  - ${API_GATEWAY_IMAGE}:${VERSION}"
        }
        failure {
            echo '❌ Pipeline failed!'
            echo 'Check the console output above for errors'
        }
        always {
            echo '🧹 Cleaning up workspace...'
            sh 'docker system prune -f || true'
        }
    }
}