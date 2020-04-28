package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var (
	loggerOutput  = os.Stderr
	defaultOutput = os.Stdout
)

// Execute executes the root command.
func Execute() error {
	rootCmd := newCommandRoot()
	return rootCmd.Execute()
}

func writeOut(line string) {
	fmt.Fprintln(defaultOutput, line)
}

func newCommandRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ws",
		Short:   "websocket cli toolkit.",
		Version: "0.1.0",
	}
	cmd.AddCommand(
		newCommandRaw(),
		newCommandServe(),
	)
	return cmd
}

func newCommandRaw() *cobra.Command {
	var uri string
	cmd := &cobra.Command{
		Use:   "text",
		Short: "Connect stdin/stdout to wsserver.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			u, err := url.Parse(uri)
			if err != nil {
				return err
			}

			log.Printf("connecting to %q", u.String())
			conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), http.Header{})
			if err != nil {
				return err
			}
			defer conn.Close()

			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)

			stdinChan := read(os.Stdin)
			done := make(chan struct{})

			// Read from websocket connection and write to stdout
			go func() {
				defer close(done)

				conn.SetReadLimit(maxMessageSize)
				conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
				for {
					_, message, err := conn.ReadMessage()
					if err != nil {
						if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
							log.Printf("read error: %v", err)
						}
						log.Printf("clean error: %v", err)
						break
					}
					message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
					log.Printf("read: %s", message)
				}
			}()

			// Read from stdin and write to websocket connection
		Loop:
			for {
				select {
				case <-done:
					break Loop
				case message := <-stdinChan:
					conn.SetWriteDeadline(time.Now().Add(writeWait))
					err := conn.WriteMessage(websocket.TextMessage, []byte(message))
					if err != nil {
						return fmt.Errorf("write error: %v", err)
					}
				case <-interrupt:
					log.Println("interrupt")
					conn.SetWriteDeadline(time.Now().Add(writeWait))
					err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					if err != nil {
						return fmt.Errorf("write close error: %v", err)
					}
					break Loop
				}
			}
			return nil
		},
	}
	cmd.Flags().SortFlags = false
	flag.StringVar(&uri, "uri", "ws://localhost:8080/ws", "uri of wsserver in the form of ws://host:port/path/to/wsserver")
	return cmd
}

func newCommandServe() *cobra.Command {
	var addr string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve wsserver.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			hub := NewHub()
			go hub.Run()

			log.Printf("HTTP Server listening on %q", addr)
			return http.ListenAndServe(addr, hub.Handler())
		},
	}
	cmd.Flags().SortFlags = false
	flag.StringVar(&addr, "addr", ":8080", "listen address")
	return cmd
}

func main() {
	if err := Execute(); err != nil {
		writeOut(err.Error())
		os.Exit(1)
	}
}

func read(r io.Reader) <-chan string {
	lines := make(chan string)
	go func() {
		defer close(lines)
		scan := bufio.NewScanner(r)
		for scan.Scan() {
			lines <- scan.Text()
		}
	}()
	return lines
}
