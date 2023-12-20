package game

type state interface {
	Enter(*Game)
	Exit(*Game)
	HandleMessage(*Game, *Message) state
}

type waiting struct {
}

func (w *waiting) Enter(_ *Game) {
	panic("not implemented") // TODO: Implement
}

func (w *waiting) Exit(_ *Game) {
	panic("not implemented") // TODO: Implement
}

func (w *waiting) HandleMessage(_ *Game, _ *Message) state {
	panic("not implemented") // TODO: Implement
}

type picking struct {
}

func (p *picking) Enter(_ *Game) {
	panic("not implemented") // TODO: Implement
}

func (p *picking) Exit(_ *Game) {
	panic("not implemented") // TODO: Implement
}

func (p *picking) HandleMessage(_ *Game, _ *Message) state {
	panic("not implemented") // TODO: Implement
}

type drawing struct {
}

func (d *drawing) Enter(_ *Game) {
	panic("not implemented") // TODO: Implement
}

func (d *drawing) Exit(_ *Game) {
	panic("not implemented") // TODO: Implement
}

func (d *drawing) HandleMessage(g *Game, m *Message) state {
	switch m.Action {
	case draw, erase, lineDraw, rectDraw, rectFill, circleDraw, circleFill,
		changeColor, changePencilSize, changeEraserSize, clearBoard:
		g.handleBoardAction(m)
	case guess:
		g.handleGuess(m)
	case chat:
		g.handleChat(m)
	case quit:
		g.handleQuit(m)
	case kick:
		g.handleKick(m)
	}
	return nil
}

type ending struct {
}

func (e *ending) Enter(_ *Game) {
	panic("not implemented") // TODO: Implement
}

func (e *ending) Exit(_ *Game) {
	panic("not implemented") // TODO: Implement
}

func (e *ending) HandleMessage(_ *Game, _ *Message) state {
	panic("not implemented") // TODO: Implement
}
