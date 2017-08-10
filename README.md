# Introduction

Usage log generates snapshots of Kubernetes nodes/pods/namespaces/resource quotas and writes them to an output directory. It expects a sidecar container to run and sync these files to S3 where they can be processed.

## Usage




```
PowerBook-G4:usage-log kamador$ ./usage-log -h
This application will log Kubernetes usage metrics to a json log file to be used for accounting and billing processes.

Usage:
  usage-log [flags]

Flags:
  -d, --destinationPath string   Destination path for usage logs (default "logs/")
  -h, --help                     help for usage-log
      --id string                Unique ID to represent the cluster
      --internal                 Running internal to cluster (default true)
      --kubeconfig string        absolute path to the kubeconfig file (default "/Users/kamador/.kube/config")
  -t, --toggle                   Help message for toggle
      --usagePeriod int          Number of seconds per collection interval (default 60)
```