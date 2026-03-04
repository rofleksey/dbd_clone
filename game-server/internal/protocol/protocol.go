package protocol

// Client -> Server messages

type ClientMessage struct {
	Type   string  `json:"type"`
	Action string  `json:"action,omitempty"`
	Target string  `json:"target,omitempty"`
	PosX   float64 `json:"pos_x,omitempty"`
	PosY   float64 `json:"pos_y,omitempty"`
	PosZ   float64 `json:"pos_z,omitempty"`
	RotY   float64 `json:"rot_y,omitempty"`
	State  string  `json:"state,omitempty"` // idle, walking, running, crouching
	Token  string  `json:"token,omitempty"`
	UserID int     `json:"user_id,omitempty"`
}

// Server -> Client messages

type ServerMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

type GameState struct {
	Tick          int64         `json:"tick"`
	Players       []PlayerState `json:"players"`
	Generators    []GenState    `json:"generators"`
	Pallets       []PalletState `json:"pallets"`
	Traps         []TrapState   `json:"traps"`
	Gates         []GateState   `json:"gates"`
	Hooks         []HookState   `json:"hooks"`
	Windows       []WindowState `json:"windows"`
	ScratchMarks  []ScratchMark `json:"scratch_marks,omitempty"`
	BloodTrails   []BloodTrail  `json:"blood_trails,omitempty"`
	TimeRemaining float64       `json:"time_remaining"`
	GensCompleted int           `json:"gens_completed"`
	GensRequired  int           `json:"gens_required"`
	GatesPowered  bool          `json:"gates_powered"`
}

type PlayerState struct {
	UserID         int     `json:"user_id"`
	Username       string  `json:"username"`
	Role           string  `json:"role"` // killer, survivor
	PosX           float64 `json:"pos_x"`
	PosY           float64 `json:"pos_y"`
	PosZ           float64 `json:"pos_z"`
	RotY           float64 `json:"rot_y"`
	Health         int     `json:"health"`                    // 0=dead/sacrificed, 1=dying, 2=injured, 3=healthy
	MoveState      string  `json:"move_state"`                // idle, walking, running, crouching
	ActionState    string  `json:"action_state"`              // none, repairing, healing, unhooking, opening_gate, carrying, being_carried, hooked, dying, trapped, stunned, attacking
	ActionTarget   string  `json:"action_target,omitempty"`
	ActionProgress float64 `json:"action_progress,omitempty"` // 0-1
	CarryingID     int     `json:"carrying_id,omitempty"`
	HookedOn       string  `json:"hooked_on,omitempty"`
	HookStage      int     `json:"hook_stage,omitempty"` // 1, 2, 3
	TrappedIn      string  `json:"trapped_in,omitempty"`
	Ping           int     `json:"ping"`
	IsAlive        bool    `json:"is_alive"`
	HasEscaped     bool    `json:"has_escaped"`
	TrapCount      int     `json:"trap_count,omitempty"` // for killer
}

type GenState struct {
	ID            string  `json:"id"`
	PosX          float64 `json:"pos_x"`
	PosY          float64 `json:"pos_y"`
	PosZ          float64 `json:"pos_z"`
	Progress      float64 `json:"progress"` // 0-1
	Done          bool    `json:"done"`
	Regressing    bool    `json:"regressing"`
	BeingRepaired bool    `json:"being_repaired"`
}

type PalletState struct {
	ID      string  `json:"id"`
	PosX    float64 `json:"pos_x"`
	PosY    float64 `json:"pos_y"`
	PosZ    float64 `json:"pos_z"`
	RotY    float64 `json:"rot_y"`
	Dropped bool    `json:"dropped"`
	Broken  bool    `json:"broken"`
}

type TrapState struct {
	ID        string  `json:"id"`
	PosX      float64 `json:"pos_x"`
	PosY      float64 `json:"pos_y"`
	PosZ      float64 `json:"pos_z"`
	Placed    bool    `json:"placed"`
	Triggered bool    `json:"triggered"`
	// Only visible to killer or when nearby
	Visible bool `json:"visible"`
}

type GateState struct {
	ID       string  `json:"id"`
	PosX     float64 `json:"pos_x"`
	PosY     float64 `json:"pos_y"`
	PosZ     float64 `json:"pos_z"`
	RotY     float64 `json:"rot_y"`
	Progress float64 `json:"progress"` // 0-1
	Open     bool    `json:"open"`
	Powered  bool    `json:"powered"`
}

type HookState struct {
	ID       string  `json:"id"`
	PosX     float64 `json:"pos_x"`
	PosY     float64 `json:"pos_y"`
	PosZ     float64 `json:"pos_z"`
	Occupied bool    `json:"occupied"`
	PlayerID int     `json:"player_id,omitempty"`
}

type WindowState struct {
	ID   string  `json:"id"`
	PosX float64 `json:"pos_x"`
	PosY float64 `json:"pos_y"`
	PosZ float64 `json:"pos_z"`
	RotY float64 `json:"rot_y"`
}

type ScratchMark struct {
	PosX float64 `json:"pos_x"`
	PosZ float64 `json:"pos_z"`
	Age  float64 `json:"age"` // seconds
}

type BloodTrail struct {
	PosX float64 `json:"pos_x"`
	PosZ float64 `json:"pos_z"`
	Age  float64 `json:"age"`
}

type GameEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data,omitempty"`
}

// Game report to master server
type GameReport struct {
	GameID  int            `json:"game_id"`
	Result  string         `json:"result"`
	Players []PlayerReport `json:"players"`
}

type PlayerReport struct {
	UserID   int    `json:"user_id"`
	Role     string `json:"role"`
	Survived bool   `json:"survived"`
	Kills    int    `json:"kills"`
	GensDone int    `json:"gens_done"`
}

type GameProgress struct {
	GameID         int      `json:"game_id"`
	GensCompleted  int      `json:"gens_completed"`
	SurvivorsAlive int      `json:"survivors_alive"`
	ElapsedSeconds float64  `json:"elapsed_seconds"`
	PlayerNames    []string `json:"player_names"`
	KillerName     string   `json:"killer_name"`
}
