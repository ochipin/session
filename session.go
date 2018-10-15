package session

import (
	"container/list"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// セッションIDを生成する際に使用されるランダムオブジェクト
var random = rand.New(rand.NewSource(time.Now().UnixNano()))

// Interface : セッション情報を操作するインターフェース
type Interface interface {
	SessionID() string
	Get(string) (interface{}, error)
	Int(string) (int, error)
	Str(string) (string, error)
	Bool(string) (bool, error)
	Struct(string, interface{}) error
	Set(string, interface{})
	Delete(string)
}

// Storage : データ管理構造体
type Storage struct {
	Access int64
	ssid   string
	Values map[string]interface{}
}

// SessionID : セッションIDを返却する
func (storage *Storage) SessionID() string {
	return storage.ssid
}

// Get : セッションに登録されたデータを取り出す
func (storage *Storage) Get(key string) (interface{}, error) {
	v, ok := storage.Values[key]
	if !ok {
		return nil, fmt.Errorf("'%s' - not found", key)
	}
	return v, nil
}

// Int : セッションに登録されたデータを整数型として取り出す
func (storage *Storage) Int(key string) (int, error) {
	// 指定されたキー名から値を取得
	v, err := storage.Get(key)
	if err != nil {
		return 0, err
	}
	// int へキャスト
	n, ok := v.(int)
	if !ok {
		// キャスト失敗の場合、文字列へ変換してから数字変換できるか試す
		n, err := strconv.Atoi(fmt.Sprint(v))
		if err != nil {
			return 0, fmt.Errorf("'%s' - not number", key)
		}
		return n, nil
	}
	return n, nil
}

// Str : セッションに登録されたデータを文字列として取り出す
func (storage *Storage) Str(key string) (string, error) {
	// 指定されたキー名から値を取得
	v, err := storage.Get(key)
	if err != nil {
		return "", err
	}
	return fmt.Sprint(v), nil
}

// Bool : セッションに登録されたデータをBool値として取り出す
func (storage *Storage) Bool(key string) (bool, error) {
	// 指定されたキー名から値を取得
	v, err := storage.Get(key)
	if err != nil {
		return false, err
	}
	if b, ok := v.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("'%s' - not boolean", key)
}

// Struct : セッションに登録された構造体の値を取り出す
func (storage *Storage) Struct(key string, i interface{}) error {
	// 指定されたキー名から値を取得
	v, err := storage.Get(key)
	if err != nil {
		return err
	}
	// 取得した値をJSONへ変換する
	buf, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("'%s' - not struct", key)
	}
	if err := json.Unmarshal(buf, i); err != nil {
		return fmt.Errorf("'%s' - not struct", key)
	}
	return nil
}

// Set : セッションにデータを登録する
func (storage *Storage) Set(key string, value interface{}) {
	storage.Values[key] = value
}

// Delete : セッションに登録されたデータを削除する
func (storage *Storage) Delete(key string) {
	if _, ok := storage.Values[key]; ok {
		delete(storage.Values, key)
	}
}

// Session : セッション管理構造体
type Session struct {
	name     string
	maxage   int
	keeptime int
	path     string
	secure   bool
	mu       sync.Mutex
	list     *list.List
	data     map[string]*list.Element
}

// New : セッション管理構造体の初期化を行う
func New(name string, maxage, keep int, secure bool) *Session {
	if name == "" {
		name = "noname"
	}
	if keep <= 0 {
		keep = 60
	}
	session := &Session{
		name:     name,
		maxage:   maxage,
		keeptime: keep,
		path:     "/",
		secure:   secure,
		list:     list.New(),
		data:     make(map[string]*list.Element),
	}

	t := time.NewTicker(time.Duration(session.keeptime) * time.Second)
	go func() {
		for {
			select {
			case <-t.C:
				// 指定時間が到来したら、セッションの期限切れチェックを行う
				session.inspection()
			}
		}
	}()

	return session
}

// セッションID,トークンIDを生成する
func (session *Session) generateID() string {
	// セッションIDを生成するランダム文字列
	const randomid = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	// シード値を設定
	random.Seed(time.Now().UnixNano())
	// SHA256オブジェクトを生成
	hash := sha256.New()
	// ランダム文字列を生成
	buf := make([]byte, 64)
	for i := range buf {
		buf[i] = randomid[random.Intn(len(randomid))]
	}
	// セッションIDを生成
	io.WriteString(hash, string(buf)+fmt.Sprint(time.Now().UnixNano()))
	return fmt.Sprintf("%X", hash.Sum(nil))
}

// Start : 管理するセッションIDを更新する
func (session *Session) Start(w http.ResponseWriter, r *http.Request) Interface {
	session.mu.Lock()
	defer session.mu.Unlock()

	// クッキー情報を得る
	cookie, err := r.Cookie(session.name)
	// 新しいセッションIDを用いて、Storageを生成
	storage := &Storage{
		Access: time.Now().Unix(),
		ssid:   session.generateID(),
		Values: make(map[string]interface{}),
	}

	// クッキーが存在する場合は、セッションを更新
	if err == nil && cookie.Value != "" {
		if v, ok := session.data[cookie.Value]; ok {
			oldstorage, _ := v.Value.(*Storage)
			oldstorage.Access = 0
			storage.Values = oldstorage.Values
		}
	}

	// 後ろへ追加する
	elem := session.list.PushBack(storage)
	session.data[storage.ssid] = elem

	// クッキー情報を登録する
	http.SetCookie(w, &http.Cookie{
		Name:     session.name,
		Value:    url.QueryEscape(storage.ssid),
		Path:     "/",
		HttpOnly: true,
		Secure:   session.secure,
		MaxAge:   session.maxage,
	})

	// セッション情報を返却する
	return storage
}

// Destroy : セッションを破棄する
func (session *Session) Destroy(w http.ResponseWriter, r *http.Request) {
	// Cookie取得失敗の場合は、何もせず復帰
	cookie, err := r.Cookie(session.name)
	if err != nil || cookie.Value == "" {
		return
	}
	session.mu.Lock()
	defer session.mu.Unlock()

	// プロバイダが所持するセッションを削除
	if v, ok := session.data[cookie.Value]; ok {
		delete(session.data, cookie.Value)
		session.list.Remove(v)
	}

	// Cookie を削除する
	http.SetCookie(w, &http.Cookie{
		Name:     session.name,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now(),
		Secure:   session.secure,
		MaxAge:   -1,
	})
}

// 期限切れのセッションIDを検査する
func (session *Session) inspection() {
	session.mu.Lock()
	defer session.mu.Unlock()

	for {
		// 前から探す
		elem := session.list.Front()
		if elem == nil {
			break
		}
		storage := elem.Value.(*Storage)

		accesstime := storage.Access + int64(session.maxage)
		if accesstime < time.Now().Unix() {
			// 期限切れのセッションが存在した場合、破棄する
			delete(session.data, storage.ssid)
			session.list.Remove(elem)
		} else {
			// 古い順で参照しているため、ここに到達した時点で期限内のセッションとなる。
			// よって、これ以降のセッション情報を参照したところで意味はないためループを抜ける
			break
		}
	}
}
