package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	port_eney        = 12345
	maxConn          = 32
	need_version     = "1.0"
	port             = "8085"
	readBufferSize   = 4096
	writeBufferSize  = 4096
	handshakeTimeout = 10 * time.Second
)

// Константы для оптимизации строк
const (
	salt1         = "HUD_career_time"
	salt2         = "fiqu90ru90[frjw903]"
	msgLowVersion = "low_version"
	msgConnected  = "connected"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
}

// ClientPool для управления соединениями
type ClientPool struct {
	clients map[*Client]bool
	mu      sync.RWMutex
}

func (cp *ClientPool) Add(client *Client) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.clients[client] = true
}

func (cp *ClientPool) Remove(client *Client) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	delete(cp.clients, client)
}

func (cp *ClientPool) Exists(client *Client) bool {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	_, exists := cp.clients[client]
	return exists
}

func (cp *ClientPool) Size() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	length := len(cp.clients)
	return length
}

type Client struct {
	conn        *websocket.Conn
	mu          sync.Mutex
	name        string
	appearances string
	weaponHand  string
	voice       string
	isFemale    string
	currentRoom int64
	ready       bool
	team        string
	id          int64
	xorKey      string
	privateKey  *rsa.PrivateKey
	PublicKey   *rsa.PublicKey
	server      *Server
	lastActive  time.Time
}

type Room struct {
	owner      *Client
	clients    []*Client
	excluded   map[*Client]bool
	mapID      string
	mapName    string
	gameMode   string
	rules      string
	order      string
	orderIndex int
	maxPlayers int
	started    bool
	loaded     bool
	password   string
	remoteID   int64
	mu         sync.RWMutex // Для безопасного доступа к комнате
}

type Server struct {
	rooms            map[int64]*Room
	clientPool       *ClientPool
	nextClientID     int64
	nextRoomID       int64
	broadcast        chan string
	mu               sync.RWMutex
	prefixToCommand  map[string]CommandHandler
	shutdown         chan struct{}
	playersDayPeak   int
	playersTotalPeak int
}

type CommandHandler func(*Client, string)

// Пул байтовых буферов для уменьшения аллокаций
var byteBufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 512)
	},
}

func NewServer() *Server {
	s := &Server{
		rooms:            make(map[int64]*Room),
		clientPool:       &ClientPool{clients: make(map[*Client]bool)},
		nextClientID:     1,
		nextRoomID:       1,
		broadcast:        make(chan string, 100), // Буферизированный канал
		shutdown:         make(chan struct{}),
		playersDayPeak:   0,
		playersTotalPeak: 0,
	}
	s.loadStats()
	s.prefixToCommand = map[string]CommandHandler{
		"create_room":              s.createRoom,
		"show_rooms":               s.showRoom,
		"leave_room":               s.leaveRoom,
		"enter_room":               s.enterRoom,
		"switch_ready":             s.switchReady,
		"current_room":             s.currentRoom,
		"start_game":               s.startGame,
		"join_the_game":            s.joinGame,
		"say_room":                 s.sayRoom,
		"say_personal":             s.sayPersonal,
		"say_team":                 s.sayTeam,
		"exclude":                  s.excludePlayer,
		"command":                  s.sendCommand,
		"send_player_data":         s.sendData,
		"gd":                       s.sendData,
		"set_room_loaded":          s.setRoomLoaded,
		"change_rules_room":        s.changeRules,
		"end_game_force":           s.endGame,
		"end_game":                 s.endGame,
		"get_game_data":            s.gameDataRequest,
		"send_game_data":           s.gameDataSend,
		"switch_team":              s.switchTeam,
		"search_rooms":             s.searchRooms,
		"send_time":                s.sendTime,
		"send_switch_team_in_game": s.sendSwitchTeamInGame,
		"leave_game":               s.leaveGame,
		"returned_to_room":         s.returnToRoom,
		"restart_the_game":         s.restartGame,
	}

	return s
}

type ServerStats struct {
	PlayersTotalPeak int       `json:"players_total_peak"`
	LastDate         time.Time `json:"last_date"`
	LastDayPeak      int       `json:"last_day_peak"`
}

func (s *Server) saveStats() {
	stats := ServerStats{
		PlayersTotalPeak: s.playersTotalPeak,
		LastDate:         time.Now(),
		LastDayPeak:      s.playersDayPeak,
	}

	data, _ := json.Marshal(stats)
	if err := os.WriteFile("server_stats.json", data, 0644); err != nil {
		log.Printf("Ошибка при сохранении статистики: %v", err)
	}
}

