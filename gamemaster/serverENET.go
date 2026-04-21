package main

import (
	"fmt"
	"github.com/codecat/go-enet"
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
			data := packet.GetData()
			if len(data) < 2 {
				fmt.Printf("Некорректный пакет: слишком мало данных (%d байт)\n", len(data))
				continue
			}
			commandID := string(data[0:2])
			payload := data[2:]
			switch commandID {
			case "in":
				if len(payload) < 4 {
					fmt.Printf("Некорректный init пакет: ожидается 4 байт, получено %d\n", len(payload))
					continue
				}
				playerID := int32(payload[0]) |
					int32(payload[1])<<8 |
					int32(payload[2])<<16 |
					int32(payload[3])<<24
				client := s.clientPool.Get(int64(playerID))
				idToClient[ev.GetPeer()] = client
				client.enetID = ev.GetPeer()
				fmt.Printf("[++] Клиент инициализирован: %s (ID: %d)\n", ev.GetPeer().GetAddress(), playerID)
				helloData := []byte("Hello")
				ev.GetPeer().SendBytes(helloData, 0, enet.PacketFlagReliable)
			case "gr":
				client := idToClient[ev.GetPeer()]
				room := s.rooms[client.currentRoom]
				for _, resident := range room.clients {
					if resident.ready && resident != client {
						resident.enetID.SendBytes(data, 0, enet.PacketFlagReliable)
					}
				}
			case "gu":
				client := idToClient[ev.GetPeer()]
				room := s.rooms[client.currentRoom]
				for _, resident := range room.clients {
					if resident.ready && resident != client {
						resident.enetID.SendBytes(data, 0, enet.PacketFlagUnsequenced)
					}
				}
			}
		}
	}
}
