package _type

type HealthResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Code    int    `json:"code"`
}
