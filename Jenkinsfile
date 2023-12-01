@Library('ckt-libraries')
import static git.Github.*
import static checks.SonarQube.*
import static docker.Docker.*

node('sre-slave') {

    ansiColor('xterm') {
        final String REGEX_FILTER = '^((false (opened|reopened|synchronize))|(true (closed)))? (develop|master|main)?$'
        triggers.setGenericTriggerConfiguration(this, repo_name, REGEX_FILTER)

        final int SONAR_TIMEOUT_MINUTES = 5
        final String CI_DOCKER_TARGET_STAGE = "builder"
        final String DEFAULT_DOCKERFILE = "Dockerfile"
        final List vulnerabilities_type = ["CRITICAL"]

        String target_branch = action == "closed" ? base_branch : head_branch
        currentBuild.displayName = "${BUILD_NUMBER} - ${repo_name} - ${target_branch}"

        try {
            notifyPRStatus(this, env.GITHUB_TOKEN_CREDENTIAL_ID, "pending", statuses_url, "${BUILD_URL}")

            stage("Checkout") {
                git(url: clone_url, branch: target_branch, credentialsId: env.GITHUB_SSH_CREDENTIAL_ID)
            }

            stage('Docker Build') {
                buildDockerImage(this, repo_name, env.GITHUB_SSH_CREDENTIAL_ID, DEFAULT_DOCKERFILE, CI_DOCKER_TARGET_STAGE)
            }
            stage('linter') {
                 runCommand(this, repo_name, " make lint-install lint")
            }
            stage('Unit Test') {
                runUnitTest(this, repo_name, "go test ./... -coverprofile cover.out", "/go/src/${repo_name}")
            }
            stage('SonarQube analysis') {
                scanRepo(this, repo_name, target_branch, SONAR_TIMEOUT_MINUTES)
            }

            notifyPRStatus(this, env.GITHUB_TOKEN_CREDENTIAL_ID, "success", statuses_url, "${BUILD_URL}")
        } catch (Exception e) {
            currentBuild.result = 'FAILURE'
            echo "Exception: ${e}"
            notifyPRStatus(this, env.GITHUB_TOKEN_CREDENTIAL_ID, "failure", statuses_url, "${BUILD_URL}")
        } finally {
            downDockerCompose(this, "docker-compose.ci.yml", BUILD_NUMBER)
            cleanWs()
        }
    }
}
