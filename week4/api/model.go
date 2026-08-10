package main

type DBInstance struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Instances int    `json:"instances"`
	Storage   string `json:"storage"`
	CPU       string `json:"cpu"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

// POST request
type CreateInstanceRequest struct {
	Name      string  `json:"name"`
	Instances int     `json:"instances"`
	Storage   *string `json:"storage,omitempty"`
	CPU       *string `json:"cpu,omitempty"`
}

// // PUT
// type UpdateInstanceRequest struct {
// 	Name      string `json:"name"`
// 	Instances int    `json:"instances"`
// }

//PATCH
type PatchInstanceRequest struct {
	Name      *string `json:"name,omitempty"`
	Instances *int    `json:"instances,omitempty"`
	Storage   *string `json:"storage,omitempty"`
	CPU       *string `json:"cpu,omitempty"`
}

type ConnectionInfo struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	URI      string `json:"uri,omitempty"`

}