func (s *Server) loadStats() {
	data, err := os.ReadFile("server_stats.json")
	if err != nil {
		return
	}

	var stats ServerStats
	if err := json.Unmarshal(data, &stats); err != nil {
		log.Printf("Ошибка при размаршалинге статистики: %v", err)
		return
	}

	s.mu.Lock()
	s.playersTotalPeak = stats.PlayersTotalPeak
	// Если сегодня новый день, начинаем с текущего количества
	if stats.LastDate.Day() != time.Now().Day() {
		s.playersDayPeak = s.clientPool.Size()
	} else {
		s.playersDayPeak = stats.LastDayPeak
	}
	s.mu.Unlock()
}

func (s *Server) resetDailyPeak() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Вычисляем время до следующего полуночи для первого срабатывания
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	duration := nextMidnight.Sub(now)

	// Ждем до полуночи
	time.Sleep(duration)

	for {
		select {
		case <-s.shutdown:
			s.saveStats()
			return
		case <-ticker.C:
			// Логируем пик за прошедший день
			log.Printf("Дневной пик подключений за %s: %d",
				time.Now().Add(-24*time.Hour).Format("2006-01-02"),
				s.playersDayPeak)

			// Сбрасываем счетчик
			s.mu.Lock()
			s.playersDayPeak = s.clientPool.Size() // Устанавливаем текущее количество как начальное значение
			s.saveStats()
			s.mu.Unlock()

		}
	}
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем таймаут для handshake
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка при подключении: %v", err)
		return
	}
	tcpConn := conn.NetConn().(*net.TCPConn)
	if err := tcpConn.SetNoDelay(true); err != nil { // Отключаем Nagle
		log.Printf("Ошибка SetNoDelay: %v", err)
	}
	if err := tcpConn.SetWriteBuffer(256 * 1024); err != nil { // Увеличиваем буфер
		log.Printf("Ошибка SetWriteBuffer: %v", err)
	}

	// Устанавливаем таймауты на соединение
	if err := conn.SetReadDeadline(time.Now().Add(handshakeTimeout)); err != nil {
		conn.Close()
		return
	}

	client := s.openConnection(conn)
	if client == nil {
		conn.Close()
		return
	}

	// Сбрасываем таймаут после успешного подключения
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		s.closeConnection(conn, client)
		return
	}

	s.handleMessages(client)
}

func (s *Server) openConnection(conn *websocket.Conn) *Client {
	// Используем пул буферов
	buf := byteBufferPool.Get().([]byte)
	defer byteBufferPool.Put(buf[:0])

	_, clientData, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Ошибка при чтении данных клиента: %v", err)
		return nil
	}

	client := s.initializeClient(conn, clientData)
	if client == nil {
		return nil
	}

	return client
}

func (s *Server) initializeClient(conn *websocket.Conn, clientData []byte) *Client {
	parts := strings.Split(string(clientData), ";")

	// Проверяем минимальное количество частей
	if len(parts) < 2 || parts[0] != need_version {
		log.Printf("Неверная версия клиента или недостаточно данных")
		s.sendLowVersionResponse(conn)
		return nil
	}

	if len(parts) != 8 {
		log.Printf("Неверный формат данных клиента: ожидалось 8 полей, получено %d", len(parts))
		s.sendLowVersionResponse(conn)
		return nil
	}

	// Проверка времени и хеша
	if !s.validateClientData(parts, clientData) {
		s.sendLowVersionResponse(conn)
		return nil
	}

	// Генерация ключей
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Printf("Не удалось создать пару ключей: %v", err)
		s.sendLowVersionResponse(conn)
		return nil
	}

	// Создание клиента
	client := &Client{
		conn:        conn,
		name:        parts[1],
		appearances: parts[2],
		weaponHand:  parts[3],
		voice:       parts[4],
		isFemale:    parts[5],
		privateKey:  privateKey,
		PublicKey:   &privateKey.PublicKey,
		server:      s,
		lastActive:  time.Now(),
	}

	// Регистрация клиента
	s.mu.Lock()
	client.id = s.nextClientID
	s.nextClientID++
	s.mu.Unlock()

	s.clientPool.Add(client)

	client.xorKey = fmt.Sprintf("%d90u9qufre902]%s", client.id, parts[1])

	// Отправка подтверждения
	client.sendMessage(fmt.Sprintf("%s%d;%s", msgConnected, client.id, client.GetPublicKeyPEM()))

	log.Printf("Клиент подключен: ID=%d, Name=%s", client.id, parts[1])
	s.mu.Lock()
	if s.clientPool.Size() > s.playersTotalPeak {
		s.playersTotalPeak = s.clientPool.Size()
		s.saveStats()
	}
	if s.clientPool.Size() > s.playersDayPeak {
		s.playersDayPeak = s.clientPool.Size()
		s.saveStats()
	}
	s.mu.Unlock()
	return client
}

