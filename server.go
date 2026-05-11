package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed frontend/dist
var embeddedFrontend embed.FS

// ─── PASSAGES ──────────────────────────────────────────────────────────────

var passages = []string{
	"The quick brown fox jumps over the lazy dog near the bank of the river. A gentle breeze carries the scent of wildflowers across the meadow as the sun begins to set behind the distant mountains. Birds return to their nests while crickets start their evening chorus.",
	"In the world of technology, change is the only constant. Every day brings new innovations that reshape how we live, work, and connect with one another. The pace of progress can be breathtaking, and those who embrace it find themselves at the forefront of tomorrow's discoveries.",
	"The road stretched ahead like a ribbon of asphalt threading through the countryside. Tall trees lined both sides, their branches forming a natural canopy that filtered the morning sunlight into dancing patterns on the ground below.",
	"She opened the old leather-bound book carefully, the pages yellowed with age. Each chapter revealed a world long forgotten, filled with characters whose stories echoed through time. Reading felt like stepping into a conversation across centuries.",
	"The stadium roared with excitement as the race began. Engines revved in harmony, tires screeched against the track, and the smell of burning rubber filled the air. Drivers pushed their machines to the limit, each curve a test of skill and courage.",
	"Programming is the art of telling a computer what to do through precise instructions. It combines logical thinking with creative problem solving, allowing developers to build everything from simple scripts to complex systems.",
	"A wise person once said that the journey of a thousand miles begins with a single step. Every great achievement starts with the decision to try, to take that first step into the unknown. The path may be uncertain, but the destination makes it worthwhile.",
	"Under the canopy of stars, the desert stretched endlessly in all directions. The silence was profound, broken only by the occasional howl of a distant coyote. Campfire light flickered against weathered faces around the warmth.",
	"Music has the power to transport us to another time and place. A single melody can evoke memories long buried, stir emotions we forgot we had, and connect strangers across languages and cultures in a shared moment of understanding.",
	"The ocean waves crashed against the shore with rhythmic precision, each one unique in its formation yet part of an eternal pattern. Seashells scattered across the sand like treasures waiting to be discovered.",
}

func randomPassage() string {
	return passages[rand.Intn(len(passages))]
}

// ─── TYPES ─────────────────────────────────────────────────────────────────

type Player struct {
	Name      string  `json:"name"`
	Color     string  `json:"color"`
	Progress  float64 `json:"progress"`
	WPM       int     `json:"wpm"`
	Accuracy  float64 `json:"accuracy"`
	Finished  bool    `json:"finished"`
	CharsTyped int    `json:"charsTyped"`
	FinishTime int64  `json:"finishTime,omitempty"`
	Place     int     `json:"place,omitempty"`
}

type Room struct {
	sync.Mutex
	ID        string
	Passage   string
	State     string // waiting | racing | finished
	Players   map[*websocket.Conn]*Player
	Finishers []string
	StartTime int64
}

type Message struct {
	Type        string    `json:"type"`
	RoomID      string    `json:"roomId,omitempty"`
	PlayerName  string    `json:"playerName,omitempty"`
	PlayerColor string    `json:"playerColor,omitempty"`
	Passage     string    `json:"passage,omitempty"`
	Players     []*Player `json:"players,omitempty"`
	Name        string    `json:"name,omitempty"`
	RoomId      string    `json:"roomId,omitempty"`
	Progress    float64   `json:"progress,omitempty"`
	WPM         int       `json:"wpm,omitempty"`
	Accuracy    float64   `json:"accuracy,omitempty"`
	CharsTyped  int       `json:"charsTyped,omitempty"`
	Finished    bool      `json:"finished,omitempty"`
	Place       int       `json:"place,omitempty"`
	Message     string    `json:"message,omitempty"`
	StartTime   int64     `json:"startTime,omitempty"`
	Rooms       []RoomSum `json:"rooms,omitempty"`
}

