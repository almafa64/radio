package appstate

import (
	"radio_site/libs/myconfig"

	"encoding/json"
	"errors"
	"os"
	"sync"

	set "github.com/deckarep/golang-set"
	"github.com/samber/lo"
)

var WriteError = errors.New("state isn't locked for writing")

// TODO: make wrap for push buttons to store client (remove from marshalling)

type PinStates map[int]bool

type State struct {
	PinStates PinStates
}

// TODO: convert old code to new system instead of this ductape
func (c *PinStates) ToByteSlice() []byte {
	status := make([]byte, 64)

	for k, v := range *c {
		var c byte = '0'
		if v {
			c = '1'
		}

		status[k] = c
	}

	return status
}

func (c *PinStates) SetPinStatusTo(pin int, enabled bool) error {
	if !locked_for_write {
		return WriteError
	}

	(*c)[pin] = enabled
	return nil
}

func (c *PinStates) TogglePinStatus(pin int) error {
	return c.SetPinStatusTo(pin, !(*c)[pin])
}

var file_lock = new(sync.Mutex)
var (
	state_lock              = new(sync.RWMutex)
	locked_for_write        = false
	global_state     *State = new(State)
)

func Get() *State {
	state_lock.RLock()
	return global_state
}

func GetWritable() *State {
	state_lock.Lock()
	locked_for_write = true
	return global_state
}

func (*State) Release() error {
	if locked_for_write {
		state_lock.Unlock()
		locked_for_write = false

		// TODO: maybe this will be too slow (debounce?)
		return global_state.Save()
	} else {
		state_lock.RUnlock()
	}

	return nil
}

func (*State) Save() error {
	file_lock.Lock()
	defer file_lock.Unlock()

	state_lock.RLock()
	data, err := json.Marshal(global_state)
	state_lock.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(myconfig.Get().StateFilePath, data, os.FileMode(0o644))
}

// TODO: dont let 2+ button have same pin
func Load() error {
	file_lock.Lock()
	defer file_lock.Unlock()

	path := myconfig.Get().StateFilePath

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if len(data) < 2 {
		data = []byte{'{', '}'}
	}

	state := new(State)

	if err := json.Unmarshal(data, state); err != nil {
		return err
	}

	// set push buttons state to their default (cannot be pushed at boot)
	for _, button := range myconfig.Get().GetAllButton() {
		if !button.IsToggle {
			state.PinStates[int(button.Pin)] = button.Default == 1
		}
	}

	state_lock.Lock()
	global_state = state
	state_lock.Unlock()

	return nil
}

func LoadOrSaveDefault() error {
	err := Load()

	if err == nil {
		state := GetWritable()

		config_list := lo.Map(myconfig.Get().GetAllButton(), func(v myconfig.Button, _ int) int {
			return int(v.Pin)
		})
		loaded_list := lo.Keys(state.PinStates)

		config_set := set.NewSetFromSlice(lo.ToAnySlice(config_list))
		loaded_set := set.NewSetFromSlice(lo.ToAnySlice(loaded_list))

		to_add_pins := config_set.Difference(loaded_set)
		to_delete_pins := loaded_set.Difference(config_set)

		for pin := range to_add_pins.Iter() {
			v, _ := pin.(int)
			state.PinStates[v] = myconfig.Get().GetButtonByPin(v).Default == 1
		}

		for pin := range to_delete_pins.Iter() {
			v, _ := pin.(int)
			delete(state.PinStates, v)
		}

		return state.Release()
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	state := GetWritable()
	state.PinStates = map[int]bool{}

	for _, button := range myconfig.Get().GetAllButton() {
		button_state := false

		if button.Default == 1 {
			button_state = true
		}

		state.PinStates[int(button.Pin)] = button_state
	}

	return state.Release()
}
