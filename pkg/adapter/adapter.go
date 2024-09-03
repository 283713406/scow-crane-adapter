package adapter

import (
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	protos "scow-crane-adapter/gen/go"
	"scow-crane-adapter/pkg/services/account"
	"scow-crane-adapter/pkg/services/app"
	"scow-crane-adapter/pkg/services/config"
	"scow-crane-adapter/pkg/services/job"
	"scow-crane-adapter/pkg/services/user"
	"scow-crane-adapter/pkg/services/version"
	"scow-crane-adapter/pkg/utils"
)

var (
	FlagConfigFilePath string
	GConfig            utils.Config
)

func NewAdapterCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "scow-crane-adapter",
		Short: "crane adapter for scow",
		Run: func(cmd *cobra.Command, args []string) {
			Run()
		},
	}

	// Initialize config
	cobra.OnInitialize(func() {
		// Use config file from the flag or search in the default paths
		if FlagConfigFilePath != "" {
			viper.SetConfigFile(FlagConfigFilePath)
		} else {
			viper.AddConfigPath(".")
			viper.AddConfigPath("/etc/scow-crane-adapter/")
			viper.SetConfigType("yaml")
			viper.SetConfigName("config")
		}

		// Read and parse config file
		viper.ReadInConfig()
		// Initialize logger
		utils.InitLogger(utils.ParseLogLevel(viper.GetString("log-level")))
		if err := viper.Unmarshal(&GConfig); err != nil {
			logrus.Fatalf("Error parsing config file: %s", err)
		}

		logrus.Debugf("Using config:\n%+v", GConfig)
	})

	// Specify config file path
	rootCmd.PersistentFlags().StringVarP(&FlagConfigFilePath, "config", "c", "", "Path to configuration file")

	// Other flags
	rootCmd.PersistentFlags().IntP("bind-port", "p", 5000, "Binding address of adapter")
	viper.BindPFlag("bind-addr", rootCmd.PersistentFlags().Lookup("addr"))

	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "Log level")
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))

	return rootCmd
}

func Run() {
	// 初始化CraneCtld客户端及鹤思配置文件
	utils.InitClientAndCraneConfig()

	s := grpc.NewServer(
		grpc.MaxRecvMsgSize(1024*1024*1024), // 最大接受size 1GB
		grpc.MaxSendMsgSize(1024*1024*1024), // 最大发送size 1GB
	) // 创建gRPC服务器

	// 注册服务
	protos.RegisterJobServiceServer(s, &job.ServerJob{})
	protos.RegisterAccountServiceServer(s, &account.ServerAccount{})
	protos.RegisterConfigServiceServer(s, &config.ServerConfig{})
	protos.RegisterUserServiceServer(s, &user.ServerUser{})
	protos.RegisterVersionServiceServer(s, &version.ServerVersion{})
	protos.RegisterAppServiceServer(s, &app.ServerApp{})

	logrus.Infof("gRPC server listening on %d", GConfig.BindPort)
	portString := fmt.Sprintf(":%d", GConfig.BindPort)
	listener, err := net.Listen("tcp", portString)
	if err != nil {
		logrus.Fatalf("failed to listen: %s", err)
		return
	}

	if err := s.Serve(listener); err != nil {
		logrus.Fatalf("gRPC server quitting: %s", err)
	}
}
