# Azure EntraID Apps Monitor

A Golang application that monitors Azure EntraID (formerly Azure AD) application credentials/secrets expiration dates and exports metrics to Prometheus via Pushgateway. The metrics can be visualized using the included Grafana dashboard.

## Features

- Monitors multiple Azure EntraID applications
- Tracks secret/credential expiration dates
- Exports metrics to Prometheus via Pushgateway
- Includes a pre-configured Grafana dashboard
- Runs in a minimal distroless container

## Prerequisites

- Azure EntraID application with the following:
  - Application (client) ID
  - Client secret
  - Microsoft Graph API permissions for reading application data
- Prometheus Pushgateway instance
- Grafana (optional, for visualization)

## Configuration

The application is configured using environment variables:

| Variable | Description | Required |
|----------|-------------|----------|
| CLIENT_ID | Azure EntraID application client ID | Yes |
| CLIENT_SECRET | Azure EntraID application client secret | Yes |
| TENANT_ID | Azure tenant ID | Yes |
| TARGET_IDS | Comma-separated list of application IDs to monitor | Yes |
| PUSHGATEWAY_URL | URL of the Prometheus Pushgateway | Yes |

## Metrics

The application exports the following Prometheus metrics:

- `azad_secret_expiry_timestamp`: Expiration timestamp for Azure application secrets
  - Labels:
    - `application_name`: Display name of the Azure application
    - `application_id`: Application ID
    - `status`: ACTIVE or EXPIRED

## Deployment

### Using Docker

1. Build the container:
```bash
docker build -t azad-app-checker .
```

2. Run the container:
```bash
docker run -e CLIENT_ID=<client_id> \
           -e CLIENT_SECRET=<client_secret> \
           -e TENANT_ID=<tenant_id> \
           -e TARGET_IDS=<app_id1>,<app_id2> \
           -e PUSHGATEWAY_URL=http://pushgateway:9091 \
           azad-app-checker
```

### Using Kubernetes

A sample Kubernetes cronjob manifest is provided in `k8s/cronjob.yaml`. Update secrets and environment variables in the manifest before deploying.

```bash
kubectl apply -f k8s/cronjob.yaml
```

## Grafana Dashboard

Import the provided dashboard JSON from `dashboards/azure-entraid-apps-status.json` into your Grafana instance. The dashboard includes:

- Total number of monitored applications
- Number of expired secrets
- Applications with secrets expiring in the next 30 days
- Detailed status table with expiration dates

## Development

### Requirements

- Go 1.20+
- Access to Azure EntraID
- Prometheus Pushgateway instance

### Building

```bash
go build -o azad-app-checker
```

### Running Locally

```bash
export CLIENT_ID=<client_id>
export CLIENT_SECRET=<client_secret>
export TENANT_ID=<tenant_id>
export TARGET_IDS=<app_id1>,<app_id2>
export PUSHGATEWAY_URL=http://localhost:9091
./azad-app-checker
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
