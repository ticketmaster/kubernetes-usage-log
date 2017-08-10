package catalog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"log"
	"time"

	coreapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // imported to allow connection to tectonic clusters
)

// GenerateCatalog creates a catalog of nodes and namespaces within the clientset cluster
func GenerateCatalog(clientset *kubernetes.Clientset, stopCh chan struct{}, clusterID string, period int64, path string) {
	defer runtime.HandleCrash()

	log.Println("Starting catalog generation")
	lastTime := time.Now().UTC()
	for {

		results := auditCluster(clientset, clusterID)

		currentTime := time.Now().UTC()

		results.Time = currentTime.Format(time.RFC3339)
		results.Duration = currentTime.Sub(lastTime).Seconds()

		periodBegin := time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())
		periodEnd := periodBegin.AddDate(0, 1, 0)

		periodPath := fmt.Sprintf("period=%s-%s", periodBegin.Format("20060102"), periodEnd.Format("20060102"))
		filename := strings.Replace(fmt.Sprintf("%s---%s.json", clusterID, currentTime), " ", "_", -1)

		outputPath := filepath.Join(path, periodPath)
		err := os.MkdirAll(outputPath, os.FileMode(0755))
		if err != nil {
			log.Println("error:", err)
		}

		outputFile := filepath.Join(outputPath, filename)

		output, err := json.Marshal(results)
		if err != nil {
			log.Println("error:", err)
		}

		log.Printf("Writing to %s", outputFile)
		err = ioutil.WriteFile(outputFile, []byte(output), os.FileMode(0755))
		if err != nil {
			log.Println("error:", err)
		}

		lastTime = currentTime
		time.Sleep(time.Duration(period) * time.Second)
	}

	<-stopCh
	log.Println("Stopping catalog generation")
}

func auditCluster(clientset *kubernetes.Clientset, clusterID string) *AuditResults {
	results := NewAuditResults()

	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		results.ErrorMessage = "Unable to return list of namespaces"
		return results
	}

	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		results.ErrorMessage = "Unable to return list of pods"
		return results
	}

	resourceQuotas, err := clientset.CoreV1().ResourceQuotas("").List(metav1.ListOptions{})
	if err != nil {
		results.ErrorMessage = "Unable to return list of resource quotas"
		return results
	}

	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		results.ErrorMessage = "Unable to return list of nodes"
		return results
	}

	results = generateAuditResults(namespaces, pods, resourceQuotas, nodes)
	results.ClusterID = clusterID

	return results
}

func generateAuditResults(namespaces *coreapi.NamespaceList, pods *coreapi.PodList, resourceQuotas *coreapi.ResourceQuotaList, nodes *coreapi.NodeList) *AuditResults {
	results := NewAuditResults()

	for _, each := range namespaces.Items {
		namespace := NewNamespace(each.Name)
		namespace.Annotations = each.Annotations
		namespace.Labels = each.Labels

		results.Namespaces = append(results.Namespaces, namespace)
	}

	for _, p := range pods.Items {
		pod := NewPod(p.Name)
		pod.Annotations = p.Annotations
		pod.Labels = p.Labels
		pod.NodeName = p.Spec.NodeName
		pod.Namespace = p.Namespace
		pod.UID = string(p.UID)

		for _, o := range p.OwnerReferences {
			owner := Owner{
				ApiVersion:         o.APIVersion,
				BlockOwnerDeletion: o.BlockOwnerDeletion,
				Controller:         o.Controller,
				Kind:               o.Kind,
				Name:               o.Name,
			}
			pod.Owners = append(pod.Owners, owner)
		}

		for _, c := range p.Spec.Containers {
			container := Container{
				Name:  c.Name,
				Image: c.Image,
				Resources: ResourceValues{
					CPULimits:      c.Resources.Limits.Cpu().Value(),
					MemoryLimits:   c.Resources.Limits.Memory().Value(),
					CPURequests:    c.Resources.Requests.Cpu().Value(),
					MemoryRequests: c.Resources.Requests.Memory().Value(),
				},
				OwnerUID: string(p.UID),
			}

			pod.Containers = append(pod.Containers, container)
		}

		n := results.Namespaces.Find(pod.Namespace)
		n.Pods = append(n.Pods, *pod)
	}

	for _, q := range resourceQuotas.Items {
		quota := NewResourceQuota(q.Name)
		quota.Namespace = q.Namespace
		quota.SpecHardResources = extractResourceQuotas("spec.hard", &q.Spec.Hard)
		quota.StatusHardResources = extractResourceQuotas("status.hard", &q.Status.Hard)
		quota.StatusUsedResources = extractResourceQuotas("status.used", &q.Status.Used)

		n := results.Namespaces.Find(quota.Namespace)
		n.ResourceQuotas = append(n.ResourceQuotas, *quota)
	}

	for _, each := range nodes.Items {
		node := Node{
			Name:       each.Name,
			ExternalID: each.Spec.ExternalID,
			Labels:     each.Labels,

			Allocatable: ResourceValues{
				CPULimits:    each.Status.Allocatable.Cpu().Value(),
				MemoryLimits: each.Status.Allocatable.Memory().Value(),
			},

			Capacity: ResourceValues{
				CPULimits:    each.Status.Capacity.Cpu().Value(),
				MemoryLimits: each.Status.Capacity.Memory().Value(),
			},
		}

		for _, address := range each.Status.Addresses {
			if address.Type == "InternalIP" {
				node.InternalIP = address.Address
			} else if address.Type == "ExternalIP" {
				node.ExternalIP = address.Address
			}
		}

		for _, condition := range each.Status.Conditions {
			if condition.Type == "Ready" {
				node.Status = string(condition.Status)
			}
		}

		results.Nodes = append(results.Nodes, node)
	}

	return results
}

func extractResourceQuotas(name string, list *coreapi.ResourceList) ResourceValues {
	return ResourceValues{
		Name:           name,
		CPULimits:      extractResourceQuota(list, "limits.cpu"),
		MemoryLimits:   extractResourceQuota(list, "limits.memory"),
		CPURequests:    extractResourceQuota(list, "requests.cpu"),
		MemoryRequests: extractResourceQuota(list, "requests.memory"),
	}
}

// Returns the resource  if specified.
func extractResourceQuota(list *coreapi.ResourceList, resource coreapi.ResourceName) int64 {
	if val, ok := (*list)[resource]; ok {
		return val.Value()
	}
	return 0
}
