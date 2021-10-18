package model

type User struct {
	LastName  string   `json:"lastname" bson:"lastname"`
	FirstName string   `json:"firstname" bson:"firstname"`
	Email     string   `json:"email" bson:"email"`
	Password  string   `json:"-" bson:"password"`
	Roles     []string `json:"roles"`
}

type Gate struct {
	Name     string   `json:"name" bson:"name"`
	Location string   `json:"location" bson:"location"`
	Machine  []string `json:"machine" bson:"machine"`
	Cameras  []Camera
}

type Camera struct {
	Name  string `json:"name" bson:"name"`
	Codec string `json:"codec" bson:"codec"`
	Rtsp  string `json:"rtsp" bson:"rtsp"`
}

type OpcUAServer struct {
	Name     string   `json:"name" bson:"name"`
	Endpoint string   `json:"endpoint" bson:"endpoint"`
	Policy   string   `json:"policy" bson:"policy"`
	Mode     string   `json:"mode" bson:"mode"`
	Cert     string   `json:"cert" bson:"cert"`
	Key      string   `json:"key" bson:"key"`
	NodeID   []string `json:"nodeid" bson:"nodeid"`
}
