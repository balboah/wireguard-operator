# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a WireGuard operator that exposes WireGuard configuration through HTTP JSON API, supporting adding, removing, and replacing peers with automatic IP allocation. The operator manages WireGuard interfaces and provides peer management capabilities.

## Build and Development Commands

### Building and Testing
- `make local_docker` - Build multi-platform Docker images locally (linux/arm64/v8, linux/amd64)
- `make docker` - Build and push Docker image using Google Cloud Build with current git version tag
- `make integration_test` - Run integration tests on remote test host (requires TEST_HOST environment variable)
- `make clean` - Remove build artifacts (test binaries)

### Testing Strategy
- Unit tests: `go test ./...` for all unit tests
- Integration tests: `make integration_test` (requires remote Linux host for WireGuard functionality)
- Integration tests are tagged with `//go:build integration` and test real WireGuard interface operations

### Deployment
- `make init` - Initialize Kubernetes resources (requires manifests/.secrets.yaml)
- `make deploy` - Deploy to Kubernetes cluster (requires manifests/.secrets.yaml)

## Architecture

### Core Components

**WgClient** (`client.go`): Wrapper around wgctrl library that binds WireGuard configuration to a specific interface link. Handles private key management and device configuration.

**WgLink** (`link.go`): Manages WireGuard network interface creation and configuration using netlink. Creates new interfaces or uses existing ones.

**IPPool** (`ippool.go`): Thread-safe IP address allocation system for managing peer IP assignments within a CIDR range. Handles allocation, deallocation, and cleanup of IP addresses.

**PeerHandler** (`peer_handler.go`): HTTP handler for peer management operations (PUT/DELETE/GET) that coordinates between WireGuard configuration, IP allocation, and peer lifecycle management.

**Main Application** (`cmd/wgo/main.go`): Entry point that sets up the WireGuard interface, configures networking, starts HTTP server with peer management endpoints and Prometheus metrics.

### Key Interfaces

- `WgDeviceConfigurator`: Abstraction for WireGuard device configuration and peer management
- `WgIdentity`: Represents WireGuard identity information (public key, port)
- `IPPool`: Interface for IP address allocation and management

### API Endpoints

- `/v1/peer` - Peer management (PUT to add/update, DELETE to remove, GET to list)
- `/v1/id` - Get operator identity information
- `/metrics` - Prometheus metrics endpoint

### Configuration

The operator accepts various command-line flags for:
- WireGuard interface configuration (interface name, port, private key)
- Network configuration (IPv4/IPv6 CIDR ranges)
- TLS/mTLS setup for secure API access
- Logging configuration

### Integration Testing Requirements

Integration tests require a Linux host with WireGuard capabilities since they test actual network interface operations. Tests are designed to run on a remote host specified by TEST_HOST environment variable.