package main

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "io"
    "log"
    "net/http"
    "os"
    "sync"
    "time"

    "github.com/davecgh/go-spew/spew"
    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
)

type Block struct {
    Index     int
    Timestamp string
    BPM       int
    Hash      string
    PrevHash  string
}

// Slice of Block
var Blockchain []Block  // TODO not use global var?

func calculateHash(block Block) string {
    record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
    h := sha256.New()
    h.Write([]byte(record))
    hashed := h.Sum(nil)
    return hex.EncodeToString(hashed)
}

func generateBlock(prevBlock Block, BPM int) Block {
    var newBlock Block

    newBlock.Index = prevBlock.Index + 1
    newBlock.Timestamp = time.Now().String()
    newBlock.BPM = BPM
    newBlock.PrevHash = prevBlock.Hash
    newBlock.Hash = calculateHash(newBlock)
    return newBlock
}

func isValidBlock(newBlock, prevBlock Block) bool {
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

// TODO move this web server stuff elsewhere
var mutex = &sync.Mutex{}

func run() error {
    mux := makeMuxRouter()
    httpAddr := os.Getenv("ADDR")
    log.Println("Listening on ", os.Getenv("ADDR"))
    s := &http.Server{
        Addr: ":" + httpAddr,
        Handler: mux,
        ReadTimeout: 5 * time.Second,
        WriteTimeout: 5 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }

    if err := s.ListenAndServe(); err != nil {
        return err
    }

    return nil
}

func makeMuxRouter() http.Handler {
    muxRouter := mux.NewRouter()
    muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
    muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
    return muxRouter
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
    bytes, err := json.MarshalIndent(Blockchain, "", "  ")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    io.WriteString(w, string(bytes))
}

type Message struct {
    BPM int
}

func handleWriteBlock(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type", "application/json")
    var msg Message

    decoder := json.NewDecoder(req.Body)
    if err := decoder.Decode(&msg); err != nil {
        respondWithJSON(res, req, http.StatusBadRequest, req.Body)
        return
    }
    defer req.Body.Close()

    prevBlock := Blockchain[len(Blockchain)-1]
    newBlock := generateBlock(prevBlock, msg.BPM)

    if !isValidBlock(newBlock, Blockchain[len(Blockchain)-1]) {
        respondWithJSON(res, req, http.StatusBadRequest, req.Body)
        return
    }
    mutex.Lock()
    Blockchain = append(Blockchain, newBlock)
    spew.Dump(Blockchain)
    mutex.Unlock()

    respondWithJSON(res, req, http.StatusCreated, newBlock)

}

// Handle errors "gracefully" is a thing in Go?
func respondWithJSON(res http.ResponseWriter, req *http.Request, code int, payload interface{}) {
    response, err := json.MarshalIndent(payload, "", "  ")
    if err != nil {
        res.WriteHeader(http.StatusInternalServerError)
        res.Write([]byte("HTTP 500: Internal Server Error"))
        return
    }
    res.WriteHeader(code)
    res.Write(response)
}

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal(err)
    }

    // We isolate the genesis block into its own go routine so we can have a
    // separation of concerns from our blockchain logic and our web server logic.
    // This will work without the go routine but it’s just cleaner this way.
    go func() {
        t := time.Now()
        genesisBlock := Block{0, t.String(), 0, "", ""}
        spew.Dump(genesisBlock)

        mutex.Lock()
        Blockchain = append(Blockchain, genesisBlock)
        mutex.Unlock()
    }()
    log.Fatal(run())
}
