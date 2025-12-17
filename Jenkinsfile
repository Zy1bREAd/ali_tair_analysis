pipeline {
    agent any

    // 定义环境变量
    environment {
        // 例如设置项目相关的变量
        PROJECT_NAME = "OceanWang"
        DOCKER_IMAGE_NAME = "xdemoapp"
        DOCKER_IMAGE_TAG = "main"
        HARBOR_URL = "124.220.17.5:8018"
        HARBOR_PROJECT = "xdemo"
        DEVELOP_SERVER_IP = "10.0.20.5"
        DEVELOP_SERVER_USER = "ubuntu"
        DEVELOP_SERVER_CRED_ID = "ssh-for-password-10.0.20.5"
    }

    // 触发构建的条件，这里是当 GitHub 仓库有推送（push）事件时触发
    // triggers {
    //     githubPush()
    // }

    // 构建步骤
    stages {
        stage('Checkout GitHub Branch and Pull Code') {
            steps {
                // 从 GitHub 仓库检出代码
                checkout([$class: 'GitSCM', 
                        branches: [[name: '*/main']], 
                        userRemoteConfigs: [[url: 'https://github.com/Zy1bREAd/ali_tair_analysis.git']]])
            }
        }
        stage('Build On Image For Develop') {
            when {
                branch 'main'
            }
            steps {
                sh "echo 'build image'"
            }
        }
        stage('Build On Image For Production') {
            steps {
                sh "echo 'prod' && env"
            }
        }
    }

    // 构建后操作，如发送通知等
    // post {
    //     success {
    //         // 构建成功时发送邮件通知等操作，需要配置 Jenkins 的邮件插件等相关设置
    //         emailext subject: 'Build Success: ${PROJECT_NAME}', 
    //                 body: 'The build of ${PROJECT_NAME} was successful.', 
    //                 to: 'your-email@example.com'
    //     }
    //     failure {
    //         // 构建失败时发送邮件通知等操作
    //         emailext subject: 'Build Failure: ${PROJECT_NAME}', 
    //                 body: 'The build of ${PROJECT_NAME} has failed.', 
    //                 to: 'your-email@example.com'
    //     }
    // }
}