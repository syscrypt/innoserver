package model

type Config struct {
	JwtSecret        string `json:"jwt_secret"`
	ServerAddress    string `json:"server_address"`
	ServerPort       string `json:"server_port"`
	Database         string `json:"database"`
	DatabaseUser     string `json:"database_user"`
	DatabasePassword string `json:"database_password"`
	DatabaseAddress  string `json:"database_address"`
	DatabasePort     string `json:"database_port"`
	MaxImageSize     int64  `json:"max_image_size"`
	MaxVideoSize     int64  `json:"max_video_size"`
}
