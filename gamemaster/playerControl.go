package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Константы для команд
const (
	cmdSendCommand      = "send_command"
	cmdGDZ              = "\u0003   gdz"
	cmdGD               = "gd"
	cmdSay              = "say"
	cmdExcluded         = "excluded"
	cmdRequestGameData  = "request_game_data"
	cmdGameDataSend     = "game_datasend"
	cmdSwitchTeamInGame = "switch_team_in_game"

	// Цветовые теги для команд
	colorRed   = "[color=#ff0000]"
	colorBlue  = "[color=#0000ff]"
	colorClose = "[/color]"
)

// Предварительно вычисленные цветовые префиксы
var (
	teamColorPrefix = map[string]string{
		"0": colorRed,
		"1": colorBlue,
	}
)

func (s *Server) returnToRoom(client *Client, msg string) {
	roomID := client.currentRoom
	if roomID == 0 {
		client.sendMessage("Вы не находитесь в комнате!")
		return
	}
	room := s.getRoomOrNull(client, roomID)
	if room == nil {
		return
	}
	if client == room.owner {
		room.mu.RLock()
		clients := make([]*Client, len(room.clients))
		copy(clients, room.clients)
		room.orderIndex = 0
		room.mu.RUnlock()

		for _, resident := range clients {
			resident.ready = false
		}
	}
	client.ready = false
	s.broadcastToRoom(roomID, cmdNeedReloadData, nil)
}

func (s *Server) sendCommand(client *Client, msg string) {
	roomID := client.currentRoom
	if roomID == 0 {
		client.sendMessage("Вы не находитесь в комнате!")
		return
	}

	room := s.getRoomOrNull(client, roomID)
	if room == nil {
		return
	}

	// Быстрое извлечение команды
	command := msg[len("command"):]
	room.owner.sendMessage(fmt.Sprintf("%s%d %s", cmdSendCommand, client.id, command))
}

func (s *Server) setRoomLoaded(client *Client, msg string) {
	roomID := client.currentRoom
	if roomID == 0 {
		client.sendMessage("Вы не находитесь в комнате!")
		return
	}

	room := s.getRoomOrNull(client, roomID)
	if room == nil {
		return
	}

	// Проверяем, является ли клиент владельцем
	if room.owner != client {
		return
	}

	room.mu.Lock()
	room.loaded = true
	room.mu.Unlock()

	// Отправляем уведомление всем в комнате
	s.broadcastToRoom(roomID, cmdNeedReloadData, nil)
}

func (s *Server) sendData(client *Client, msg string) {
	roomID := client.currentRoom
	if roomID == 0 {
		client.sendMessage("Вы не находитесь в комнате!")
		return
	}

	room := s.getRoomOrNull(client, roomID)
	if room == nil {
		return
	}

	// Только владелец может отправлять данные
	if room.owner != client {
		return
	}

	// Извлекаем данные
	data := msg[len("gd"):]

	room.mu.RLock()
	clients := make([]*Client, 0, len(room.clients))
	for _, resident := range room.clients {
		if resident != room.owner && resident.ready {
			clients = append(clients, resident)
		}
	}
	room.mu.RUnlock()

	// Отправляем данные всем готовым игрокам
	for _, resident := range clients {
		resident.sendMessage(fmt.Sprintf("%s%s", cmdGD, data))
	}
}

// sendMessageCheck проверяет, может ли клиент получить сообщение
func sendMessageCheck(sender, receiver *Client, room *Room) bool {
	if !room.started {
		return true
	}
	return sender.ready == receiver.ready
}

func (s *Server) sayRoom(client *Client, msg string) {
	message := msg[len("say_room"):]
	room := s.getRoomOrNull(client, client.currentRoom)
	if room == nil {
		return
	}

	// Получаем цветное имя отправителя
	senderName := getColoredName(client)
	formattedMsg := fmt.Sprintf("%s%s: %s", cmdSay, senderName, message)

	room.mu.RLock()
	clients := make([]*Client, len(room.clients))
	copy(clients, room.clients)
	room.mu.RUnlock()

	// Отправляем сообщение всем, кто может его получить
	for _, resident := range clients {
		if sendMessageCheck(client, resident, room) {
			resident.sendMessage(formattedMsg)
		}
	}
}

