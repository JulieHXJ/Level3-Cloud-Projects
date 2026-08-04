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
