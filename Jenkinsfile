pipeline {

    agent {
       label 'ondemand'
    }

    options {
          timeout(time: 15, unit: 'MINUTES')
    }

    stages {
        stage('Gerrit status') {
            steps {
                withCredentials([sshUserPrivateKey(credentialsId: 'gerrit-trigger-ssh', keyFileVariable: 'SSH_KEY', usernameVariable: 'SSH_USER')]) {
                    sh './scripts/gerrit-status.sh verify start 0'
                }
            }
        }
        stage('Checkout') {
            steps {
               checkout scm
            }
        }
        stage('Build') {
            steps {
                 sh 'docker buildx bake build'
            }
        }

        stage('Test') {
            steps {
                sh 'docker buildx bake integration-test'
            }
        }
    }
    post {
        success {
            withCredentials([sshUserPrivateKey(credentialsId: 'gerrit-trigger-ssh', keyFileVariable: 'SSH_KEY', usernameVariable: 'SSH_USER')]) {
                sh './scripts/gerrit-status.sh verify success +1'
            }
        }
        failure {
            withCredentials([sshUserPrivateKey(credentialsId: 'gerrit-trigger-ssh', keyFileVariable: 'SSH_KEY', usernameVariable: 'SSH_USER')]) {
                sh './scripts/gerrit-status.sh verify failure -1'
            }
        }
        aborted {
            withCredentials([sshUserPrivateKey(credentialsId: 'gerrit-trigger-ssh', keyFileVariable: 'SSH_KEY', usernameVariable: 'SSH_USER')]) {
                sh './scripts/gerrit-status.sh verify failure -1'
            }
        }
    }
}
