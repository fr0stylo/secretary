# Secretary

![CodeRabbit Pull Request Reviews](https://img.shields.io/coderabbit/prs/github/fr0stylo/secretary?utm_source=oss&utm_medium=github&utm_campaign=fr0stylo%2Fsecretary&labelColor=171717&color=FF570A&link=https%3A%2F%2Fcoderabbit.ai&label=CodeRabbit+Reviews)

Secretary is a lightweight utility that securely fetches secrets from multiple secret management providers and makes them available to your application as files. Currently supports AWS Secrets Manager with upcoming support for AWS Systems Manager Parameter Store, Google Cloud Secret Manager, and HashiCorp Vault.

## Features

- **Multi-provider support**: Currently AWS Secrets Manager, with more providers coming soon
- **Secure file storage**: Secrets stored as files in `/tmp` directory with restricted permissions
- **Environment variable mapping**: Secrets accessible via environment variables pointing to file paths
- **Real-time monitoring**: Automatic detection and handling of secret changes
- **Signal management**: Proper signal handling for graceful process management
- **Process wrapper**: Acts as a lightweight wrapper around your application

## Supported Providers

### Currently Available
- **AWS Secrets Manager**: Full support with automatic rotation detection

### Coming Soon
- **AWS Systems Manager Parameter Store**: Hierarchical parameter management
- **Google Cloud Secret Manager**: Native GCP secrets integration
- **HashiCorp Vault**: Enterprise-grade secret management

## Installation

### Using Go

```bash
go install github.com/fr0stylo/secretary@latest
```

### Using GitHub Container Registry

```bash
docker pull ghcr.io/fr0stylo/secretory:latest
```

### From Source

```bash
git clone https://github.com/fr0stylo/secretary.git
cd secretary
go build -o secretary
```

## Usage

Secretary acts as a wrapper for your application, fetching secrets from configured providers and then running your application with those secrets available as files.

### AWS Secrets Manager

#### Basic Usage

```bash
SECRETARY_DB_PASSWORD=arn:aws:secretsmanager:us-west-2:123456789012:secret:prod/db/password-AbCdEf \
SECRETARY_API_KEY=arn:aws:secretsmanager:us-west-2:123456789012:secret:prod/api-key-GhIjKl \
secretary your-application [args...]
```

In this example:
- `SECRETARY_DB_PASSWORD` tells Secretary to fetch a secret and make it available as `DB_PASSWORD`
- The secret will be stored in `/tmp/DB_PASSWORD`
- Your application will receive `DB_PASSWORD=/tmp/DB_PASSWORD` as an environment variable
- The value is the full ARN of the secret in AWS Secrets Manager

#### Docker Example

```dockerfile
FROM ghcr.io/fr0stylo/secretory:latest

# Copy your application
COPY your-application .

# AWS configuration (prefer IAM roles in production)
ENV AWS_REGION=us-west-2

# Define secrets to fetch
ENV SECRETARY_DB_PASSWORD=arn:aws:secretsmanager:us-west-2:123456789012:secret:prod/db/password-AbCdEf
ENV SECRETARY_API_KEY=arn:aws:secretsmanager:us-west-2:123456789012:secret:prod/api-key-GhIjKl
ENV SECRETARY_JWT_SECRET=arn:aws:secretsmanager:us-west-2:123456789012:secret:prod/jwt-secret-MnOpQr

# Run your application with Secretary
ENTRYPOINT ["./secretary", "./your-application"]
```

#### Kubernetes with IAM Roles for Service Accounts (IRSA)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: your-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: your-app
  template:
    metadata:
      labels:
        app: your-app
    spec:
      serviceAccountName: your-app-service-account
      containers:
      - name: your-app
        image: your-app:latest
        env:
        - name: AWS_REGION
          value: "us-west-2"
        - name: SECRETARY_DB_PASSWORD
          value: "arn:aws:secretsmanager:us-west-2:123456789012:secret:prod/db/password-AbCdEf"
        - name: SECRETARY_API_KEY  
          value: "arn:aws:secretsmanager:us-west-2:123456789012:secret:prod/api-key-GhIjKl"
        command: ["./secretary"]
        args: ["./your-application"]
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
```

### Future Provider Examples

#### AWS Systems Manager Parameter Store (Coming Soon)

```bash
# Hierarchical parameters
SECRETARY_DB_CONFIG=ssm:///myapp/prod/database/config \
SECRETARY_API_KEYS=ssm:///myapp/prod/api/keys \
secretary your-application
```

#### Google Cloud Secret Manager (Coming Soon)

```bash
# GCP Secret Manager
SECRETARY_DB_PASSWORD=gcp://projects/my-project/secrets/db-password/versions/latest \
SECRETARY_SERVICE_ACCOUNT=gcp://projects/my-project/secrets/service-account/versions/1 \
secretary your-application
```

#### HashiCorp Vault (Coming Soon)

```bash
# Vault KV v2
SECRETARY_DB_CREDS=vault://secret/data/myapp/database \
SECRETARY_API_TOKEN=vault://auth/approle/login \
secretary your-application
```

## Configuration

### Environment Variable Format

Secretary uses environment variables prefixed with `SECRETARY_` to determine which secrets to fetch:

- **Format**: `SECRETARY_<SECRET_NAME>=<provider_specific_identifier>`
- **File location**: Secrets are stored in `/tmp/<SECRET_NAME>`
- **Environment variable**: Your application receives `<SECRET_NAME>=/tmp/<SECRET_NAME>`
- **Permissions**: Secret files are created with `0600` permissions (owner read/write only)

### Provider Selection

The provider is automatically determined by the secret identifier format:

- **AWS Secrets Manager**: `arn:aws:secretsmanager:...`
- **AWS SSM Parameter Store**: `ssm://...` (coming soon)
- **Google Cloud Secret Manager**: `gcp://...` (coming soon)
- **HashiCorp Vault**: `vault://...` (coming soon)

### Monitoring and Rotation

Secretary continuously monitors secrets for changes:

- **Check frequency**: Every 15 seconds (configurable)
- **Change detection**: Version/revision comparison
- **Application notification**: Sends `SIGHUP` to your application on secret changes
- **Automatic reload**: Secrets are automatically rewritten to files when changed

## Provider-Specific Configuration

### AWS Secrets Manager

**Required Permissions**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue",
        "secretsmanager:DescribeSecret"
      ],
      "Resource": "arn:aws:secretsmanager:*:*:secret:*"
    }
  ]
}
```

**Credential Sources** (in order of precedence):
1. Environment variables: `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`
2. AWS credentials file (`~/.aws/credentials`)
3. IAM roles for EC2/ECS/Lambda
4. IAM Roles for Service Accounts (IRSA) in Kubernetes
5. AWS IAM Identity Center (SSO)

### AWS Systems Manager Parameter Store (Coming Soon)

**Required Permissions**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm:GetParameter",
        "ssm:GetParameters",
        "ssm:GetParametersByPath"
      ],
      "Resource": "arn:aws:ssm:*:*:parameter/*"
    }
  ]
}
```

