package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/pion/turn/v4"
)

// Вспомогательная функция для создания UDP слушателя
func createUDPListener(port int) net.PacketConn {
	conn, err := net.ListenPacket("udp4", "0.0.0.0:"+strconv.Itoa(port))
	if err != nil {
		log.Printf("Не удалось создать слушатель: %s", err)
	} else {
		log.Printf("Cоздан слушатель: %s", conn)
	}
	return conn
}

var IP = "0.0.0.0"

// Секретный ключ для генерации паролей (храните в безопасности!)
const turnSecretKey = "your-super-secret-key-change-this"

func (ms *Server) startTurnServerSimple(publicIP string, port int) {
	// Статические пользователи (сгенерируйте один раз)
	realm := "lifefirelea-tuRN.com"

	authHandler := func(username string, realm string, srcAddr net.Addr) ([]byte, bool) {
		log.Printf("TURN аутентификация: username=%s, realm=%s", username, realm)
		if key, ok := ms.turnCredentials[username]; ok {
			return key, true
		}
		return nil, false
	}

	s, err := turn.NewServer(turn.ServerConfig{
		Realm:       realm,
		AuthHandler: authHandler, // Теперь типы совпадают
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: createUDPListener(port),
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: net.ParseIP(publicIP),
					Address:      "0.0.0.0",
				},
			},
		},
	})
	if err != nil {
		log.Panicf("Ошибка: %s", err)
	}
	defer s.Close()
	log.Printf("✅ TURN сервер запущен на %s:%d", publicIP, port)
	select {}
}

// Генерирует временный пароль на основе username
func (ms *Server) generateTurnPassword(username string) string {
	h := hmac.New(sha1.New, []byte(turnSecretKey))
	h.Write([]byte(username))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Генерирует имя пользователя с временной меткой (истекает через 24 часа)
func (ms *Server) generateTurnUsername(originalUserID string) string {
	expires := time.Now().Add(24 * time.Hour).Unix()
	return fmt.Sprintf("%d:%s", expires, originalUserID)
}

// ГЛАВНАЯ ФУНКЦИЯ: генерирует пару username/password для клиента
func (ms *Server) generateTurnCredentials(userID string) (string, string) {
	username := ms.generateTurnUsername(userID)
	password := ms.generateTurnPassword(username)
	ms.turnCredentials[username] = turn.GenerateAuthKey(username, "lifefirelea-tuRN.com", password)
	return username, password
}