type RoomSum struct {
	ID      string `json:"id"`
	Players int    `json:"players"`
	Passage string `json:"passage"`
}

// ─── SERVER STATE ──────────────────────────────────────────────────────────

var (
	rooms      = make(map[string]*Room)
	roomsMu    sync.Mutex
	colors     = []string{"#ff4757", "#2ed573", "#1e90ff", "#ffa502", "#a55eea", "#ff6b81", "#00d2d3", "#f368e0"}
	colorIdx   int
	upgrader   = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func nextColor() string {
	c := colors[colorIdx%len(colors)]
	colorIdx++
	return c
}

func genRoomID() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 4)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func getPlayers(room *Room) []*Player {
	room.Lock()
	defer room.Unlock()
	list := make([]*Player, 0, len(room.Players))
	for _, p := range room.Players {
		list = append(list, p)
	}
	return list
}

// ─── WEBSOCKET ─────────────────────────────────────────────────────────────

func handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}
	defer conn.Close()

	var (
		currentRoom   *Room
		currentRoomID string
		playerName    string
		playerColor   string
	)

	sendJSON := func(msg Message) {
		data, _ := json.Marshal(msg)
		conn.WriteMessage(websocket.TextMessage, data)
	}

	broadcast := func(msg Message, exclude *websocket.Conn) {
		if currentRoom == nil {
			return
		}
		data, _ := json.Marshal(msg)
		currentRoom.Lock()
		for c := range currentRoom.Players {
			if c != exclude {
				c.WriteMessage(websocket.TextMessage, data)
			}
		}
		currentRoom.Unlock()
	}

	sendJSON(Message{Type: "connected"})

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}

		switch msg.Type {

		case "create_room":
			playerName = msg.Name
			if playerName == "" {
				playerName = "Anonymous"
			}
			playerColor = nextColor()

			roomID := genRoomID()
			room := &Room{
				ID:        roomID,
				Passage:   randomPassage(),
				State:     "waiting",
				Players:   make(map[*websocket.Conn]*Player),
				Finishers: []string{},
			}
			room.Players[conn] = &Player{
				Name:     playerName,
				Color:    playerColor,
				Progress: 0,
				WPM:      0,
				Accuracy: 100,
				Finished: false,
			}

			roomsMu.Lock()
			rooms[roomID] = room
			roomsMu.Unlock()

			currentRoom = room
			currentRoomID = roomID

			sendJSON(Message{
				Type:        "room_joined",
				RoomID:      roomID,
				PlayerName:  playerName,
				PlayerColor: playerColor,
				Passage:     room.Passage,
				Players:     getPlayers(room),
			})

		case "join_room":
			roomID := strings.ToUpper(msg.RoomId)
			roomsMu.Lock()
			room, exists := rooms[roomID]
			roomsMu.Unlock()

			if !exists {
				sendJSON(Message{Type: "error", Message: "Room not found"})
				continue
			}
			if room.State != "waiting" {
				sendJSON(Message{Type: "error", Message: "Race already in progress"})
				continue
			}

			playerName = msg.Name
			if playerName == "" {
				playerName = "Anonymous"
			}
			playerColor = nextColor()

			room.Lock()
			room.Players[conn] = &Player{
				Name:     playerName,
				Color:    playerColor,
				Progress: 0,
				WPM:      0,
				Accuracy: 100,
				Finished: false,
			}
			room.Unlock()

			currentRoom = room
			currentRoomID = roomID

			sendJSON(Message{
				Type:        "room_joined",
				RoomID:      roomID,
				PlayerName:  playerName,
				PlayerColor: playerColor,
				Passage:     room.Passage,
				Players:     getPlayers(room),
			})

			broadcast(Message{
				Type:    "player_joined",
				Players: getPlayers(room),
			}, conn)

		case "start_race":
			if currentRoom == nil {
				continue
			}
			currentRoom.Lock()
			currentRoom.State = "racing"
			startTime := time.Now().UnixMilli() + 3000
			currentRoom.StartTime = startTime
			currentRoom.Unlock()

			broadcast(Message{
				Type:      "race_starting",
				StartTime: startTime,
			}, nil)

		case "progress":
			if currentRoom == nil {
				continue
			}
			currentRoom.Lock()
			player := currentRoom.Players[conn]
			if player == nil {
				currentRoom.Unlock()
				continue
			}
			player.Progress = msg.Progress
			player.WPM = msg.WPM
			player.Accuracy = msg.Accuracy
			player.CharsTyped = msg.CharsTyped

			if msg.Finished && !player.Finished {
				player.Finished = true
				player.FinishTime = time.Now().UnixMilli()
				currentRoom.Finishers = append(currentRoom.Finishers, playerName)
				place := len(currentRoom.Finishers)
				player.Place = place
				currentRoom.Unlock()

				broadcast(Message{
					Type:       "player_finished",
					PlayerName: playerName,
					Place:      place,
					WPM:        msg.WPM,
				}, nil)

				currentRoom.Lock()
				if len(currentRoom.Finishers) == len(currentRoom.Players) {
					currentRoom.State = "finished"
					currentRoom.Unlock()
					broadcast(Message{Type: "race_finished"}, nil)
				} else {
					currentRoom.Unlock()
				}
			} else {
				currentRoom.Unlock()
			}

			broadcast(Message{
				Type:    "players_update",
				Players: getPlayers(currentRoom),
			}, conn)

		case "list_rooms":
			roomsMu.Lock()
			var list []RoomSum
			for id, r := range rooms {
				if r.State == "waiting" {
					r.Lock()
					list = append(list, RoomSum{
						ID:      id,
						Players: len(r.Players),
						Passage: r.Passage,
					})
					r.Unlock()
				}
			}
			roomsMu.Unlock()
			sendJSON(Message{Type: "rooms_list", Rooms: list})
		}
	}

	// Cleanup on disconnect
	if currentRoom != nil {
		currentRoom.Lock()
		delete(currentRoom.Players, conn)
		left := len(currentRoom.Players)
		currentRoom.Unlock()

		if left == 0 {
			roomsMu.Lock()
			delete(rooms, currentRoomID)
			roomsMu.Unlock()
		} else {
			broadcast(Message{
				Type:    "player_left",
				Players: getPlayers(currentRoom),
			}, nil)
		}
	}
}