### Google Cloud Secret Manager (Coming Soon)

**Required Permissions**:
- `secretmanager.versions.access`
- `secretmanager.secrets.get`

**Authentication**:
- Service Account Keys
- Workload Identity (GKE)
- Application Default Credentials

### HashiCorp Vault (Coming Soon)

**Authentication Methods**:
- Token authentication
- AppRole authentication
- Kubernetes service account tokens
- Cloud provider authentication (AWS, GCP, Azure)

## Advanced Usage

### Custom Configuration

```bash
# Custom check frequency (30 seconds)
SECRETARY_CHECK_FREQUENCY=30s \
SECRETARY_DB_PASSWORD=arn:aws:secretsmanager:us-west-2:123456789012:secret:db-password-AbCdEf \
secretary your-application
```

### Multiple Providers

```bash
# Mix different providers (future capability)
SECRETARY_AWS_SECRET=arn:aws:secretsmanager:us-west-2:123456789012:secret:app-secret-AbCdEf \
SECRETARY_VAULT_TOKEN=vault://secret/data/myapp/token \
SECRETARY_GCP_KEY=gcp://projects/my-project/secrets/api-key/versions/latest \
secretary your-application
```

### Health Checking

Secretary provides health check capabilities for orchestration platforms:

```bash
# Check if Secretary is running and secrets are loaded
secretary --health-check
```