func (s *Server) sayTeam(client *Client, msg string) {
	message := msg[len("say_team"):]
	room := s.getRoomOrNull(client, client.currentRoom)
	if room == nil {
		return
	}

	// Получаем цветное имя отправителя
	senderName := getColoredName(client)
	formattedMsg := fmt.Sprintf("%s%s: %s", cmdSay, senderName, message)
	clientTeam := client.team

	room.mu.RLock()
	clients := make([]*Client, 0, len(room.clients))
	for _, resident := range room.clients {
		if resident.team == clientTeam {
			clients = append(clients, resident)
		}
	}
	room.mu.RUnlock()

	// Отправляем сообщение только игрокам той же команды
	for _, resident := range clients {
		resident.sendMessage(formattedMsg)
	}
}

func (s *Server) sayPersonal(client *Client, msg string) {
	message := msg[len("say_personal"):]
	parts := strings.SplitN(message, ";", 3)
	if len(parts) < 3 {
		return
	}

	targetID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}

	personalMsg := parts[2]
	room := s.getRoomOrNull(client, client.currentRoom)
	if room == nil {
		return
	}

	senderName := getColoredName(client)

	room.mu.RLock()
	var targetClient *Client
	for _, resident := range room.clients {
		if resident.id == targetID {
			targetClient = resident
			break
		}
	}
	room.mu.RUnlock()

	if targetClient == nil {
		return
	}

	targetName := getColoredName(targetClient)
	formattedMsg := fmt.Sprintf("%s%s (%s): %s", cmdSay, senderName, targetName, personalMsg)

	// Отправляем сообщение целевому игроку
	targetClient.sendMessage(formattedMsg)

	// Отправляем копию отправителю, если это не он сам
	if targetClient != client {
		client.sendMessage(formattedMsg)
	}
}

func (s *Server) excludePlayer(client *Client, msg string) {
	logDebug("Исключение игрока")

	playerIDStr := msg[len("exclude"):]
	playerID, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		return
	}

	room := s.getRoomOrNull(client, client.currentRoom)
	if room == nil {
		return
	}

	// Только владелец может исключать игроков
	if room.owner != client {
		return
	}

	room.mu.Lock()

	// Ищем игрока для исключения
	var targetClient *Client
	var targetIndex = -1
	for i, resident := range room.clients {
		if resident.id == playerID {
			targetClient = resident
			targetIndex = i
			break
		}
	}

	if targetClient == nil {
		room.mu.Unlock()
		return
	}

	// Добавляем в исключенные (используем map для O(1) доступа)
	if room.excluded == nil {
		room.excluded = make(map[*Client]bool)
	}
	room.excluded[targetClient] = true

	// Удаляем из списка клиентов
	if targetIndex >= 0 {
		copy(room.clients[targetIndex:], room.clients[targetIndex+1:])
		room.clients = room.clients[:len(room.clients)-1]
	}

	room.mu.Unlock()

	// Уведомляем исключенного игрока
	targetClient.currentRoom = 0
	targetClient.sendMessage(cmdExcluded)

	// Обновляем комнату для оставшихся игроков
	s.fullRoomUpdate(room)
}

func (s *Server) gameDataRequest(client *Client, msg string) {
	logDebug("Запрос на игровые данные")

	room := s.getRoomOrNull(client, client.currentRoom)
	if room == nil {
		return
	}

	// Отправляем запрос владельцу комнаты
	room.owner.sendMessage(fmt.Sprintf("%s%d", cmdRequestGameData, client.id))
}

