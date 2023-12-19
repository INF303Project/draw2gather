package game

type state interface {
	Enter(*Game)
	Exit(*Game)
	Handle(*Game, *Message) state
}
