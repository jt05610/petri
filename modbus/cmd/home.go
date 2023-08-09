/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"core/axis/v1/axis"
	"log"
	"time"

	"github.com/spf13/cobra"
)

// homeCmd represents the home command
var homeCmd = &cobra.Command{
	Use:   "home",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		conn := connect()
		defer func() {
			if err := conn.Close(); err != nil {
				log.Println(err)
			}
		}()
		sc := storeClient(conn)
		dev := deviceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Minute)
		defer cancel()
		all, err := sc.ListAxis(ctx, &axis.Empty{})
		if err != nil {
			log.Fatalf("could not list axis: %v", err)
		}
		for _, a := range all.Devices {
			if *a.Axis.UnitId == unitID {
				_, err = dev.Home(ctx, &axis.HomeRequest{Id: a.Axis.Id})
				if err != nil {
					log.Fatalf("could not home axis: %v", err)
				}
			}
		}
	},
}

func init() {
	axisCmd.AddCommand(homeCmd)

}