func (s *Server) validateClientData(parts []string, fullData []byte) bool {
	sendedTime := parts[6]
	hash := parts[7]

	parsedTime, err := time.Parse("2006-01-02T15:04:05", sendedTime)
	if err != nil {
		log.Printf("Ошибка парсинга времени: %v", err)
		return false
	}

	// Проверка разницы во времени
	if time.Since(parsedTime).Abs() > 60*time.Second {
		log.Printf("Слишком большая разница во времени: %v", time.Since(parsedTime).Abs())
		return false
	}

	// Вычисляем хеш
	lastIndex := strings.LastIndex(string(fullData), ";")
	if lastIndex < 0 {
		return false
	}

	dataLine := string(fullData)[:lastIndex]
	dataHash := sha256.Sum256([]byte(dataLine))

	// Используем strings.Builder для эффективной конкатенации
	var sb strings.Builder
	sb.Grow(hex.EncodedLen(len(dataHash)) + len(salt1) + len(salt2))
	sb.WriteString(hex.EncodeToString(dataHash[:]))
	sb.WriteString(salt1)
	sb.WriteString(salt2)

	finalHash := sha256.Sum256([]byte(sb.String()))
	finalHashHex := hex.EncodeToString(finalHash[:])

	return finalHashHex == hash
}

func (s *Server) sendLowVersionResponse(conn *websocket.Conn) {
	msg := []byte(fmt.Sprintf("%s%s", msgLowVersion, need_version))
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Printf("Ошибка при отправке low_version: %v", err)
	}
	time.Sleep(500 * time.Millisecond)
}

// GetPublicKeyPEM возвращает публичный ключ в PEM формате
func (cm *Client) GetPublicKeyPEM() string {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(cm.PublicKey)
	if err != nil {
		log.Printf("Error marshaling public key: %v", err)
		return ""
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return string(pubKeyPEM)
}

func encryptMessage(publicKey *rsa.PublicKey, message string) ([]byte, error) {
	messageBytes := []byte(message)

	encrypted, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		messageBytes,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("ошибка шифрования: %w", err)
	}

	return encrypted, nil
}

func xorDecrypt(encryptedText []byte, fastXorKey string) string {
	keyBytes := []byte(fastXorKey)
	result := make([]byte, len(encryptedText))

	keyLen := len(keyBytes)
	for i := 0; i < len(encryptedText); i++ {
		result[i] = encryptedText[i] ^ keyBytes[i%keyLen]
	}

	return string(result)
}

//func (s *Server) handleMessages(client *Client) {
//	defer s.closeConnection(client.conn, client)
//
//	for {
//		select {
//		case <-s.shutdown:
//			return
//		default:
//			if err := s.processMessage(client); err != nil {
//				log.Printf("Ошибка при обработке сообщения от клиента %d: %v", client.id, err)
//				return
//			}
//		}
//	}
//}

func (s *Server) handleMessages(client *Client) {
	defer s.closeConnection(client.conn, client)

	for {
		select {
		case <-s.shutdown:
			return
		default:
			_, msgByte, err := client.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Ошибка чтения от клиента %d: %v", client.id, err)
				}
				return
			}

			client.lastActive = time.Now()

			go func(msg []byte) {
				msgText := s.decryptMessage(client, msg)
				for prefix, handler := range s.prefixToCommand {
					if strings.HasPrefix(msgText, prefix) {
						handler(client, msgText)
						return
					}
				}
				log.Printf("Неизвестная команда от клиента %d: %s", client.id, msgText)
			}(msgByte)
		}
	}
}

