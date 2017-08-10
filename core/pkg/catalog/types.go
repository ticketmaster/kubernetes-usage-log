package catalog

type AuditResults struct {
	ClusterID    string     `json:"clusterId"`
	Time         string     `json:"time"`
	Duration     float64    `json:"duration"`
	Nodes        []Node     `json:"nodes,omitempty"`
	Namespaces   Namespaces `json:"namespaces,omitempty"`
	Success      bool       `json:"success"`
	ErrorMessage string     `json:"errorMessage"`
}

func NewAuditResults() *AuditResults {
	r := new(AuditResults)
	r.Namespaces = make(Namespaces, 0)
	return r
}

type Namespace struct {
	Name           string            `json:"name"`
	Annotations    map[string]string `json:"annotations,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	ResourceQuotas []ResourceQuota   `json:"resourceQuotas,omitempty"`
	Pods           []Pod             `json:"pods,omitempty"`
}

type Namespaces []*Namespace

func (n Namespaces) Find(s string) *Namespace {
	for _, namespace := range n {
		if namespace.Name == s {
			return namespace
		}
	}
	return nil
}

func NewNamespace(name string) *Namespace {
	r := new(Namespace)
	r.Name = name
	return r
}

type ResourceQuota struct {
	Name                string         `json:"name"`
	Namespace           string         `json:"namespace"`
	SpecHardResources   ResourceValues `json:"specHardResources,omitempty"`
	StatusHardResources ResourceValues `json:"statusHardResources,omitempty"`
	StatusUsedResources ResourceValues `json:"statusUsedResources,omitempty"`
}

func NewResourceQuota(name string) *ResourceQuota {
	r := new(ResourceQuota)
	r.Name = name
	return r
}

type Node struct {
	Name        string            `json:"name"`
	ExternalID  string            `json:"externalId"`
	InternalIP  string            `json:"internalIp"`
	ExternalIP  string            `json:"externalIp"`
	Allocatable ResourceValues    `json:"allocatable,omitempty"`
	Capacity    ResourceValues    `json:"capacity,omitempty"`
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type ResourceValues struct {
	Name           string `json:"name"`
	CPULimits      int64  `json:"cpuLimits"`
	MemoryLimits   int64  `json:"memoryLimits"`
	CPURequests    int64  `json:"cpuRequests"`
	MemoryRequests int64  `json:"memoryRequests"`
}

type Pod struct {
	Name        string
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Owners      []Owner           `json:"owner,omitempty"`
	Containers  []Container       `json:"containers,omitempty"`
	Status      string            `json:"status"`
	IP          string            `json:"ip"`
	NodeName    string            `json:"nodeName"`
	Namespace   string            `json:"namespace"`
	UID         string            `json:"uid"`
}

func NewPod(name string) *Pod {
	r := new(Pod)
	r.Name = name
	r.Containers = make([]Container, 0)
	return r
}

type Owner struct {
	ApiVersion         string `json:"apiVersion"`
	BlockOwnerDeletion *bool  `json:"blockOwnerDeletion"`
	Controller         *bool  `json:"controller"`
	Kind               string `json:"kind"`
	Name               string `json:"name"`
}

type Container struct {
	Name      string         `json:"name"`
	Image     string         `json:"image"`
	Resources ResourceValues `json:"resources,omitempty"`
	OwnerUID  string         `json:"ownerUid"`
}
