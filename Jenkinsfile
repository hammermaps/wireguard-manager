pipeline {
    agent any
    
    environment {
        GO_VERSION = '1.23'
        APP_VERSION = "${env.GIT_BRANCH == 'origin/master' ? 'stable' : env.GIT_BRANCH}"
        BUILD_TIME = sh(script: 'date', returnStdout: true).trim()
        GIT_COMMIT = "${env.GIT_COMMIT}"
        GIT_REF = "${env.GIT_BRANCH}"
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        
        stage('Setup Node.js') {
            steps {
                script {
                    // Install Node.js if not available
                    sh '''
                        if ! command -v node &> /dev/null; then
                            echo "Node.js not found, please ensure Node.js 18+ is installed"
                            exit 1
                        fi
                        node --version
                        npm --version
                    '''
                }
            }
        }
        
        stage('Prepare Assets') {
            steps {
                sh '''
                    chmod +x ./prepare_assets.sh
                    ./prepare_assets.sh
                '''
            }
        }
        
        stage('Build') {
            parallel {
                stage('Build Linux x64') {
                    agent {
                        label 'linux'
                    }
                    environment {
                        GOOS = 'linux'
                        GOARCH = 'amd64'
                        CGO_ENABLED = '0'
                    }
                    steps {
                        sh '''
                            go version
                            go mod download
                            go build -v \
                                -ldflags="-X 'main.appVersion=${APP_VERSION}' -X 'main.buildTime=${BUILD_TIME}' -X 'main.gitCommit=${GIT_COMMIT}' -X 'main.gitRef=${GIT_REF}'" \
                                -o wireguard-manager-linux-amd64 \
                                .
                        '''
                    }
                    post {
                        success {
                            archiveArtifacts artifacts: 'wireguard-manager-linux-amd64', fingerprint: true
                        }
                    }
                }
                
                stage('Build Windows x64') {
                    agent {
                        label 'windows || linux'
                    }
                    environment {
                        GOOS = 'windows'
                        GOARCH = 'amd64'
                        CGO_ENABLED = '0'
                    }
                    steps {
                        script {
                            if (isUnix()) {
                                sh '''
                                    go version
                                    go mod download
                                    go build -v \
                                        -ldflags="-X 'main.appVersion=${APP_VERSION}' -X 'main.buildTime=${BUILD_TIME}' -X 'main.gitCommit=${GIT_COMMIT}' -X 'main.gitRef=${GIT_REF}'" \
                                        -o wireguard-manager-windows-amd64.exe \
                                        .
                                '''
                            } else {
                                bat '''
                                    go version
                                    go mod download
                                    go build -v ^
                                        -ldflags="-X 'main.appVersion=%APP_VERSION%' -X 'main.buildTime=%BUILD_TIME%' -X 'main.gitCommit=%GIT_COMMIT%' -X 'main.gitRef=%GIT_REF%'" ^
                                        -o wireguard-manager-windows-amd64.exe ^
                                        .
                                '''
                            }
                        }
                    }
                    post {
                        success {
                            archiveArtifacts artifacts: 'wireguard-manager-windows-amd64.exe', fingerprint: true
                        }
                    }
                }
            }
        }
        
        stage('Test') {
            steps {
                sh '''
                    go test -v ./...
                '''
            }
        }
    }
    
    post {
        success {
            echo 'Build completed successfully!'
        }
        failure {
            echo 'Build failed!'
        }
        always {
            cleanWs()
        }
    }
}
