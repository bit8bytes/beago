package stores

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/bit8bytes/beago/llms"
)

// Stamp sets Timestamp and Hash on msg, chaining from prevHash.
// Hash = SHA256(prevHash + role + timestamp + content).
// Call this inside Add() before persisting, passing the hash of the last
// stored message as prevHash (empty string for the first message).
func Stamp(msg *llms.Message, prevHash string) {
	msg.Timestamp = time.Now().UTC()

	h := sha256.New()
	h.Write([]byte(prevHash))
	h.Write([]byte(msg.Role))
	h.Write([]byte(msg.Timestamp.Format(time.RFC3339Nano)))
	h.Write([]byte(msg.Content))
	msg.Hash = hex.EncodeToString(h.Sum(nil))
}

// Verify replays the hash chain over msgs and returns the index of the first
// message whose hash does not match, or -1 if the chain is intact.
// Pass the hash that preceded the first message as prevHash
// (empty string if msgs starts from the beginning of the store).
func Verify(msgs []llms.Message, prevHash string) int {
	for i, msg := range msgs {
		h := sha256.New()
		h.Write([]byte(prevHash))
		h.Write([]byte(msg.Role))
		h.Write([]byte(msg.Timestamp.Format(time.RFC3339Nano)))
		h.Write([]byte(msg.Content))
		want := hex.EncodeToString(h.Sum(nil))
		if msg.Hash != want {
			return i
		}
		prevHash = msg.Hash
	}
	return -1
}
