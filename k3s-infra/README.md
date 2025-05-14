# OTel Observability (o11y) Infrastructure

This repository is a derivitive of the
[k3s-ansible](https://github.com/k3s-io/k3s-ansible) collection, with specific
configurations to deploy observability infrastructure using OpenTelemetry
(OTel) and other tools.

The tools deployed into the cluster are:

- Cert-Manager
- OpenTelemetry Operator
- Liatrio OTel Collector (for GitHub metrics w/ exponential backoffs and Gitub App Authentication, to be upstreamed to OTel Collector Contrib)
- The OpenTelemetry Collector Contrib (for GitHub Action Tracing)
- Signoz (for visualization of OTel data) (installed manually through Helm)
  - Clickhouse as a byproduct of the Signoz installation
- CloudflareD (for secure external routing into specific applications)

See [deployment.md](./deployment.md) for specific deployment instructions.

**NOTE:** The below is the original documentation for the k3s-ansible collection.

---

## Supported Platforms
Easily bring up a cluster on machines running:

- [X] Debian
- [X] Ubuntu
- [X] RHEL Family (CentOS, Redhat, Rocky Linux...)
- [X] SUSE Family (SLES, OpenSUSE Leap, Tumbleweed...)
- [X] ArchLinux

on processor architectures:

- [X] x64
- [X] arm64

## System requirements

The control node **must** have Ansible 8.0+ (ansible-core 2.15+)

All managed nodes in inventory must have:
- Passwordless SSH access
- Root access (or a user with equivalent permissions)

It is also recommended that all managed nodes disable firewalls and swap. See
[K3s Requirements](https://docs.k3s.io/installation/requirements) for more
information.

## Usage

First copy the sample inventory to `inventory.yml`.

```bash
cp inventory-sample.yml inventory.yml
```

Second edit the inventory file to match your cluster setup. For example:
```bash
k3s_cluster:
  children:
    server:
      hosts:
        192.16.35.11:
    agent:
      hosts:
        192.16.35.12:
        192.16.35.13:
```

If needed, you can also edit `vars` section at the bottom to match your
environment.

If multiple hosts are in the server group the playbook will automatically setup
k3s in HA mode with embedded etcd. An odd number of server nodes is required
(3,5,7). Read the [official
documentation](https://docs.k3s.io/datastore/ha-embedded) for more information.

Setting up a loadbalancer or VIP beforehand to use as the API endpoint is
possible but not covered here.


Start provisioning of the cluster using the following command:

```bash
ansible-playbook playbooks/site.yml -i inventory.yml --vault-password-file=~/.vault_pass
```

### Using an external database

If an external database is preferred, this can be achieved by passing the
`--datastore-endpoint` as an extra server argument as well as setting the
`use_external_database` flag to true.

```bash
k3s_cluster:
  children:
    server:
      hosts:
        192.16.35.11:
        192.16.35.12:
    agent:
      hosts:
        192.16.35.13:

  vars:
    use_external_database: true
    extra_server_args: "--datastore-endpoint=postgres://username:password@hostname:port/database-name"
```

The `use_external_database` flag is required when more than one server is
defined, as otherwise an embedded etcd cluster will be created instead.

The format of the datastore-endpoint parameter is dependent upon the datastore
backend, please visit the [K3s datastore endpoint
format](https://docs.k3s.io/datastore#datastore-endpoint-format-and-functionality)
for details on the format and supported datastores.

## Upgrading

A playbook is provided to upgrade K3s on all nodes in the cluster. To use it,
update `k3s_version` with the desired version in `inventory.yml` and run:

```bash
ansible-playbook playbooks/upgrade.yml -i inventory.yml
```

## Kubeconfig

After successful bringup, the kubeconfig of the cluster is copied to the control
node and merged with `~/.kube/config` under the `k3s-ansible` context. Assuming
you have [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) installed,
you can confirm access to your **Kubernetes** cluster with the following:

```bash
kubectl config use-context k3s-ansible
kubectl get nodes
```

If you wish for your kubeconfig to be copied elsewhere and not merged, you can
set the `kubeconfig` variable in `inventory.yml` to the desired path.
