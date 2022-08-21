package main

import (
	"fmt"
	"ghostviewer/client"
	"ghostviewer/server"
	"ghostviewer/ui"
	"net"
	"os"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Println(os.Args)
		fmt.Fprintf(os.Stderr, "Usage: %s <client/server> <ip> <port> <http/tcp/udp>\n", os.Args[0])
		os.Exit(1)
	}

	instance := os.Args[1]
	addr := net.ParseIP(os.Args[2])
	port, err := strconv.Atoi(os.Args[3])
	commtype := os.Args[4]

	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid port value %s", os.Args[3])
		os.Exit(1)
	}

	if addr == nil {
		fmt.Fprintf(os.Stderr, "Invalid IP address\n")
		os.Exit(1)
	} else if port < 1 || port > 65535 {
		fmt.Fprintf(os.Stderr, "Invalid port\n")
		os.Exit(1)
	}

	ln, err := net.Listen("tcp", ":"+os.Args[3])

	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't listen on port %d: %s\n", port, err)
		os.Exit(1)
	}

	_ = ln.Close()

	if commtype != "https" && commtype != "tcp" && commtype != "udp" {
		fmt.Fprintf(os.Stderr, "Invalid commtype %s, choose from https, tcp or udp\n", commtype)
		os.Exit(1)
	}

	if instance == "server" {
		var ghostserver server.GServer
		var ghostrenderer *ui.GRenderer

		if commtype == "tcp" {
			ghostserver = &server.TCPGServer{addr.String(), port, nil}
		} else if commtype == "https" {
			ghostserver = &server.HTTPSGServer{addr.String(), port, false}
		}

		ghostrenderer = ui.NewGRenderer()
		ghostserver.Listen()
		go server.ServerViewer(ghostserver, ghostrenderer)
		for !ghostserver.IsConnected() {
		}
		go func() {
			for {
				if !ghostserver.IsConnected() {
					fmt.Println("Client disconnected")
					os.Exit(1)
				}
			}
		}()

		ebiten.SetWindowTitle("Ghostviewer")
		ebiten.SetWindowSize(1280, 720)
		ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
		if err := ebiten.RunGame(ghostrenderer); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to run GUI: %s\n", err)
			os.Exit(1)
		}
	} else if instance == "client" {
		var ghostclient client.GClient
		if commtype == "tcp" {
			ghostclient = &client.TCPGClient{addr.String(), port, nil}
		} else if commtype == "https" {
			ghostclient = &client.HTTPSGClient{addr.String(), port, nil}
		}

		err := ghostclient.Connect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Connection error: %s", err)
			os.Exit(1)
		}

		fmt.Println("Connect success")
		client.ClientCommunicate(ghostclient)
	} else {
		fmt.Fprintf(os.Stderr, "Invalid instance value - use server or client\n")
		os.Exit(1)
	}
}
