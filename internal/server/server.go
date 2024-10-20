package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type client struct {
	nickname string
	msgChan  chan string
	conn     net.Conn
}

var (
	entering   = make(chan client)
	leaving    = make(chan client)
	messages   = make(chan string)
	privateMsg = make(chan string)
	clients    = make(map[string]client) // nickname como chave para os clientes
	mu         sync.Mutex
)

func broadcaster() {
	for {
		select {
		case msg := <-messages:
			// Enviar mensagens públicas para todos os usuários
			for _, cli := range clients {
				cli.msgChan <- msg
			}

		case cli := <-entering:
			mu.Lock()
			clients[cli.nickname] = cli
			mu.Unlock()

		case cli := <-leaving:
			mu.Lock()
			delete(clients, cli.nickname)
			close(cli.msgChan)
			mu.Unlock()
		}
	}
}

func handleConn(conn net.Conn) {
	// Inicializar cliente
	ch := make(chan string)
	nickname := setNickname(conn)
	cli := client{nickname: nickname, msgChan: ch, conn: conn}

	// Informar entrada do usuário
	messages <- fmt.Sprintf("Usuário @%s acabou de entrar", nickname)
	entering <- cli

	// Gorrotina para enviar mensagens ao cliente
	go clientWriter(conn, ch)

	// Processar entradas
	input := bufio.NewScanner(conn)
	for input.Scan() {
		handleInput(input.Text(), cli)
	}

	// Desconectar usuário
	leaving <- cli
	messages <- fmt.Sprintf("Usuário @%s saiu", cli.nickname)
	conn.Close()
}

func handleInput(input string, cli client) {
	fmt.Printf("Entrada recebida: %s\n", input) // Adicione esta linha para ver a entrada
	switch {
	case strings.HasPrefix(input, "\\changenick "):
		handleNicknameChange(input, cli)

	case strings.HasPrefix(input, "\\msg "):
		handleMessageCommand(input, cli)

	case strings.HasPrefix(input, "\\exit"):
		cli.conn.Close()

	default:
		messages <- fmt.Sprintf("@%s disse: %s", cli.nickname, input)
	}
}

func handleNicknameChange(input string, cli client) {
	newNickname := strings.TrimPrefix(input, "\\changenick ")
	if newNickname == cli.nickname {
		cli.msgChan <- "Você já está usando esse apelido."
		return
	}
	messages <- fmt.Sprintf("Usuário @%s agora é @%s", cli.nickname, newNickname)
	mu.Lock()
	delete(clients, cli.nickname)
	cli.nickname = newNickname
	clients[newNickname] = cli
	mu.Unlock()
}

func handleMessageCommand(input string, cli client) {
	parts := strings.SplitN(input, " ", 3)
	if len(parts) < 3 {
		cli.msgChan <- "Comando inválido. Use: \\msg @nick mensagem"
		return
	}

	recipient := strings.TrimPrefix(parts[1], "@")
	message := parts[2]

	mu.Lock()
	if recipientClient, ok := clients[recipient]; ok {
		recipientClient.msgChan <- fmt.Sprintf("@%s disse em privado: %s", cli.nickname, message)
		cli.msgChan <- fmt.Sprintf("Mensagem privada enviada para @%s", recipient)
	} else {
		cli.msgChan <- "Usuário não encontrado."
	}
	mu.Unlock()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		if msg != "" {
			fmt.Fprintln(conn, msg)
		}
	}
}

func setNickname(conn net.Conn) string {
	conn.Write([]byte("Escolha um apelido: "))
	input := bufio.NewScanner(conn)
	input.Scan()
	return input.Text()
}

func StartServer() {
	listener, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	go broadcaster()

	fmt.Println("Servidor de chat iniciado...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}
