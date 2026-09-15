package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var nextRoomId int64 = 1
var roomIdMutex sync.Mutex

// Константы для оптимизации строк
const (
	cmdCreateRoom     = "create_room"
	cmdEnterRoom      = "enter_room"
	cmdCurrentRoom    = "current_room"
	cmdGameStarted    = "game_started"
	cmdNeedReloadData = "need_reload_data"
	cmdPlayerLeftGame = "player_left_game"
	cmdPlayerJoined   = "player_joined"
	cmdErrorEnterRoom = "error_enter_room"
	cmdLeaveRoom      = "leave_room"
	cmdShowRooms      = "show_rooms"
)

// RoomUpdate представляет данные для обновления комнаты
type RoomUpdate struct {
	roomID   int64
	ownerID  int64
	mapID    string
	mapName  string
	gameMode string
	rules    string
	players  []PlayerInfo
	started  bool
	loaded   bool
}

type PlayerInfo struct {
	ID          int64
	Name        string
	Ready       bool
	Appearances string
	WeaponHand  string
	Voice       string
	Team        string
	IsFemale    string
}

func (s *Server) createRoom(client *Client, msg string) {
	logDebug("Создаём комнату: %s", msg)

	parts := strings.Split(msg, ";")
	if len(parts) < 8 {
		client.sendMessage("error_invalid_room_data")
		return
	}

	maxPlayers, err := strconv.ParseInt(parts[5], 10, 64)
	if err != nil || maxPlayers <= 0 {
		client.sendMessage("error_invalid_max_players")
		return
	}

	roomID := generateRoomID()

	newRoom := &Room{
		owner:      client,
		mapID:      parts[1],
		mapName:    parts[2],
		gameMode:   parts[3],
		rules:      parts[4],
		order:      parts[7],
		orderIndex: 0,
		maxPlayers: int(maxPlayers),
		password:   hashPassword(parts[6]),
		clients:    make([]*Client, 0, maxPlayers),
		excluded:   make(map[string]bool),
		mu:         sync.RWMutex{},
	}

	s.mu.Lock()
	s.rooms[roomID] = newRoom
	s.mu.Unlock()

	client.currentRoom = roomID
	newRoom.clients = append(newRoom.clients, client)
	client.ready = true

	client.sendMessage(fmt.Sprintf("%s%d", cmdCreateRoom, roomID))
}

func generateRoomID() int64 {
	roomIdMutex.Lock()
	defer roomIdMutex.Unlock()
	id := nextRoomId
	nextRoomId++
	return id
}

func (s *Server) changeRules(client *Client, msg string) {
	logDebug("Меняем правила комнаты: %s", msg)

	parts := strings.Split(msg, ";")
	if len(parts) < 8 {
		return
	}

	maxPlayers, err := strconv.ParseInt(parts[5], 10, 64)
	if err != nil {
		return
	}

	room := s.getRoomOrNull(client, client.currentRoom)
	if room == nil {
		return
	}

	room.mu.Lock()
	room.mapID = parts[1]
	room.mapName = parts[2]
	room.gameMode = parts[3]
	room.rules = parts[4]
	room.order = parts[7]
	room.orderIndex = 0
	room.maxPlayers = int(maxPlayers)
	room.password = hashPassword(parts[6])
	room.mu.Unlock()

	// Отправляем обновление всем в комнате
	s.broadcastToRoom(client.currentRoom, cmdNeedReloadData, nil)
}

func (s *Server) showRoom(client *Client, msg string) {
	logDebug("Показываем все комнаты")
	searchParams := []string{"", "any_mode", "any_map", "0", "0", "0"}
	client.sendMessage(s.getRoomList(client, searchParams))
}

func (s *Server) searchRooms(client *Client, msg string) {
	logDebug("Ищем комнаты: %s", msg)
	searchParams := strings.Split(msg, ";")
	client.sendMessage(s.getRoomList(client, searchParams))
}

