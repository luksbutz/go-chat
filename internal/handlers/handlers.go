package handlers

import (
	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sort"
)

var wsChan = make(chan WsJSONPayload)

var clients = make(map[WsConnection]string)

// views is the jet view set
var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

// upgradeConnection is the websocket upgrade from gorilla/websockets
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Home renders the home page
func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}
}

// WsConnection is a wrapper for our websocket connection, in case
// we ever need to put more data into the struct
type WsConnection struct {
	*websocket.Conn
}

// WsJSONResponse defines the response sent back from websocket
type WsJSONResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

// WsJSONPayload defines the websocket request from the client
type WsJSONPayload struct {
	Action   string       `json:"action"`
	Username string       `json:"username"`
	Message  string       `json:"message"`
	Conn     WsConnection `json:"-"`
}

// WsEndpoint upgrade connection to websocket
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Client connected to endpoint", r.RemoteAddr)

	response := WsJSONResponse{
		Message: `<em><small>Connected to server</small></em>`,
	}

	conn := WsConnection{Conn: ws}
	clients[conn] = ""

	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	go ListenForWs(&conn)
}

// ListenForWs is a goroutine that handles communication between server and client, and
// feeds data into the wsChan
func ListenForWs(conn *WsConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Error %v\n", r)
		}
	}()

	var payload WsJSONPayload

	for {
		if err := conn.ReadJSON(&payload); err == nil {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

// ListenToWsChan is a goroutine that waits for an entry on the wsChan, and handles it according to the
// specified action
func ListenToWsChan() {
	var response WsJSONResponse

	for {
		e := <-wsChan

		switch e.Action {
		case "username":
			// get a list of all users and send it back via broadcast
			clients[e.Conn] = e.Username
			response.Action = "list_users"
			response.ConnectedUsers = getUserList()
			broadcastToAll(response)
		case "left":
			response.Action = "list_users"
			delete(clients, e.Conn)
			response.ConnectedUsers = getUserList()
			broadcastToAll(response)
		}

		//response.Action = "Got here"
		//response.Message = fmt.Sprintf("Some message, and action was %s", e.Action)
	}
}

// getUserList returns a list with all the connected users
func getUserList() []string {
	var userList []string

	for _, c := range clients {
		userList = append(userList, c)
	}

	sort.Strings(userList)

	return userList
}

// broadcastToAll sends a ws response to all connected clients
func broadcastToAll(response WsJSONResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("Websocket err", err)
			_ = client.Close()
			delete(clients, client)
		}
	}
}

// renderPage renders a jet template
func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		return err
	}

	return view.Execute(w, data, nil)
}
