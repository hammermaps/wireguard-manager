# Development Setup Guide

This document describes the development setup for the WireGuard Manager project, including IDE configuration and CI/CD pipeline.

## JetBrains GoLand / IntelliJ IDEA Setup

The project now includes GoLand/IntelliJ IDEA configuration files in the `.idea` directory. These configurations allow you to easily debug and build the application within the IDE.

### Prerequisites

1. **GoLand 2023.x or newer** (or IntelliJ IDEA Ultimate with Go plugin)
2. **Go 1.23 or newer** installed on your system
3. **Node.js 18 or newer** installed on your system
4. **npm** or **yarn** package manager

### Opening the Project

1. Open GoLand/IntelliJ IDEA
2. Select "Open" and navigate to the project directory
3. The IDE will automatically recognize the `.idea` configuration

### Available Run Configurations

The following run configurations are pre-configured and available in the IDE:

#### 1. Prepare Assets
- **Type**: Shell Script
- **Description**: Prepares frontend assets by installing node modules and copying required files
- **Usage**: Run this first before building or running the application
- **Location**: Run → Run 'Prepare Assets'

#### 2. Build WireGuard Manager
- **Type**: Go Build
- **Description**: Builds the application binary with development flags
- **Auto-runs**: Prepare Assets (before build)
- **Output**: Creates executable in project root
- **Location**: Run → Run 'Build WireGuard Manager'

#### 3. Run WireGuard Manager
- **Type**: Go Application
- **Description**: Runs the application with default environment variables
- **Auto-runs**: Prepare Assets (before running)
- **Default Environment Variables**:
  - `WGUI_ENDPOINT_ADDRESS=127.0.0.1`
  - `WGUI_DNS=1.1.1.1`
  - `WGUI_SERVER_INTERFACE_ADDRESSES=10.252.1.0/24`
  - `WGUI_SERVER_LISTEN_PORT=51820`
- **Location**: Run → Run 'Run WireGuard Manager'

#### 4. Debug WireGuard Manager
- **Type**: Go Application (with debugger)
- **Description**: Runs the application in debug mode with breakpoint support
- **Auto-runs**: Prepare Assets (before debugging)
- **Usage**: Set breakpoints in code, then Run → Debug 'Debug WireGuard Manager'
- **Location**: Run → Debug 'Debug WireGuard Manager'

### Debugging Tips

1. **Set Breakpoints**: Click in the gutter (left margin) next to line numbers to set breakpoints
2. **Step Through Code**: Use F8 (Step Over), F7 (Step Into), Shift+F8 (Step Out)
3. **Evaluate Expressions**: Select code and press Alt+F8 to evaluate
4. **View Variables**: The Variables panel shows all variables in the current scope

### Modifying Environment Variables

To customize environment variables for your local development:

1. Open Run → Edit Configurations
2. Select "Run WireGuard Manager" or "Debug WireGuard Manager"
3. Modify the environment variables in the "Environment variables" section
4. Click OK to save

## Jenkins CI/CD Pipeline

The project includes a `Jenkinsfile` for automated building on Jenkins.

### Pipeline Features

- **Multi-platform builds**: Automatically builds for Linux x64 and Windows x64
- **Parallel execution**: Linux and Windows builds run in parallel for faster builds
- **Asset preparation**: Automatically prepares frontend assets before building
- **Artifact archiving**: Built binaries are archived and available for download
- **Test execution**: Runs Go tests after successful builds

### Build Outputs

The pipeline produces the following artifacts:

1. `wireguard-manager-linux-amd64` - Linux x64 executable
2. `wireguard-manager-windows-amd64.exe` - Windows x64 executable

### Jenkins Setup Requirements

#### Prerequisites

1. **Jenkins 2.x or newer**
2. **Go plugin** for Jenkins
3. **Node.js plugin** for Jenkins (or Node.js installed on agents)
4. **Pipeline plugin** for Jenkins

#### Agent Requirements

##### Linux Agent
- Label: `linux`
- Go 1.23+ installed
- Node.js 18+ installed
- npm or yarn installed

##### Windows Agent (optional)
- Label: `windows`
- Go 1.23+ installed
- Node.js 18+ installed
- npm installed

**Note**: Windows builds can also be cross-compiled on Linux agents.

### Setting up the Pipeline

1. Create a new Pipeline job in Jenkins
2. Configure the SCM to point to this repository
3. Set the Script Path to `Jenkinsfile`
4. Save and run the pipeline

### Pipeline Environment Variables

The pipeline automatically sets the following environment variables:

- `APP_VERSION`: Set to 'stable' for master branch, or the branch name for other branches
- `BUILD_TIME`: Timestamp of the build
- `GIT_COMMIT`: Git commit hash
- `GIT_REF`: Git reference (branch name)

These are embedded in the binary via Go build ldflags.

### Customizing the Pipeline

To customize the Jenkins pipeline:

1. Edit the `Jenkinsfile` in the project root
2. Modify stages, build parameters, or add new build targets
3. Commit and push changes
4. Jenkins will automatically use the updated pipeline on next run

## Manual Building

If you prefer to build manually without IDE or CI/CD:

### Prepare Assets
```bash
chmod +x ./prepare_assets.sh
./prepare_assets.sh
```

### Build for Linux x64
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
  -ldflags="-X 'main.appVersion=dev' -X 'main.buildTime=$(date)' -X 'main.gitCommit=$(git rev-parse HEAD)' -X 'main.gitRef=$(git branch --show-current)'" \
  -o wireguard-manager-linux-amd64 \
  .
```

### Build for Windows x64
```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build \
  -ldflags="-X 'main.appVersion=dev' -X 'main.buildTime=$(date)' -X 'main.gitCommit=$(git rev-parse HEAD)' -X 'main.gitRef=$(git branch --show-current)'" \
  -o wireguard-manager-windows-amd64.exe \
  .
```

## Troubleshooting

### Assets Not Found
If you get errors about missing assets, make sure to run "Prepare Assets" first:
```bash
./prepare_assets.sh
```

### GoLand Doesn't Recognize Go Project
1. Ensure Go SDK is configured: File → Settings → Go → GOROOT
2. Enable Go modules: File → Settings → Go → Go Modules → Enable Go modules integration

### Jenkins Build Fails on Asset Preparation
1. Ensure Node.js is installed on the Jenkins agent
2. Check that npm/yarn is accessible in the PATH
3. Verify network access for downloading node modules

## Additional Resources

- [WireGuard Manager Documentation](README.md)
- [API Documentation](API_DOCUMENTATION.md)
- [Contributing Guidelines](CONTRIBUTING.md)
