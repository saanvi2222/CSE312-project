package main

type IncomingMessage struct { //anything a browser sends us
	MessageType string  `json:"messageType"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	StartX      float64 `json:"startX"`
	StartY      float64 `json:"startY"`
	EndX        float64 `json:"endX"`
	EndY        float64 `json:"endY"`
	Color       string  `json:"color"`
	Radius      float64 `json:"radius"`
	Text        string  `json:"text"`
	Size        float64 `json:"size"`
}

type WelcomeMessage struct { //tells a new client what name the server gave it
	MessageType string `json:"messageType"`
	Username    string `json:"username"`
}

type CursorMessage struct { //one user's mouse position, sent to everyone else
	MessageType string  `json:"messageType"`
	Username    string  `json:"username"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
}
type DrawingMessage struct {
	MessageType string  `json:"messageType"`
	Username    string  `json:"username"`
	StartX      float64 `json:"startX"`
	StartY      float64 `json:"startY"`
	EndX        float64 `json:"endX"`
	EndY        float64 `json:"endY"`
	Color       string  `json:"color"`
}

type InitStrokesMessage struct { //sends the whole board to someone who just joined
	MessageType string           `json:"messageType"`
	Strokes     []DrawingMessage `json:"strokes"`
	Texts       []TextMessage    `json:"texts"`
}

type ActiveUsersMessage struct { //list of people currently on the board
	MessageType string   `json:"messageType"`
	Users       []string `json:"users"`
}

type UserLeftMessage struct { //tells everyone to remove a user's cursor.
	MessageType string `json:"messageType"`
	Username    string `json:"username"`
}

type EraseMessage struct { //removes every segment within Radius of (X, Y); anyone can erase anyone's strokes
	MessageType string  `json:"messageType"`
	Username    string  `json:"username"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Radius      float64 `json:"radius"`
}

type TextMessage struct { //a piece of text placed on the board; (X, Y) is its top-left corner
	MessageType string  `json:"messageType"`
	Username    string  `json:"username"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Text        string  `json:"text"`
	Color       string  `json:"color"`
	Size        float64 `json:"size"`
}
