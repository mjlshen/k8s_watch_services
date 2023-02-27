package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	ansibler "github.com/apenella/go-ansible"
)

func onAdd(obj interface{}) {
	var (
		nodePort       int32
		deploymentName string
	)

	svc := obj.(*corev1.Service)
	nodePortSvc := false

	for key, element := range svc.GetLabels() {
		if key == "run" {
			deploymentName = element
			nodePortSvc = true
			break
		}
	}

	for _, p := range svc.Spec.Ports {
		if p.NodePort != 0 {
			nodePort = p.NodePort
			nodePortSvc = true
			break
		}
		nodePortSvc = false
	}

	if nodePortSvc {
		ansiblePlaybookConnectionOptions := &ansibler.AnsiblePlaybookConnectionOptions{
			Connection: "local",
		}

		ansiblePlaybookOptions := &ansibler.AnsiblePlaybookOptions{
			Inventory: "127.0.0.1,",
		}

		ansiblePlaybookOptions.AddExtraVar("deployment_name", deploymentName)
		ansiblePlaybookOptions.AddExtraVar("nodePort", nodePort)
		ansiblePlaybookOptions.AddExtraVar("cluster", "sds-dev")

		playbook := &ansibler.AnsiblePlaybookCmd{
			Playbook:          "create-ns-services.yml",
			ConnectionOptions: ansiblePlaybookConnectionOptions,
			Options:           ansiblePlaybookOptions,
		}

		err := playbook.Run()
		if err != nil {
			panic(err)
		}
	}
	// fmt.Println(deploymentName + "-" + strconv.FormatInt(int64(nodePort), 10))
}

func onDelete(obj interface{}) {
	svc := obj.(*corev1.Service)
	s := fmt.Sprintf("Deleted %s\n\n", svc.GetName())
	fmt.Println(s)
}

func onUpdate(oldObj, newObj interface{}) {
	oldSvc := oldObj.(*corev1.Service)
	newSvc := newObj.(*corev1.Service)
	s := fmt.Sprintf("%s Updated to %s\n\n", oldSvc.GetName(), newSvc.GetUID())
	fmt.Println(s)
}

func main() {
	var (
		kubeconfig = flag.String("kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "(OPTIONAL) absolute path to kubeconfig")
	)

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		log.Panic(err.Error())
	}

	// getWorkerNodeNames(clientset)

	factory := informers.NewSharedInformerFactory(clientset, 0)
	svcInformer := factory.Core().V1().Services().Informer()

	svcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		DeleteFunc: onDelete,
		UpdateFunc: onUpdate,
	})

	stop := make(chan struct{})
	defer close(stop)

	svcInformer.Run(stop)
}
