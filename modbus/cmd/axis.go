/*
Copyright © 2023 Jonathan Taylor <jonrtaylor12@gmail.com>
*/

package cmd

import (
	ax "core/axis/v1/axis"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"modbus/v1/axis"
)

//go:embed secrets/cert.pem
var certFs embed.FS

var serverAddr string
var unitID int32
var timeout int32

// axisCmd represents the axis command
var axisCmd = &cobra.Command{
	Use:   "axis",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func init() {
	rootCmd.AddCommand(axisCmd)

	axisCmd.Flags().StringVarP(&serverAddr, "server", "s", "localhost:55056", "server address")
	axisCmd.Flags().Int32VarP(&unitID, "unit-id", "u", 2, "modbus unit id")
	axisCmd.Flags().Int32VarP(&timeout, "timeout", "t", 5, "timeout in seconds")
}

func connect() *grpc.ClientConn {
	certPool := x509.NewCertPool()
	caCert, err := certFs.ReadFile("secrets/cert.pem")
	if err != nil {
		panic(err)
	}
	certPool.AppendCertsFromPEM(caCert)
	conn, err := grpc.Dial(
		serverAddr,
		grpc.WithTransportCredentials(
			credentials.NewTLS(
				&tls.Config{
					CurvePreferences: []tls.CurveID{tls.CurveP256},
					MinVersion:       tls.VersionTLS12,
					RootCAs:          certPool,
				},
			),
		),
	)
	if err != nil {
		panic(err)
	}
	return conn
}

func storeClient(conn *grpc.ClientConn) axis.StoreClient {

	return axis.NewStoreClient(conn)
}

func deviceClient(conn *grpc.ClientConn) ax.DeviceClient {

	return ax.NewDeviceClient(conn)
}
