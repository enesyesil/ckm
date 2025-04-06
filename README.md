# ☁️ CKM: Cloud Kernel Manager

> A modular kernel simulation framework for managing both simulated processes and virtual machines using real OS principles — with full DevOps-style observability.
>
> ## What Is CKM?

**Cloud Kernel Manager (CKM)** is a hybrid cloud-native project that:

- Simulates OS-level behavior (scheduling, memory, sync)
- Orchestrates both simulated workloads and virtual machines
- Provides Prometheus-style metrics for resource observability
- Is modular, pluggable, and DevOps-ready

## Features

| Component          | Description                                                    |
|--------------------|----------------------------------------------------------------|
| CPU Scheduling      | Pluggable strategies like FIFO, Round Robin, and Priority     |
| Memory Management   | Simulated allocation, overcommit tracking, paging (planned)   |
| Synchronization     | Mutex, semaphore simulation, deadlock detection               |
| Workload Types      | Unified struct for both simulated tasks and VMs               |
| Metrics             | Prometheus `/metrics` endpoint for real-time system stats     |
| VM Integration      | Optional: QEMU or Firecracker backend                         |
| YAML Configuration  | Workload and VM specs defined via config files                |

## Architecture and Design

![Design](./arch.png)