func (s *Server) getRoom(client *Client, msg string) {
	searchParams := strings.Split(msg, ":")[1]
	roomID, _ := strconv.ParseInt(searchParams, 10, 64)
	room := s.getRoomOrNull(client, roomID)
	if room == nil {
		client.sendMessage(fmt.Sprintf("got_roomno"))
		return
	}
	if checkExcluded(client, room) {
		client.sendMessage(fmt.Sprintf("got_roomno"))
		return
	}
	if client.demoVersion && !passDemoRooms(room) {
		client.sendMessage(fmt.Sprintf("got_roomno"))
	} else {
		client.sendMessage(fmt.Sprintf("got_room%d;%d;%d;%s;%s;%t;%s",
			roomID,
			room.maxPlayers,
			len(room.clients),
			room.gameMode,
			room.mapID,
			room.started,
			room.password,
		))
	}

}

func (s *Server) leaveRoom(client *Client, msg string) {
	logDebug("Выходим из комнаты")

	roomID := client.currentRoom
	if roomID == 0 {
		return
	}

	room := s.getRoomOrNull(client, roomID)
	if room == nil {
		return
	}

	room.mu.Lock()

	// Удаляем клиента из комнаты
	room.clients = removeClientFromRoom(room.clients, client)
	wasOwner := client == room.owner

	// Проверяем, нужно ли сменить владельца
	if wasOwner && len(room.clients) > 1 {
		if room.clients[0].is_authoian == "true" {
			room.owner = room.clients[1]
		} else {
			room.owner = room.clients[0]
		}
	}

	room.mu.Unlock()

	// Если комната пуста, удаляем её
	if len(room.clients) <= 1 && room.was_started || len(room.clients) == 0 && !room.was_started {
		s.mu.Lock()
		delete(s.rooms, roomID)
		s.mu.Unlock()
		if len(room.clients) <= 1 && room.was_started {
			room.clients[0].sendMessage("close_window")
		}
		logDebug("Удалена пустая комната %d", roomID)
	} else if room.started {
		// Отправляем уведомление об уходе игрока
		msg := fmt.Sprintf("%s%d", cmdPlayerLeftGame, client.id)
		if wasOwner && len(room.clients) > 0 {
			msg = fmt.Sprintf("%s%d %d", cmdPlayerLeftGame, client.id, room.owner.id)
		}
		s.broadcastToRoom(roomID, msg, nil)
	} else {
		// Обновляем информацию о комнате для всех
		s.fullRoomUpdate(room)
	}

	client.sendMessage(cmdLeaveRoom)
}

func (s *Server) leaveGame(client *Client, msg string) {
	logDebug("Покидаем игру")

	room := s.getRoomOrNull(client, client.currentRoom)
	if room == nil {
		return
	}

	client.ready = false
	s.broadcastToRoom(client.currentRoom, fmt.Sprintf("%s%d", cmdPlayerLeftGame, client.id), client)
}

func (s *Server) enterRoom(client *Client, msg string) {
	logDebug("Входим в комнату: %s", msg)

	if client.currentRoom != 0 {
		client.sendMessage(fmt.Sprintf("Вы уже находитесь в комнате %d", client.currentRoom))
		return
	}

	parts := strings.Split(msg, ":")
	if len(parts) < 2 {
		client.sendMessage("error_invalid_room_code")
		return
	}

	roomID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		client.sendMessage("error_invalid_room_code")
		return
	}

	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	if !exists {
		client.sendMessage(fmt.Sprintf("%s_exist", cmdErrorEnterRoom))
		return
	}

	room.mu.RLock()
	currentPlayers := len(room.clients)
	maxPlayers := room.maxPlayers
	hasPassword := room.password != ""
	room.mu.RUnlock()

	if currentPlayers >= maxPlayers {
		client.sendMessage(fmt.Sprintf("%s_full", cmdErrorEnterRoom))
		return
	}

	// Проверка пароля
	if hasPassword && client.is_authoian == "false" {
		if len(parts) < 3 || parts[2] == "" {
			client.sendMessage(fmt.Sprintf("%s_passworded", cmdErrorEnterRoom))
			return
		}
		if parts[2] != room.password {
			client.sendMessage(fmt.Sprintf("%s_wrong_password", cmdErrorEnterRoom))
			return
		}
	}

	// Добавляем клиента в комнату
	client.currentRoom = roomID
	client.ready = false
	client.team = "-1"

	room.mu.Lock()
	room.clients = append(room.clients, client)
	room.mu.Unlock()

	client.sendMessage(fmt.Sprintf("%s%d", cmdEnterRoom, roomID))
	s.fullRoomUpdate(room)
}

