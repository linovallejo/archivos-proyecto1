package Global

type UserInfo struct {
	Id     string
	Status bool
}

type Sesion struct {
	Id_user        int
	Id_grp         int
	System_type    int
	User_name      string
	Path           string
	PartitionId    string
	PartitionStart int32
	Fit            [1]byte
}

var Usuario UserInfo
var SesionActual Sesion
