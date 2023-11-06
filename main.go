/* chatserver.go -- A simple chat server in Go.
 *
 * This program sets up a TCP server that listens for incoming connections
 * and allows clients to send messages to each other in a chat-room style interface.
 * Clients can set their nicknames with the /nick command and send messages to all
 * other connected clients.
 *
 * Features:
 * - Concurrent handling of multiple chat clients.
 * - Nickname assignment for clients.
 * - Broadcasting messages to all clients.
 * - Go-routine for each client handling.
 * - Graceful shutdown on receiving interrupt or terminate signals.
 *
 * Copyright (c) 2023, cheney
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 *   * Redistributions of source code must retain the above copyright notice,
 *     this list of conditions and the following disclaimer.
 *   * Redistributions in binary form must reproduce the above copyright
 *     notice, this list of conditions and the following disclaimer in the
 *     documentation and/or other materials provided with the distribution.
 *   * Neither the name of the project nor the names of its contributors may be used
 *     to endorse or promote products derived from this software without
 *     specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

// Constants
const (
	ServerPort     = "7712"                                                                  // Port on which the chat server listens
	MaxClients     = 1000                                                                    // Maximum number of allowed clients
	welcomeMessage = "Welcome to the chat server! Type '/nick NAME' to set your nickname.\n" // Welcome message for clients
	unknownCmdMsg  = "Unsupported command\n"                                                 // Message for unsupported commands
)

// ChatObserver interface defines methods that chat clients should implement.
type ChatObserver interface {
	Notify(message string, senderID int)
}

// ChatSystem represents the chat server.
type ChatSystem struct {
	observers  []ChatObserver // List of chat observers (clients)
	mu         sync.Mutex     // Mutex to protect concurrent access to the observers list
	serversock net.Listener   // Listener for incoming client connections
}

// addObserver adds a chat observer (client) to the list.
func (chat *ChatSystem) addObserver(observer ChatObserver) {
	chat.mu.Lock()
	defer chat.mu.Unlock()
	chat.observers = append(chat.observers, observer)
}

// removeObserver removes a chat observer (client) from the list.
func (chat *ChatSystem) removeObserver(observer ChatObserver) {
	chat.mu.Lock()
	defer chat.mu.Unlock()
	for i, obs := range chat.observers {
		if obs == observer {
			chat.observers = append(chat.observers[:i], chat.observers[i+1:]...)
			break
		}
	}
}

// broadcast sends a message to all connected chat clients.
func (chat *ChatSystem) broadcast(message string, senderID int) {
	chat.mu.Lock()
	defer chat.mu.Unlock()
	for _, observer := range chat.observers {
		observer.Notify(message, senderID)
	}
}

// Client represents a connected chat client.
type Client struct {
	id     int           // Unique client ID
	nick   string        // Nickname of the client
	conn   net.Conn      // Network connection
	chat   *ChatSystem   // Reference to the chat system
	reader *bufio.Reader // Buffered reader for reading client input
}

// Notify sends a message to the client.
func (client *Client) Notify(message string, senderID int) {
	// Send a message to the client
	_, err := client.conn.Write([]byte(message))
	if err != nil {
		log.Printf("Error sending message to client %d: %v", client.id, err)
	}
}

// listen listens for messages from the client and handles them.
func (client *Client) listen() {
	// Send the welcome message to the client
	_, err := client.conn.Write([]byte(welcomeMessage))
	if err != nil {
		log.Printf("Error sending message to client %d: %v", client.id, err)
	}

	for {
		// Read a message from the client
		msg, err := client.reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from client %d: %v", client.id, err)
			}
			break
		}

		// Remove any potential carriage return characters
		msg = strings.ReplaceAll(msg, "\r", "")

		// Handle commands
		client.handleCommand(msg)
	}

	// Remove the client from the chat
	client.chat.removeObserver(client)
	client.conn.Close()
	fmt.Printf("Disconnected client clientID=%d\n", client.id)
}

// handleCommand handles commands sent by the client.
func (client *Client) handleCommand(msg string) {
	// Trim leading and trailing whitespace
	msg = strings.TrimSpace(msg)

	// Check if the message is empty
	if msg == "" {
		// Ignore empty messages (e.g., when clients just hit enter)
		return
	}

	// Check if the message is a command
	if strings.HasPrefix(msg, "/") {
		parts := strings.SplitN(msg, " ", 2)
		command := strings.ToLower(parts[0])

		switch command {
		case "/nick":
			client.handleNickCommand(parts)
		default:
			// Handle unknown commands
			client.Notify(unknownCmdMsg, client.id)
		}
	} else {
		// Regular message broadcasting
		displayMsg := fmt.Sprintf("%s> %s\n", client.nick, msg)
		if client.nick == "" {
			displayMsg = fmt.Sprintf("user:%d> %s\n", client.id, msg)
		}
		client.chat.broadcast(displayMsg, client.id)
	}
}

// handleNickCommand handles the /nick command to set a client's nickname.
func (client *Client) handleNickCommand(parts []string) {
	if len(parts) != 2 {
		client.Notify("Usage: /nick <nickname>\n", client.id)
		return
	}

	newNick := strings.TrimSpace(parts[1])
	if newNick == "" {
		client.Notify("Nickname cannot be empty\n", client.id)
		return
	}

	client.nick = newNick
	notifyMsg := fmt.Sprintf("User %d is now known as %s\n", client.id, client.nick)
	log.Print(notifyMsg)
	client.chat.broadcast(notifyMsg, client.id)
}

// main function
func main() {
	chat := &ChatSystem{}

	err := chat.initChat(ServerPort)
	if err != nil {
		log.Fatalf("Error initializing chat: %v", err)
	}
	defer chat.serversock.Close()

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			conn, err := chat.serversock.Accept()
			if err != nil {
				select {
				case <-exitSignal:
					return
				default:
					log.Printf("Error accepting connection: %v", err)
					continue
				}
			}

			clientID := chat.generateClientID()
			client := &Client{
				id:     clientID,
				conn:   conn,
				chat:   chat,
				reader: bufio.NewReader(conn),
			}

			chat.addObserver(client)
			if len(chat.observers) > MaxClients {
				conn.Close() // Close the new connection if max clients exceeded
				continue
			}

			fmt.Printf("Connected client clientid=%d\n", clientID)
			go client.listen()
		}
	}()

	<-exitSignal
	fmt.Println("Server shutting down...")
}

// initChat initializes the chat server and listens on the specified port.
func (chat *ChatSystem) initChat(port string) error {
	var err error
	chat.serversock, err = net.Listen("tcp", ":"+port)
	return err
}

// generateClientID generates a unique client ID for a new client.
func (chat *ChatSystem) generateClientID() int {
	chat.mu.Lock()
	defer chat.mu.Unlock()
	return len(chat.observers) + 1
}