func (s *Server) currentRoom(client *Client, msg string) {
	logDebug("%s Запрос информации о комнате", client.name)

	room := s.getRoomOrNull(client, client.currentRoom)
	if room == nil {
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	// Собираем информацию об игроках
	var playersBuilder strings.Builder
	for _, resident := range room.clients {
		ready := "X"
		if resident.ready {
			ready = "V"
		}

		playersBuilder.WriteString(fmt.Sprintf("%s%s*%d*%s*%s*%s*%s*%s*%s*%s*%s*%s*%s*%s*%s|",
			ready,
			resident.name,
			resident.id,
			resident.appearances,
			resident.weaponHand,
			resident.voice,
			resident.team,
			resident.isFemale,
			resident.colors,
			resident.border_index,
			resident.border_color,
			resident.hands_index,
			resident.hands_color1,
			resident.hands_color2,
			resident.is_authoian,
		))
	}

	// Формируем ответ
	response := fmt.Sprintf("%s%d;%d;%s;%s;%s;%s;%s;%t;%d;%t;%s;%d",
		cmdCurrentRoom,
		client.currentRoom,
		room.owner.id,
		room.mapID,
		room.mapName,
		room.gameMode,
		room.rules,
		playersBuilder.String(),
		room.started,
		client.currentRoom,
		room.loaded,
		room.order,
		room.orderIndex,
	)

	client.sendMessage(response)
}

func (s *Server) switchReady(client *Client, msg string) {
	logDebug("%s Меняет статус готовности", client.name)

	room := s.getRoomOrNull(client, client.currentRoom)
	if room == nil {
		return
	}

	client.ready = !client.ready
	s.fullRoomUpdate(room)
}

func (s *Server) startGame(client *Client, msg string) {
	logDebug("Игра начинается")

	room := s.getRoomOrNull(client, client.currentRoom)
	if room == nil {

		return
	}

	room.mu.RLock()
	isOwner := room.owner == client
	room.mu.RUnlock()

	if !isOwner {

		return
	}

	room.mu.Lock()
	room.started = true
	room.was_started = true
	room.loaded = true
	room.mu.Unlock()

	// Отправляем уведомления игрокам
	room.mu.RLock()
	clients := make([]*Client, len(room.clients))
	copy(clients, room.clients)
	room.mu.RUnlock()

	//cmd := exec.Command(".\\lfl.exe", "--headless", "--", "autority"+strconv.FormatInt(client.currentRoom, 10))
	logDebug("Включаем хоста-авторитета")
	cmd := exec.Command("./lfl.x86_64", "--headless", "--", "autority"+strconv.FormatInt(client.currentRoom, 10))

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	logDebug("Хост-авторитет включен")
	for _, resident := range clients {
		if resident.ready || resident == client {
			resident.ready = true
			resident.sendMessage(cmdGameStarted)
		} else {
			resident.sendMessage(cmdNeedReloadData)
		}
	}
}

func (s *Server) joinGame(client *Client, msg string) {
	logDebug("Присоединяемся к игре")

	room := s.getRoomOrNull(client, client.currentRoom)
	if room == nil || (!room.loaded && client.is_authoian == "false") {
		return
	}

	client.ready = true
	client.sendMessage(cmdGameStarted)

	// Формируем данные игрока
	playerData := fmt.Sprintf("%s*%d*%s*%s*%s*%s*%s*%s*%s*%s*%s*%s*%s",
		client.name,
		client.id,
		client.appearances,
		client.weaponHand,
		client.voice,
		client.isFemale,
		client.team,
		client.colors,
		client.border_index,
		client.border_color,
		client.hands_index,
		client.hands_color1,
		client.hands_color2,
	)

	// Уведомляем других игроков
	room.mu.RLock()
	clients := make([]*Client, len(room.clients))
	copy(clients, room.clients)
	room.mu.RUnlock()

	for _, resident := range clients {
		if resident == client {
			continue
		}
		if resident.ready {
			resident.sendMessage(fmt.Sprintf("%s%s", cmdPlayerJoined, playerData))
		} else {
			resident.sendMessage(cmdNeedReloadData)
		}
	}
}

func (s *Server) endGame(client *Client, msg string) {
	isForce := msg == "end_game_force"
	code := "game_ended"

	if isForce {
		logDebug("Игра заканчивается досрочно")
		code = "game_ended_force"
	} else {
		logDebug("Игра заканчивается")
	}

	room := s.getRoomOrNull(client, client.currentRoom)
	if room == nil {
		return
	}

	room.mu.RLock()
	isOwner := room.owner == client
	room.mu.RUnlock()

	if !isOwner {
		return
	}

	room.mu.Lock()
	room.started = false
	room.was_started = false
	if isForce {
		room.orderIndex = 0
	}
	room.mu.Unlock()

	// Уведомляем игроков
	room.mu.RLock()
	clients := make([]*Client, len(room.clients))
	copy(clients, room.clients)
	room.mu.RUnlock()

	for _, resident := range clients {
		if (resident.ready || resident == client) && isForce {
			resident.ready = false
			resident.sendMessage(code)
		} else {
			resident.sendMessage(cmdNeedReloadData)
		}
	}
}

func (s *Server) restartGame(client *Client, msg string) {
	parts := strings.Split(msg, ";")
	nextMapOrder, err := strconv.Atoi(parts[1])
	if err != nil {
		client.sendMessage("Мало данных")
		return
	}
	logDebug("Перезапуск игры:")
	// Отправляем данные всем готовым игрокам
	roomID := client.currentRoom
	if roomID == 0 {
		client.sendMessage("Вы не находитесь в комнате!")
		return
	}
	room := s.getRoomOrNull(client, roomID)
	room.mu.RLock()
	clients := make([]*Client, 0, len(room.clients))
	copy(clients, room.clients)
	room.mu.RUnlock()
	room.mu.Lock()
	room.started = true
	room.loaded = true
	if nextMapOrder > -1 {
		room.orderIndex = nextMapOrder
		room.mapID = strings.Split(room.order, ":")[nextMapOrder]
	}
	room.mu.Unlock()

	for _, resident := range room.clients {
		if resident.ready || resident == client {
			resident.ready = true

			resident.sendMessage(fmt.Sprintf("drop_and_start_game%d", nextMapOrder))
		} else {
			resident.sendMessage(cmdNeedReloadData)
		}
	}
}

func (s *Server) sendTime(client *Client, msg string) {
	logDebug("Ведём отсчёт: %s", msg)

	parts := strings.Split(msg, ";")
	if len(parts) < 2 {
		return
	}

	newTime, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}

	s.broadcastToRoom(client.currentRoom, fmt.Sprintf("new_time%d", newTime), nil)
}

