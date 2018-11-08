<<<<<<< HEAD
package request



type DsnAddFileReq struct {
	Cid         string  `json:"cid"`
	Redundency  int     `json:"redundency"`
	IsDir       bool    `json:"dir"`
	Chunk       uint64  `json:"chunk"`
	FileSize    uint64  `json:"filesize"`

}

type DsnIpInfoReq struct {

	Iplists     []string  `iplists`

=======
package request



type DsnAddFileReq struct {

	Cid         string  `json:"cid"`
	Redundency  int     `json:"redundency"`
	IsDir       bool    `json:"dir"`
	Chunk       uint64  `json:"chunk"`
	FileSize    uint64  `json:"filesize"`

}

type DsnIpInfoReq struct {

	Iplists     []string  `iplists`

>>>>>>> 5667e7017b9ea155c77e72ae0039c9aadbc55360
}