package main

type DBInstance struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Instances int    `json:"instances"` // dababase node count
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

// POST request
type CreateInstanceRequest struct {
	Name      string `json:"name"`
	Instances int    `json:"instances"`
}

// PUT
type UpdateInstanceRequest struct {
	Name      string `json:"name"`
	Instances int    `json:"instances"`
}

//PATCH
type PatchInstanceRequest struct {
	Name      *string `json:"name,omitempty"`
	Instances *int    `json:"instances,omitempty"`
}

type ConnectionInfo struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	URI      string `json:"uri,omitempty"`

}