## Deployment Examples

### Docker Compose

```yaml
version: '3.8'
services:
  app:
    image: your-app:latest
    environment:
      - AWS_REGION=us-west-2
      - SECRETARY_DB_PASSWORD=arn:aws:secretsmanager:us-west-2:123456789012:secret:db-password-AbCdEf
      - SECRETARY_API_KEY=arn:aws:secretsmanager:us-west-2:123456789012:secret:api-key-GhIjKl
    command: ["./secretary", "./your-application"]
    volumes:
      - /tmp:/tmp
```

### AWS ECS Task Definition

```json
{
  "family": "your-app",
  "taskRoleArn": "arn:aws:iam::123456789012:role/your-app-task-role",
  "containerDefinitions": [
    {
      "name": "your-app",
      "image": "your-app:latest",
      "environment": [
        {"name": "AWS_REGION", "value": "us-west-2"},
        {"name": "SECRETARY_DB_PASSWORD", "value": "arn:aws:secretsmanager:us-west-2:123456789012:secret:db-password-AbCdEf"}
      ],
      "entryPoint": ["./secretary"],
      "command": ["./your-application"]
    }
  ]
}
```

### AWS Lambda Custom Runtime

```bash
#!/bin/bash
# bootstrap file
export SECRETARY_DB_PASSWORD=arn:aws:secretsmanager:us-west-2:123456789012:secret:db-password-AbCdEf
exec ./secretary ./lambda-handler
```

## Troubleshooting

### Common Issues

1. **Permission denied errors**: Ensure your IAM role/user has the required permissions for the secret provider
2. **Secret not found**: Verify the secret identifier format and that the secret exists
3. **File permission issues**: Secretary creates files with `0600` permissions; ensure your application can read them
4. **Signal handling**: If your application doesn't handle `SIGHUP`, secret rotation notifications won't work

### Debug Mode

```bash
SECRETARY_DEBUG=true \
SECRETARY_DB_PASSWORD=arn:aws:secretsmanager:us-west-2:123456789012:secret:db-password-AbCdEf \
secretary your-application
```

### Logging

Secretary logs important events:
- Secret retrieval and updates
- Version changes detected
- Signal forwarding to child processes
- Error conditions

## Security Considerations

- **File Permissions**: Secret files are created with `0600` permissions (owner only)
- **Temporary Storage**: Secrets are stored in `/tmp` which should be mounted as `tmpfs`
- **Memory**: Secrets are not stored in environment variables, reducing exposure
- **Process Isolation**: Secretary runs as a separate process from your application
- **Credential Rotation**: Automatic handling of secret rotation without application restart

## Roadmap

### Version 2.0 (Q2 2025)
- AWS Systems Manager Parameter Store support
- Google Cloud Secret Manager integration
- Enhanced configuration options
- Health check endpoints

### Version 2.5 (Q3 2025)
- HashiCorp Vault integration
- Azure Key Vault support
- Plugin architecture for custom providers
- Metrics and monitoring endpoints

### Version 3.0 (Q4 2025)
- Multi-region secret replication
- Secret caching and offline mode  
- Configuration file support
- Advanced secret transformation features

## Contributing

Contributions are welcome! Please see our contributing guidelines for details on how to submit pull requests, report issues, and suggest improvements.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

Zymantas Maumevicius
