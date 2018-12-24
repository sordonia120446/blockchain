package blockchain

import (
    "crypto/sha256"
    "encoding/hex"
    "time"
)

type Block struct {
    Index     int
    Timestamp string
    BPM       int
    Hash      string
    PrevHash  string
}

func calculateHash(block Block) string {
    record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
    h := sha256.New()
    h.Write([]byte(record))
    hashed := h.Sum(nil)
    return hex.EncodeToString(hashed)
}

func GenerateBlock(prevBlock Block, BPM int) Block {
    var newBlock Block

    newBlock.Index = prevBlock.Index + 1
    newBlock.Timestamp = time.Now().String()
    newBlock.BPM = BPM
    newBlock.PrevHash = prevBlock.Hash
    newBlock.Hash = calculateHash(newBlock)
    return newBlock
}

func IsValidBlock(newBlock, prevBlock Block) bool {
    if prevBlock.Index + 1 != newBlock.Index {
        return false
    }

    if prevBlock.Hash != newBlock.PrevHash {
        return false
    }

    if calculateHash(newBlock) != newBlock.Hash {
        return false
    }

    return true
}
