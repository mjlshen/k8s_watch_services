# k8s_watch_services

### Build Instructions
`go build main.go`

### What this does
Using a provided kubeconfig, watches for changes to NodePort services on a cluster. If one is added, it will run an Ansible Playbook to simulate further tasks. On first launch, it also runs the Ansible Playbook on existing NodePort services.

![Demo](https://github.com/mjlshen/k8s_watch_services/raw/master/demo.gif "Demo")

### Usage
`./main --kubeconfig=$FULL_PATH_TO_KUBECONFIG`