func (s *Server) gameDataSend(client *Client, msg string) {
	data := msg[len("send_game_data"):]
	parts := strings.SplitN(data, ";", 2)
	if len(parts) < 2 {
		return
	}

	targetID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return
	}

	gameData := parts[1]

	room := s.getRoomOrNull(client, client.currentRoom)
	if room == nil {
		return
	}

	room.mu.RLock()
	var targetClient *Client
	for _, resident := range room.clients {
		if resident.id == targetID {
			targetClient = resident
			break
		}
	}
	room.mu.RUnlock()

	if targetClient == nil {
		return
	}

	logDebug("Игровые данные отправлены")
	targetClient.sendMessage(fmt.Sprintf("%s%s", cmdGameDataSend, gameData))
}

func (s *Server) switchTeam(client *Client, msg string) {
	logDebug("%s Смена команды", client.name)

	parts := strings.Split(msg, " ")
	if len(parts) < 3 {
		return
	}

	playerID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}

	newTeam := parts[2]

	room := s.getRoomOrNull(client, client.currentRoom)
	if room == nil {
		return
	}

	// Проверяем права на смену команды
	canSwitch := room.owner == client || playerID == client.id
	if !canSwitch {
		return
	}

	room.mu.Lock()
	// Обновляем команду игрока
	for _, resident := range room.clients {
		if resident.id == playerID {
			resident.team = newTeam
			break
		}
	}
	room.mu.Unlock()

	// Обновляем комнату для всех
	s.fullRoomUpdate(room)
}

func (s *Server) sendSwitchTeamInGame(client *Client, msg string) {
	logDebug("%s Смена команды в игре", client.name)

	parts := strings.Split(msg, " ")
	if len(parts) < 3 {
		return
	}

	playerID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}

	newTeam := parts[2]

	room := s.getRoomOrNull(client, client.currentRoom)
	if room == nil {
		return
	}

	// Проверяем права на смену команды
	canSwitch := room.owner == client || playerID == client.id
	if !canSwitch {
		return
	}

	room.mu.Lock()
	// Обновляем команду игрока
	for _, resident := range room.clients {
		if resident.id == playerID {
			resident.team = newTeam
			break
		}
	}
	room.mu.Unlock()

	// Уведомляем всех в комнате о смене команды
	s.broadcastToRoom(client.currentRoom, fmt.Sprintf("%s%d", cmdSwitchTeamInGame, playerID), nil)
}

// getColoredName возвращает цветное имя игрока на основе его команды
func getColoredName(client *Client) string {
	name := client.name

	// Используем map для быстрого получения цвета
	if prefix, ok := teamColorPrefix[client.team]; ok {
		return prefix + name + colorClose
	}

	return name
}

// fullRoomUpdate обновляет информацию о комнате для всех игроков
func (s *Server) fullRoomUpdate(room *Room) {
	if room == nil {
		return
	}

	room.mu.RLock()
	clients := make([]*Client, len(room.clients))
	copy(clients, room.clients)
	room.mu.RUnlock()

	for _, client := range clients {
		client.sendMessage(cmdNeedReloadData)
	}
}

func (s *Server) sendOffer(client *Client, msg string) {
	logDebug("Отправлено предложение: %s", msg)

	parts := strings.Split(msg, ";")
	player_id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}

	player := s.clientPool.Get(player_id)
	player.sendMessage(fmt.Sprintf("%s", msg))
}

func (s *Server) sendAnswer(client *Client, msg string) {
	logDebug("Отправлен ответ: %s", msg)

	parts := strings.Split(msg, ";")
	player_id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}

	player := s.clientPool.Get(player_id)
	player.sendMessage(fmt.Sprintf("%s", msg))
}

func (s *Server) sendCandidate(client *Client, msg string) {
	logDebug("Отправлены кандидаты: %s", msg)

	parts := strings.Split(msg, ";")
	player_id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}

	player := s.clientPool.Get(player_id)
	player.sendMessage(fmt.Sprintf("%s", msg))
}

func (s *Server) sendConnection(client *Client, msg string) {
	logDebug("Проводим соединение с: %s", msg)

	parts := strings.Split(msg, ";")
	player_id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}

	player := s.clientPool.Get(player_id)
	player.sendMessage(fmt.Sprintf("%s", msg))
}
