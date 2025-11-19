package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB

const bucketName = "todos"

type Todo struct {
	ID   uint64 `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

func main() {
	storage := os.Getenv("Storage")
	if storage == "" {
		storage = "todos.db"
	}
	listenPort := os.Getenv("listenPort")
	if listenPort == "" {
		listenPort = "8080"
	}

	// ensure directory exists
	if dir := path.Dir(storage); dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}

	var err error
	db, err = bolt.Open(storage, 0o600, nil)
	if err != nil {
		log.Fatalf("{\"level\":\"error\", \"message\":\"open db: %v\"}", err)
	}
	defer db.Close()

	// create bucket if not exists
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	if err != nil {
		log.Fatalf("{\"level\":\"error\", \"message\":\"create bucket: %v\"}", err)
	}

	http.HandleFunc("/healthz", healthHandler)
	http.HandleFunc("/todos", todosHandler)
	http.HandleFunc("/todos/", todoHandler)

	addr := ":" + listenPort
	log.Printf("{\"level\":\"info\", \"listening on\":\"%s\", \"storage\":\"%s\"}", addr, storage)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	err := db.View(func(tx *bolt.Tx) error {
		_ = tx.Bucket([]byte(bucketName))
		return nil
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "{\"level\":\"error\", \"message\":\"DB not accessible\"}")
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func idToBytes(id uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, id)
	return b
}

// func bytesToUint64(b []byte) uint64 {
//	return binary.BigEndian.Uint64(b)
// }

func todosHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listTodos(w)
	case http.MethodPost:
		createTodo(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func todoHandler(w http.ResponseWriter, r *http.Request) {
	idStr := path.Base(r.URL.Path)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("{\"level\":\"error\", \"message\":\"Todo not found %d\"}", id)
		fmt.Fprint(w, "{\"level\":\"error\", \"message\":\"Todo not found\"}")
		return
	}
	switch r.Method {
	case http.MethodGet:
		getTodo(w, id)
	case http.MethodPut:
		updateTodo(w, r, id)
	case http.MethodDelete:
		deleteTodo(w, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func listTodos(w http.ResponseWriter) {
	var out []Todo
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.ForEach(func(k, v []byte) error {
			if len(k) != 8 {
				return nil
			}
			var t Todo
			if err := json.Unmarshal(v, &t); err != nil {
				return err
			}
			out = append(out, t)
			return nil
		})
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("{\"level\":\"error\", \"message\":\"invalid body\"}")
		fmt.Fprint(w, "{\"level\":\"error\", \"message\":\"invalid body\"}")
		return
	}
	var created Todo
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		id, _ := b.NextSequence()
		created = Todo{ID: id, Text: in.Text, Done: false}
		buf, err := json.Marshal(created)
		if err != nil {
			return err
		}
		return b.Put(idToBytes(id), buf)
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	log.Printf("{\"level\":\"info\", \"Todo created\":\"%+v\"}", created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

func getTodo(w http.ResponseWriter, id uint64) {
	var t Todo
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		v := b.Get(idToBytes(id))
		if v == nil {
			log.Printf("{\"level\":\"error\", \"message\":\"Todo not found %d\"}", id)
			return fmt.Errorf("{\"level\":\"error\", \"message\":\"TTodo not found %d\"}", id)
		}
		return json.Unmarshal(v, &t)
	})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Printf("{\"level\":\"error\", \"message\":\"Todo not found %d\"}", id)
		fmt.Printf("{\"level\":\"error\", \"message\":\"Todo not found %d\"}", id)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}

func updateTodo(w http.ResponseWriter, r *http.Request, id uint64) {
	var in struct {
		Text *string `json:"text"`
		Done *bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("{\"level\":\"error\", \"message\":\"invalid body\"}")
		fmt.Fprint(w, "{\"level\":\"error\", \"message\":\"invalid body\"}")
		return
	}
	var updated Todo
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		v := b.Get(idToBytes(id))
		if v == nil {
			log.Printf("{\"level\":\"error\", \"message\":\"Todo not found %d\"}", id)
			return fmt.Errorf("{\"level\":\"error\", \"message\":\"Todo not found %d\"}", id)
		}
		if err := json.Unmarshal(v, &updated); err != nil {
			return err
		}
		if in.Text != nil {
			updated.Text = *in.Text
		}
		if in.Done != nil {
			updated.Done = *in.Done
		}
		buf, err := json.Marshal(updated)
		if err != nil {
			return err
		}
		return b.Put(idToBytes(id), buf)
	})

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, err.Error())
		return
	}

	log.Printf("{\"level\":\"info\", \"Todo updated\":\"%+v\"}", updated)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updated)
}

func deleteTodo(w http.ResponseWriter, id uint64) {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		v := b.Get(idToBytes(id))
		if v == nil {
			log.Printf("{\"level\":\"error\", \"message\":\"Todo not found %d\"}", id)
			return fmt.Errorf("{\"level\":\"error\", \"message\":\"Todo not found %d\"}", id)
		}
		return b.Delete(idToBytes(id))
	})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, err.Error())
		return
	}
	log.Printf("{\"level\":\"info\", \"Todo deleted\":\"%d\"}", id)
	w.WriteHeader(http.StatusNoContent)
}
