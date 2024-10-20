package bot

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func StartBot() {
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	done := make(chan struct{})

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			msg := scanner.Text()
			fmt.Println("Mensagem de:", msg)

			// Resposta invertida
			reversed := reverseString(msg)
			fmt.Fprintln(conn, reversed)
		}
		done <- struct{}{}
	}()

	// Bot faz login
	fmt.Fprintln(conn, "\\changenick bot-inverter")
	<-done
}

func reverseString(input string) string {
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