// ─── STATIC FILE SERVER ────────────────────────────────────────────────────

func staticHandler(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path
	if filePath == "/" {
		filePath = "/index.html"
	}
	// Try embedded frontend/dist first (production)
	subFS, err := fs.Sub(embeddedFrontend, "frontend/dist")
	if err == nil {
		data, err := fs.ReadFile(subFS, path.Clean(filePath))
		if err == nil {
			ext := path.Ext(filePath)
			mime := map[string]string{
				".html": "text/html",
				".css":  "text/css",
				".js":   "application/javascript",
				".json": "application/json",
				".svg":  "image/svg+xml",
				".png":  "image/png",
				".ico":  "image/x-icon",
			}
			if ct, ok := mime[ext]; ok {
				w.Header().Set("Content-Type", ct)
			}
			w.Write(data)
			return
		}
	}
	// Fallback: try frontend/dist from disk (dev)
	diskPath := "frontend/dist" + filePath
	if _, err := os.Stat(diskPath); err == nil {
		http.ServeFile(w, r, diskPath)
		return
	}
	// Last fallback: frontend/ from disk (dev without build)
	devPath := "frontend" + filePath
	if _, err := os.Stat(devPath); err == nil {
		http.ServeFile(w, r, devPath)
		return
	}
	http.NotFound(w, r)
}

// ─── MAIN ──────────────────────────────────────────────────────────────────

func main() {
	rand.Seed(time.Now().UnixNano())

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.HandleFunc("/ws", handleWS)
	http.HandleFunc("/", staticHandler)

	fmt.Printf("🏎️  TypeRacer (Go+Vue) running on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
