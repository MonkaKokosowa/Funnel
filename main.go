package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/urfave/cli/v2"
)

type Server struct {
	Name               string "json: name"
	IP                 string "json: ip"
	Port               int    "json: port"
	Username           string "json: username"
	IdentityFilePath   string "json: identityFilePath"
	IdentityPassphrase string "json: identityPassphrase"
	Password           string "json: password"
}

func main() {
	app := &cli.App{
		Name:  "funnel",
		Usage: "Main CLI!",
		Action: func(*cli.Context) error {
			fmt.Println("Hi girl! What we doin today :3")
			fmt.Println("Select option:\n")
			fmt.Println("1) Connect to server")
			fmt.Println("2) Add a new server")
			fmt.Println("3) Remove old server")
			fmt.Print("> ")

			// Capture user input
			scanner := bufio.NewScanner(os.Stdin)
			if !scanner.Scan() {
				log.Fatal("Failed to read input")
			}

			// Parse the user input (convert the input to an integer)
			input := scanner.Text()
			option, err := strconv.Atoi(input)
			if err != nil {
				log.Fatalf("Invalid input: %v\n", err)
			}

			switch option {
			case 1:
				if len(get_servers()) == 0 {
					log.Fatal("No servers to connect to")
				}
				fmt.Println("Which one?\n")
				for index, server := range get_servers() {
					fmt.Printf("%d) %s\n", index, server.Name)
				}
				output := new_scanner().Text()
				option, err := strconv.Atoi(output)
				if err != nil {
					log.Fatalf("Invalid input: %v\n", err)
				}
				if option < 0 || option >= len(get_servers()) {
					log.Fatalf("Out of range: %v\n", len(get_servers()))
				}
				server := get_servers()[option]
				fmt.Println("Connecting to server:", server.Name)
				execute_command(server)
			case 2:
				fmt.Println("Adding a new server")

				fmt.Print("Name:\n>")
				name := new_scanner().Text()

				fmt.Print("IP:\n>")
				ip := new_scanner().Text()

				fmt.Print("Port:\n>")
				port, err := strconv.Atoi(new_scanner().Text())
				if err != nil {
					log.Fatalf("Invalid input: %v\n", err)
				}
				if port <= 0 {
					port = 22
				}

				fmt.Print("Username:\n> ")
				username := new_scanner().Text()

				fmt.Print("Identity file path (absolute, empty for none):\n> ")
				identityFilePath := new_scanner().Text()

				fmt.Print("Identity passphrase (empty for none):\n> ")
				identityPassphrase := new_scanner().Text()

				if identityPassphrase != "" {
					fmt.Print("Identity passphrase again:\n> ")
					if identityPassphrase != new_scanner().Text() {
						log.Fatal("Passphrases do not match")

					}
				}

				fmt.Print("Password (empty for none):\n> ")
				password := new_scanner().Text()

				if password != "" {
					fmt.Print("Password again:\n> ")
					if password != new_scanner().Text() {
						log.Fatal("Passwords do not match")
						return nil
					}
				}
				server := Server{
					Name:               name,
					IP:                 ip,
					Port:               port,
					Username:           username,
					IdentityFilePath:   identityFilePath,
					IdentityPassphrase: identityPassphrase,
					Password:           password,
				}
				add_server(server)

				fmt.Println(get_servers())

			case 3:
				if len(get_servers()) == 0 {
					log.Fatal("No servers to connect to")
				}
				fmt.Println("Which one?\n")
				for index, server := range get_servers() {
					fmt.Printf("%d) %s\n", index, server.Name)
				}
				fmt.Print("> ")

				output := new_scanner().Text()
				option, err := strconv.Atoi(output)

				if err != nil {
					log.Fatalf("Invalid input: %v\n", err)
				}
				if option < 0 || option >= len(get_servers()) {
					log.Fatalf("Out of range: %v\n", len(get_servers()))
				}
				remove_server(option)
				fmt.Println("Removed server")

			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
func execute_command(server Server) {
	command := "ssh"
	var args []string

	if server.IdentityFilePath != "" {
		args = append(args, "-i", server.IdentityFilePath)
	}
	if server.IdentityPassphrase != "" {
		fmt.Println("Identity passphrase: ", server.IdentityPassphrase)
	}
	if server.Password != "" {
		fmt.Println("Password: ", server.Password)
	}
	args = append(args, fmt.Sprintf("%s@%s:%d", server.Username, server.IP, server.Port))
	fmt.Println(command, args)

	cmd := exec.Command(command, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}

	fmt.Println("\nCommand executed successfully.")
}

func new_scanner() *bufio.Scanner {
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		log.Fatal("Failed to read input")
	}
	return scanner
}
func remove_server(index int) {
	// get config dir and read contenst of servers.json

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error getting home directory: ", err)
	}

	dirPath := filepath.Join(homeDir, ".funnel")
	filePath := filepath.Join(dirPath, "server.json")

	// Check if the directory exists, create it if not
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Println("Directory does not exist. Creating:", dirPath)
		err := os.MkdirAll(dirPath, 0755) // create directory with permissions
		if err != nil {
			log.Fatal("Error creating directory:", err)
		}
	}

	// Read the contents of the file
	content, err := os.ReadFile(filePath) // You can also use os.ReadFile() in Go 1.16+
	if err != nil {
		log.Fatal("Error reading file:", err)
	}
	if string(content) == "[]" || string(content) == "" {
		log.Fatal("No servers to remove")
	}

	var servers []Server

	// Unmarshal
	err = json.Unmarshal([]byte(content), &servers)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	servers = append(servers[:index], servers[index+1:]...)

	// Write the updated data back to the file
	updatedContent, err := json.Marshal(servers)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	err = os.WriteFile(filePath, updatedContent, 0644)

	if err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}

}
func add_server(server Server) {
	// get config dir and read contenst of servers.json

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error getting home directory: ", err)
	}

	dirPath := filepath.Join(homeDir, ".funnel")
	filePath := filepath.Join(dirPath, "server.json")

	// Check if the directory exists, create it if not
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Println("Directory does not exist. Creating:", dirPath)
		err := os.MkdirAll(dirPath, 0755) // create directory with permissions
		if err != nil {
			log.Fatal("Error creating directory:", err)
		}
	}

	// Read the contents of the file
	content, err := os.ReadFile(filePath) // You can also use os.ReadFile() in Go 1.16+
	if err != nil {
		log.Fatal("Error reading file:", err)
	}
	if string(content) != "[]" && string(content) != "" {

		var servers []Server
		err := json.Unmarshal([]byte(content), &servers)
		if err != nil {
			log.Fatalf("Error unmarshalling JSON: %v", err)
		}
		servers = append(servers, server)

		// Write the updated data back to the file
		updatedContent, err := json.Marshal(servers)
		if err != nil {
			log.Fatalf("Error marshalling JSON: %v", err)
		}
		fmt.Println("upadtedContent ", updatedContent)
		err = os.WriteFile(filePath, updatedContent, 0644)
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}

	} else {
		servers := []Server{server}
		// Write the updated data back to the file
		updatedContent, err := json.Marshal(servers)
		if err != nil {
			log.Fatalf("Error marshalling JSON: %v", err)
		}
		err = os.WriteFile(filePath, updatedContent, 0644)
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}

	}

	//var servers []Server

	// Unmarshal
}

func get_servers() []Server {

	// get config dir and read contenst of servers.json

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error getting home directory: ", err)
	}

	dirPath := filepath.Join(homeDir, ".funnel")
	filePath := filepath.Join(dirPath, "server.json")

	// Check if the directory exists, create it if not
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Println("Directory does not exist. Creating:", dirPath)
		err := os.MkdirAll(dirPath, 0755) // create directory with permissions
		if err != nil {
			log.Fatal("Error creating directory:", err)
		}
	}

	// Check if the file exists, create it if not
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println("File does not exist. Creating:", filePath)
		// Create the file
		file, err := os.Create(filePath)
		if err != nil {
			log.Fatal("Error creating file:", err)
		}
		defer file.Close()

		// Write some initial data to the file
		initialContent := "[]"
		_, err = file.WriteString(initialContent)
		if err != nil {
			log.Fatal("Error writing to file:", err)
		}
	}

	// Read the contents of the file
	content, err := os.ReadFile(filePath) // You can also use os.ReadFile() in Go 1.16+
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	var servers []Server

	// Unmarshal the JSON into the people slice
	err = json.Unmarshal([]byte(content), &servers)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	return servers
}
