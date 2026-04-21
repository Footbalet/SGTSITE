package main

import (
	"fmt"
	"github.com/codecat/go-enet"
	"strconv"
	"strings"
)

func StartServerENET(s *Server) {
	// Инициализация (без возвращаемого значения)

	var idToClient = make(map[enet.Peer]*Client)

	enet.Initialize()
	defer enet.Deinitialize()

	// Создаем хост (сервер) - слушаем порт 7777
	// NewListenAddress(port) создает адрес 0.0.0.0:port
	host, err := enet.NewHost(enet.NewListenAddress(7777), 32, 2, 0, 0)
	if err != nil {
		fmt.Printf("Ошибка создания хоста: %s\n", err.Error())
		return
	}
	defer host.Destroy()

	fmt.Println("ENet сервер запущен на порту 7777")

	// Основной цикл обработки событий
	for {
		// Ожидаем событие (таймаут 1000 мс)
		ev := host.Service(1000)

		// Если событий нет - просто продолжаем
		if ev.GetType() == enet.EventNone {
			continue
		}

		switch ev.GetType() {
		case enet.EventConnect:
			fmt.Printf("[+] Клиент подключен: %s\n", ev.GetPeer().GetAddress())

		case enet.EventDisconnect:
			fmt.Printf("[-] Клиент отключен: %s\n", ev.GetPeer().GetAddress())

		case enet.EventReceive:
			packet := ev.GetPacket()
			// ВАЖНО: уничтожаем пакет после использования
			defer packet.Destroy()

			// Получаем данные как строку
			data := string(packet.GetData())
			fmt.Printf("[RECV] %s: %s\n", ev.GetPeer().GetAddress(), data)
			if strings.HasPrefix(data, "init") {
				playerID := data[len("init"):]
				id, _ := strconv.ParseInt(playerID, 10, 64)
				client := s.clientPool.Get(id)
				idToClient[ev.GetPeer()] = client
				client.enetID = ev.GetPeer()
				fmt.Printf("[++] Клиент инициализирован: %s\n", ev.GetPeer().GetAddress())
				ev.GetPeer().SendString("Hello", 0, enet.PacketFlagReliable)
			} else if strings.HasPrefix(data, "gd") {
				client := idToClient[ev.GetPeer()]
				room := s.rooms[client.currentRoom]
				for _, resident := range room.clients {
					if resident.ready && resident != client {
						resident.enetID.SendString(data, 0, enet.PacketFlagReliable)
					}
				}
			}
		}
	}
}
