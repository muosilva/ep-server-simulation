package client

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func StartClient() {
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	done := make(chan struct{})

	go func() {
		io.Copy(os.Stdout, conn)
		done <- struct{}{}
	}()

	// Leitura de entradas do usuário
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		text := input.Text()
		if text == "\\exit" {
			conn.Close()
			break
		}
		fmt.Fprintln(conn, text)
	}

	<-done
}