func (s *Server) processMessage(client *Client) error {
	// Устанавливаем таймаут на чтение
	if err := client.conn.SetReadDeadline(time.Now().Add(900 * time.Second)); err != nil {
		return err
	}

	_, msgByte, err := client.conn.ReadMessage()
	if err != nil {
		return err
	}

	// Обновляем время последней активности
	client.lastActive = time.Now()

	// Пытаемся расшифровать сообщение
	msg := s.decryptMessage(client, msgByte)

	// Поиск и выполнение команды
	for prefix, handler := range s.prefixToCommand {
		if strings.HasPrefix(msg, prefix) {
			handler(client, msg)
			return nil
		}
	}

	log.Printf("Неизвестная команда от клиента %d: %s", client.id, msg)
	return nil
}

func (s *Server) decryptMessage(client *Client, msgByte []byte) string {
	plain, err := rsa.DecryptPKCS1v15(rand.Reader, client.privateKey, msgByte)
	if err == nil {
		return string(plain)
	}
	return string(msgByte)
}

func (s *Server) closeConnection(conn *websocket.Conn, client *Client) {
	if client == nil {
		if conn != nil {
			conn.Close()
		}
		return
	}

	// Удаляем клиента из комнаты, если он там находится
	if client.currentRoom != 0 {
		s.leaveRoom(client, "")
	}

	// Удаляем из пула клиентов
	s.clientPool.Remove(client)

	// Закрываем соединение
	if err := conn.Close(); err != nil {
		log.Printf("Ошибка при закрытии соединения клиента %d: %v", client.id, err)
	}

	log.Printf("Клиент отключен: ID=%d, Name=%s", client.id, client.name)
}

func (client *Client) sendMessage(msg string) {
	if client == nil || client.conn == nil {
		return
	}

	if !client.server.clientPool.Exists(client) {
		return
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	// Устанавливаем таймаут на запись
	if err := client.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		log.Printf("Ошибка установки таймаута записи для клиента %d: %v", client.id, err)
		return
	}

	if err := client.conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		log.Printf("Ошибка при отправке сообщения клиенту %d: %v", client.id, err)
		client.server.clientPool.Remove(client)
		client.conn.Close()
	}
}

func (s *Server) broadcastToRoom(roomID int64, msg string, excludeClient *Client) {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.RLock()
	clients := make([]*Client, len(room.clients))
	copy(clients, room.clients)
	room.mu.RUnlock()

	for _, client := range clients {
		if client != excludeClient && s.clientPool.Exists(client) {
			client.sendMessage(msg)
		}
	}
}

// Оптимизированная функция удаления клиента из комнаты
func removeClientFromRoom(slice []*Client, value *Client) []*Client {
	if value != nil {
		value.currentRoom = 0
	}

	for i, v := range slice {
		if v == value {
			// Используем copy для эффективного удаления
			copy(slice[i:], slice[i+1:])
			return slice[:len(slice)-1]
		}
	}
	return slice
}

func main() {
	server := NewServer()

	http.HandleFunc("/ws", server.handleConnections)

	// Запуск горутины для очистки неактивных клиентов
	go server.cleanupInactiveClients()

	// Запуск горутины для обработки широковещательных сообщений
	go server.handleBroadcast()

	go server.resetDailyPeak()

	var networkIP string
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			networkIP = ipnet.IP.String()
			break
		}
	}

	if networkIP != "" {
		fmt.Printf("Сервер запущен на: %s:8085\n", networkIP)
	} else {
		fmt.Println("Сервер запущен на: http://localhost:8085")
	}

	err := http.ListenAndServe(":8085", nil)
	if err != nil {
		log.Fatal("Ошибка при запуске сервера:", err)
	}
}

// Очистка неактивных клиентов
func (s *Server) cleanupInactiveClients() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.removeInactiveClients()
		case <-s.shutdown:
			return
		}
	}
}

func (s *Server) removeInactiveClients() {
	var inactiveClients []*Client

	s.clientPool.mu.RLock()
	for client := range s.clientPool.clients {
		if time.Since(client.lastActive) > 10*time.Minute {
			inactiveClients = append(inactiveClients, client)
		}
	}
	s.clientPool.mu.RUnlock()

	for _, client := range inactiveClients {
		log.Printf("Удаление неактивного клиента: ID=%d, Name=%s", client.id, client.name)
		s.closeConnection(client.conn, client)
	}
}

func (s *Server) handleBroadcast() {
	for {
		select {
		case msg := <-s.broadcast:
			s.clientPool.mu.RLock()
			for client := range s.clientPool.clients {
				client.sendMessage(msg)
			}
			s.clientPool.mu.RUnlock()
		case <-s.shutdown:
			return
		}
	}
}
