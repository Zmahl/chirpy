package db

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mu   *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Password     []byte `json:"password"`
	RefreshToken string `json:"refresh_token"`
	IsRed        bool   `json:"is_chirpy_red"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mu:   &sync.RWMutex{},
	}

	err := db.ensureDB()
	return db, err
}

func (db *DB) CreateChirp(authorId int, body string) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, nil
	}

	id := len(dbStructure.Chirps) + 1

	chirp := Chirp{
		ID:       id,
		Body:     body,
		AuthorID: authorId,
	}
	dbStructure.Chirps[id] = chirp

	err = db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *DB) DeleteChirp(chirpId int) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	delete(dbStructure.Chirps, chirpId)
	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

func (db *DB) GetUser(email string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	existingUser := User{}
	for _, user := range dbStructure.Users {
		if user.Email == email {
			existingUser.Email = user.Email
			existingUser.ID = user.ID
			existingUser.Password = user.Password
			existingUser.RefreshToken = user.RefreshToken
			existingUser.IsRed = user.IsRed
		}
	}

	return existingUser, nil
}

func (db *DB) GetRefreshToken(refreshToken string) (int, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return 0, err
	}
	for i := range dbStructure.Users {
		if dbStructure.Users[i].RefreshToken == refreshToken {
			return dbStructure.Users[i].ID, nil
		}
	}

	return 0, errors.New("could not find refresh token")
}

func (db *DB) RefreshToken(id int, refreshToken string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	for i := range dbStructure.Users {
		if dbStructure.Users[i].ID == id {
			dbStructure.Users[i] = User{
				ID:           dbStructure.Users[i].ID,
				Email:        dbStructure.Users[i].Email,
				Password:     dbStructure.Users[i].Password,
				RefreshToken: refreshToken,
				IsRed:        dbStructure.Users[i].IsRed,
			}
		}
	}

	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) RevokeToken(refreshToken string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	for i := range dbStructure.Users {
		if dbStructure.Users[i].RefreshToken == refreshToken {
			dbStructure.Users[i] = User{
				ID:           dbStructure.Users[i].ID,
				Password:     dbStructure.Users[i].Password,
				RefreshToken: "",
				IsRed:        dbStructure.Users[i].IsRed,
			}
		}
	}
	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) CreateUser(email string, password string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	id := len(dbStructure.Users) + 1
	user := User{
		ID:       id,
		Email:    email,
		Password: hashedPassword,
		IsRed:    false,
	}
	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (db *DB) UpdateUser(id int, email string, hashedPassword string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	for i := range dbStructure.Users {
		if dbStructure.Users[i].ID == id {
			dbStructure.Users[i] = User{
				Email:        email,
				Password:     []byte(hashedPassword),
				ID:           id,
				RefreshToken: dbStructure.Users[i].RefreshToken,
			}
		}
	}

	db.writeDB(dbStructure)
	return nil
}

func (db *DB) UpgradeUser(userId int) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	for i := range dbStructure.Users {
		if dbStructure.Users[i].ID == userId {
			dbStructure.Users[i] = User{
				ID:           dbStructure.Users[i].ID,
				Password:     dbStructure.Users[i].Password,
				Email:        dbStructure.Users[i].Email,
				RefreshToken: dbStructure.Users[i].RefreshToken,
				IsRed:        true,
			}
			err = db.writeDB(dbStructure)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return errors.New("could not find user")
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return db.createDB()
	}

	return err
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	dbStructure := DBStructure{}
	data, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return dbStructure, err
	}

	err = json.Unmarshal(data, &dbStructure)
	if err != nil {
		return dbStructure, err
	}

	return dbStructure, nil
}

func (db *DB) createDB() error {
	dbStructure := DBStructure{
		Chirps: map[int]Chirp{},
		Users:  map[int]User{},
	}
	return db.writeDB(dbStructure)
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, data, 0600)
	if err != nil {
		return err
	}
	return nil
}