func (s *Server) getRoomOrNull(client *Client, roomID int64) *Room {
	if roomID == 0 {
		client.sendMessage("Вы не находитесь в комнате")
		return nil
	}

	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	if !exists {
		client.currentRoom = 0
		client.sendMessage(fmt.Sprintf("Комнаты с указанным кодом %d уже не существует!", roomID))
		return nil
	}

	return room
}

func hashPassword(password string) string {
	if password == "" {
		return ""
	}

	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func checkExcluded(client *Client, room *Room) bool {
	room.mu.RLock()
	defer room.mu.RUnlock()

	_, excluded := room.excluded[client.unique_id]
	return excluded
}

func (s *Server) getRoomList(client *Client, searchParams []string) string {
	if len(searchParams) < 5 {
		return cmdShowRooms
	}

	free := searchParams[4]
	status := searchParams[3]
	gameMap := searchParams[2]
	gameMode := searchParams[1]
	searchPage, _ := strconv.ParseInt(searchParams[5], 10, 64)

	var builderMaps strings.Builder

	s.mu.RLock()
	rooms := make(map[int64]*Room, len(s.rooms))
	for id, room := range s.rooms {
		rooms[id] = room
	}
	s.mu.RUnlock()

	keys := make([]int64, 0, len(rooms))
	for k := range rooms {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	limit := int64(22)
	start_index := searchPage * limit
	index := int64(0)

	for _, id := range keys {
		if start_index > 0 {
			start_index -= 1
			continue
		}
		room := rooms[id]
		if checkExcluded(client, room) {
			continue
		}

		room.mu.RLock()
		roomGameMode := room.gameMode
		roomMapID := room.mapID
		roomStarted := room.started
		currentPlayers := len(room.clients)
		if roomStarted {
			currentPlayers -= 1
		}
		maxPlayers := room.maxPlayers
		room.mu.RUnlock()

		//Комната только для полной версии
		if client.demoVersion && !passDemoRooms(room) || len(room.clients) <= 0 {
			continue
		}

		// Применяем фильтры
		if gameMode != "any_mode" && gameMode != roomGameMode {
			continue
		}
		if gameMap != "any_map" && gameMap != roomMapID {
			continue
		}
		if status == "1" && roomStarted {
			continue
		}
		if status == "2" && !roomStarted {
			continue
		}
		if free == "1" && currentPlayers >= maxPlayers {
			continue
		}
		if index < limit {
			builderMaps.WriteString(fmt.Sprintf("\n%d;%d;%d;%s;%s;%t;%s",
				id,
				maxPlayers,
				currentPlayers,
				roomGameMode,
				roomMapID,
				roomStarted,
				room.password,
			))
		}
		index++
	}

	var builder strings.Builder
	builder.WriteString(cmdShowRooms)

	builder.WriteString(fmt.Sprintf("%d;%d;%d;%d;%d;%d%s",
		len(s.rooms),
		s.clientPool.Size(),
		s.playersDayPeak,
		s.playersTotalPeak,
		searchPage,
		index+searchPage*limit,
		builderMaps.String(),
	))

	return builder.String()
}

func passDemoRooms(room *Room) bool {
	if isAllowedDemoMode(room.gameMode) && isAllowedDemoMap(room.mapID) {
		return true
	}
	return false
}

func isAllowedDemoMode(mode string) bool {
	var allowedGameModes [4]string = [4]string{"OneForOne", "LastHero", "TeamOnTeam", "CaptureFlag"}
	for _, m := range allowedGameModes {
		if m == mode {
			return true
		}
	}
	return false
}

func isAllowedDemoMap(mapName string) bool {
	var allowedGameModes [9]string = [9]string{"PassangerShip", "Park", "Dungeons", "Asteroid", "BadBlock", "Drednout", "GreatCathedral", "Library", "OceanStation"}
	for _, m := range allowedGameModes {
		if m == mapName {
			return true
		}
	}
	return false
}

// Вспомогательная функция для отладки
func logDebug(format string, args ...interface{}) {
	if true { // Замените на условную компиляцию или флаг
		fmt.Printf(format+"\n", args...)
	}
}
