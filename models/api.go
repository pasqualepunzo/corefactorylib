package models

type CallGetResponse struct {
	Kind      string
	BodyJson  map[string]interface{}
	BodyArray []map[string]interface{}
	Errore    int32
	Log       string
}
