package v1

import "k8s.io/apimachinery/pkg/apis/meta/v1"


// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Network struct {
	v1.TypeMeta `json:",inline"`
	v1.ObjectMeta `json:"metadata,omitempty"`
	Spec networkspec `json:"spec"`
}

type networkspec struct {
	Cidr string `json:"cidr"`
	Gateway string `json:"gateway"`
}


// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NetworkList struct {
	v1.TypeMeta `json:",inline"`
	v1.ListMeta `json:"metadata"`
	Items []Network `json:"items"`
}