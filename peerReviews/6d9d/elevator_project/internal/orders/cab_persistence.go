package orders

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func (om *OrderManager) cabCallsFilename() string {
	return filepath.Join(om.storageDir, fmt.Sprintf("cab_calls_%d.json", om.selfState.ID))
}

func (om *OrderManager) cabCallsTempFilename() string {
	return filepath.Join(om.storageDir, fmt.Sprintf("cab_calls_%d.tmp", om.selfState.ID))
}

// Cab calls are local and persisted on this node (not shared across the network).
// saveCabCalls persists the local cab-call table to disk.
// It uses a temp file + rename to ensure atomic writes.
func (om *OrderManager) saveCabCalls() error {
	tempFileName := om.cabCallsTempFilename()
	finalFileName := om.cabCallsFilename()

	file, err := os.Create(tempFileName)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)

	if err := encoder.Encode(om.cabCalls); err != nil {
		file.Close()
		_ = os.Remove(tempFileName)
		return err
	}

	if err := file.Sync(); err != nil {
		file.Close()
		_ = os.Remove(tempFileName)
		return err
	}

	if err := file.Close(); err != nil {
		_ = os.Remove(tempFileName)
		return err
	}

	if err := os.Rename(tempFileName, finalFileName); err != nil {
		_ = os.Remove(tempFileName)
		return err
	}

	directory, err := os.Open(filepath.Dir(finalFileName))
	if err != nil {
		return err
	}
	defer directory.Close()

	if err := directory.Sync(); err != nil {
		return err
	}

	return nil
}

// loadCabCalls restores cab-call state from disk if it exists.
func (om *OrderManager) loadCabCalls() error {
	file, err := os.Open(om.cabCallsFilename())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	var stored []bool

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&stored); err != nil {
		return err
	}

	if len(stored) != len(om.cabCalls) {
		return fmt.Errorf("invalid cab call file: expected %d floors, got %d", len(om.cabCalls), len(stored))
	}

	copy(om.cabCalls, stored)
	return nil
}
