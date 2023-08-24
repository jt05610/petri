package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/codegen"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/prisma"
	"github.com/jt05610/petri/prisma/db"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func getIntFromUser(prompt string) (int, error) {
	return strconv.Atoi(getStringFromUser(prompt))
}

func getStringFromUser(prompt string) string {
	var input string
	fmt.Print(prompt)
	_, err := fmt.Scanln(&input)
	if err != nil {
		return ""
	}
	return strings.Trim(input, " \r\n\t")
}

func getDeviceID(c *prisma.DeviceClient) *device.ListItem {
	if devID != "" {
		return &device.ListItem{ID: devID, Name: devID}
	}
	dd, err := c.List(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Print("Please select a device:")
	for i, d := range dd {
		fmt.Printf("\n%d. %s", i, d.Name)
	}
	i, err := getIntFromUser("\n> ")
	if err != nil {
		panic(err)
	}
	return dd[i]
}

func getInstanceName() string {
	return getStringFromUser("Please enter an instance name: \n> ")
}

func validateLogin(c *db.PrismaClient, email, password string) (string, error) {
	u, err := c.User.FindUnique(db.User.Email.Equals(email)).With(db.User.Password.Fetch()).Exec(context.Background())
	if err != nil {
		panic(err)
	}
	if m, found := u.Password(); !found {
		return "", errors.New("user has no password")
	} else if bcrypt.CompareHashAndPassword([]byte(m.Hash), []byte(password)) != nil {
		return "", errors.New("password does not match")
	}
	return u.ID, nil
}

func saveLogin(id string) {
	fmt.Println("Saving login")
	df, err := os.OpenFile(".env", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := df.Close()
		if err != nil {
			panic(err)
		}
	}()
	_, err = df.WriteString(fmt.Sprintf("\nAUTHOR_ID=%s\n", id))
	if err != nil {
		panic(err)
	}
}

func login(c *db.PrismaClient) context.Context {
	fmt.Print("Enter email: ")
	var email string
	_, err := fmt.Scanf("%s", &email)
	if err != nil {
		panic(err)
	}
	fmt.Print("Enter password: ")
	// hide text for user entry
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}
	id, err := validateLogin(c, email, string(password))
	if err != nil {
		panic(err)
	}
	fmt.Print("\nSave login? (y/n): ")
	var save string
	_, err = fmt.Scanf("%s", &save)
	if err != nil {
		panic(err)
	}
	if save == "y" {
		saveLogin(id)
	}
	return context.WithValue(context.Background(), "authorID", id)
}

func withAuthorID(ctx context.Context, c *db.PrismaClient) context.Context {
	authorID, found := os.LookupEnv("AUTHOR_ID")
	if !found {
		return login(c)
	}
	return context.WithValue(ctx, "authorID", authorID)
}

func getLanguage() string {
	langOpts := []string{"go", "python"}
	if language != "" {
		return language
	}
	fmt.Print("Please select a language:")
	for i, o := range langOpts {
		fmt.Printf("\n%d. %s", i, o)
	}
	i, err := getIntFromUser("\n> ")
	if err != nil {
		panic(err)
	}
	return langOpts[i]
}

var rootCmd = &cobra.Command{
	Use:   "codegen [-d deviceID -l Language]",
	Short: "Generate code for a device",
	Long:  `Generate code for a device. If no device is specified, a list of devices will be shown.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := godotenv.Load()
		if err != nil {
			panic(err)
		}
		c := db.NewClient()
		if err := c.Connect(); err != nil {
			panic(err)
		}
		defer func() {
			_ = c.Disconnect()
		}()

		ctx := withAuthorID(context.Background(), c)
		devClient := &prisma.DeviceClient{PrismaClient: c}
		device := getDeviceID(devClient)
		instanceName := getInstanceName()
		lang := getLanguage()
		fmt.Printf("Generating %s code for device id %s\n", lang, device)
		params := new(codegen.RabbitMQParams)
		params.FromURI(rabbitMQURI, amqpExchange)

		g := codegen.NewGenerator(devClient, &codegen.Params{
			Language:     codegen.Language(lang),
			Port:         port,
			OutDir:       codegen.ToSnakeCaseFromSentence(instanceName),
			InstanceName: instanceName,
			DeviceID:     device.ID,
			RabbitMQ:     params,
		})
		err = g.Generate(ctx)
		if err != nil {
			panic(err)
		}
	},
}

var devID string
var language string
var port int
var rabbitMQURI string
var amqpExchange string

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&devID, "device", "d", "", "device id")
	rootCmd.Flags().StringVarP(&language, "language", "l", "", "language")
	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "port")
	rootCmd.Flags().StringVarP(&rabbitMQURI, "rabbitMQURI", "r", "amqp://guest:guest@localhost:5672/", "rabbitMQ URI")
	rootCmd.Flags().StringVarP(&amqpExchange, "amqpExchange", "e", "topic_devices", "rabbitMQ exchange")
}
