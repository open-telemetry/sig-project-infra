# OTel Project Telemetry

**NOTE:** Most, if not all, of this is temporary and will be completed
replaced, including being put on different domains.

## Setup Steps

### Initial Setup & Connection to VM

1. Click button deploy a server
2. Get the ssh key
3. configure `inventory.yaml`
4. Run `ansible-playbook playbooks/site.yml -i inventory.yml
   --vault-password-file=./.vault_pass` (requires that you have the password
   for vault)
5. Use `ssh` to connect to the VM

You now can begin to configure and setup the manifests within the cluster.

### Configure K3S Cluster (assume previous steps)

1. The following manifests will get installed through the Ansible Playbook
    1. cert manager
    2. otel operator
    3. otel collectors (tracing and metrics)
    4. Cloudflare D
        1. https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/get-started/create-remote-tunnel/
2. Created a WAF rule in Cloudflare only allowing traffic from hooks to otel.adrielperkins.dev/*
7. Added a [viz.adrielperkins.dev](http://viz.adrielperkins.dev) route
8. `helm -n signoz upgrade -i signoz signoz/signoz -f signoz-values.yaml
   --create-namespace` was manually run on the server due to helm chart
   templating issues.
