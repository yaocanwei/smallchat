# Chat Server in Go

This repository contains a simple chat server written in Go. The server allows multiple clients to connect and communicate with each other using TCP connections. Clients can set their nicknames and send messages to the entire chat room.

## Features

- Multi-client support using TCP connections.
- Nickname setting for clients with the `/nick` command.
- Broadcast messages to all connected clients.
- Max client limitation to prevent overloading the server.
- Graceful handling of client disconnection.

## Getting Started

### Prerequisites

Before you can run this server, you need to have Go installed on your system. You can download and install Go from [the official site](https://golang.org/dl/).

### Installing

To start using this chat server, you need to clone this repository to your local machine. You can do this by running the following command:

```sh
git clone https://github.com/yaocanwei/smallchat.git
```

After cloning the repository, navigate to the directory where the repository is located and build the server using the Go build command:

```sh
go build
```

### Running the Server

To run the server, simply execute the built binary:

```sh
./main
```

The server will start and listen for incoming TCP connections on port 7712.

### Connecting Clients

To connect a client to the server, use any TCP client, such as `netcat` or `telnet`:

```sh
telnet localhost 7712
```

Once connected, you can set your nickname using the `/nick` command followed by your desired nickname:

```
/nick YourNickname
```

After setting a nickname, any text you type will be sent to all connected clients.

## Architecture

The server uses a concurrent design where each client connection is handled in a separate goroutine. This allows for scalability and responsive client handling.

- `ChatSystem` struct - Holds the core chat logic, including observer management and message broadcasting.
- `Client` struct - Represents a connected client and is responsible for sending and receiving messages.

## Code Structure

- `main.go` - The entry point of the server that initializes the chat system and listens for client connections.

## Development

Feel free to fork and extend the functionality of this chat server. You can add new features like private messaging, chat rooms, and user authentication as needed.

## Contributions

Contributions are welcome! For major changes, please open an issue first to discuss what you would like to change. Please make sure to update tests as appropriate.

## License

This project is open-sourced under the MIT License. See the [LICENSE](LICENSE) file for details.

---

The README template provided above should be added to your repository with modifications reflecting your specific GitHub username and repository information where applicable. Make sure to add a LICENSE file containing the MIT License text to your repository as well.