package model

type Config struct {
	JwtSecret                     string `json:"jwt_secret"`
	ServerAddress                 string `json:"server_address"`
	ServerPort                    string `json:"server_port"`
	Database                      string `json:"database"`
	DatabaseUser                  string `json:"database_user"`
	DatabasePassword              string `json:"database_password"`
	DatabaseAddress               string `json:"database_address"`
	DatabasePort                  string `json:"database_port"`
	AccessControlAllowOrigin      string `json:"access_control_allow_origin"`
	AccessControlAllowCredentials string `json:"access_control_allow_credentials"`
	AccessControlAllowMethods     string `json:"access_control_allow_methods"`
	AccessControlAllowHeaders     string `json:"access_control_allow_headers"`
	Swaggerfile                   string `json:"swagger_file"`
	ApiKey                        string `json:"api_key"`
	MaxImageSize                  int64  `json:"max_image_size"`
	MaxVideoSize                  int64  `json:"max_video_size"`
}
