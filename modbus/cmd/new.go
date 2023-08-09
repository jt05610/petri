/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/

package cmd

import (
	"context"
	ax "core/axis/v1/axis"
	"github.com/spf13/cobra"
	"log"
	"modbus/v1/axis"
	"modbus/v1/modbus"
	"time"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		regMap := &modbus.RegisterMap{
			Coils: []*modbus.Coil{
				{
					Name:    "Start",
					Address: 0x00,
				},
				{
					Name:    "Stop",
					Address: 0x01,
				},
				{
					Name:    "Home",
					Address: 0x02,
				},
				{
					Name:    "MoveToStall",
					Address: 0x03,
				},
				{
					Name:    "SetZero",
					Address: 0x04,
				},
				{
					Name:    "Enable",
					Address: 0x05,
				},
			},

			DiscreteInputs: []*modbus.DiscreteInput{
				{
					Name:    "IsMoving",
					Address: 0x00,
				},
			},
			HoldingRegisters: []*modbus.HoldingRegister{
				{
					Name:    "TargetPosition",
					Address: 0x00,
				},
				{
					Name:    "TargetVelocity",
					Address: 0x01,
				},
				{
					Name:    "MoveTo",
					Address: 0x02,
				},
				{
					Name:    "Acceleration",
					Address: 0x03,
				},
			},
			InputRegisters: []*modbus.InputRegister{
				{
					Name:    "CurrentPosition",
					Address: 0x00,
				},
				{
					Name:    "CurrentVelocity",
					Address: 0x01,
				},
				{
					Name:    "TStep",
					Address: 0x02,
				},
				{
					Name:    "Force",
					Address: 0x03,
				},
			},
		}
		axs := axis.AddAxisRequest{
			Name:        "Test Axis",
			UnitId:      unitID,
			RegisterMap: regMap,
			Calibration: &ax.Calibration{},
		}

		conn := connect()
		defer func() {
			if err := conn.Close(); err != nil {
				log.Println(err)
			}
		}()
		client := storeClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()

		newAx, err := client.AddAxis(ctx, &axs)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(newAx)
	},
}

func init() {
	axisCmd.AddCommand(newCmd)

}
