package session

import (
	"fmt"
	"testing"
)

func Test__STORAGE_STRING(t *testing.T) {
	storage := &Storage{
		ssid:   "SSID",
		Values: make(map[string]interface{}),
	}

	// SessionID を取得できるか確認する
	if storage.SessionID() != "SSID" {
		t.Fatal("Storage.SessionID: FATAL")
	}

	// Set Value
	storage.Set("NAME", "SAMPLE")

	value, err := storage.Get("NAME")
	// エラーが発生した場合、テスト失敗
	if err != nil {
		t.Fatal("Storage.Get: FATAL")
	}
	// value の値が SAMPLE ではない場合、テスト失敗
	if fmt.Sprint(value) != "SAMPLE" {
		t.Fatal("Storage.Get: FATAL")
	}

	str, err := storage.Str("NAME")
	// エラーが発生した場合、テスト失敗
	if err != nil {
		t.Fatal("Storage.Str: FATAL")
	}
	if str != "SAMPLE" {
		t.Fatal("Storage.Str: FATAL")
	}
	str, err = storage.Str("NONAME")
	// 存在しない名前でGetされた場合、エラーにならなければテスト失敗
	if err == nil {
		t.Fatal("Storage.Str: FATAL")
	}

	// 存在するデータを削除
	storage.Delete("NAME")
	// 削除後、データが存在しないか確認する
	if _, err := storage.Get("NAME"); err == nil {
		t.Fatal("Storage.Delete: Fatal")
	}
}

func Test__STORAGE_BOOL(t *testing.T) {
	storage := &Storage{
		Values: make(map[string]interface{}),
	}

	// Set Value
	storage.Set("BOOL", true)

	bFlag, err := storage.Bool("BOOL")
	// Set された値が true ではない場合テスト失敗
	if err != nil || bFlag != true {
		t.Fatal("Storage.Bool: FATAL")
	}

	bFlag, err = storage.Bool("NOBOOL")
	// 存在しないキー名を指定された場合、エラーにならなければテスト失敗
	if err == nil {
		t.Fatal("Storage.Bool: FATAL")
	}

	storage.Set("BOOL", "true")
	bFlag, err = storage.Bool("BOOL")
	// 存在するが、文字列の"true"はエラーを返す
	if err == nil {
		t.Fatal("Storage.Bool: FATAL")
	}
}

func Test__STORAGE_INT(t *testing.T) {
	storage := &Storage{
		Values: make(map[string]interface{}),
	}

	// Set Value
	storage.Set("INT", 100)

	n, err := storage.Int("INT")
	// Setされた値が整数ではない場合テスト失敗
	if err != nil || n != 100 {
		t.Fatal("Storage.Int: FATAL")
	}

	storage.Set("INT", "100")
	n, err = storage.Int("INT")
	// 文字列の数字を処理できない場合はテスト失敗
	if err != nil || n != 100 {
		t.Fatal("Storage.Int: FATAL")
	}

	storage.Set("INT", "100Z")
	// 整数変換失敗エラーが発生しない場合テスト失敗
	n, err = storage.Int("INT")
	if err == nil {
		t.Fatal("Storage.Int: FATAL")
	}

	// 存在しないキー名を指定された場合のテスト
	if _, err := storage.Int("NOINT"); err == nil {
		t.Fatal("Storage.Int: FATAL")
	}
}

type SampleTest struct {
	Name string
	Info int
}

func Test__STORAGE_STRUCT(t *testing.T) {
	storage := &Storage{
		Values: make(map[string]interface{}),
	}
	// 存在しないキー名を指定された場合のテスト
	if err := storage.Struct("STORAGE", nil); err == nil {
		t.Fatal("Storage.Struct: FATAL")
	}

	// 構造体に値がセットできるかテスト
	storage.Set("STORAGE", map[string]interface{}{
		"name": "struct",
		"info": 2,
	})
	var i = &SampleTest{}
	if err := storage.Struct("STORAGE", i); err != nil {
		t.Fatal("Storage.Struct: FATAL")
	}
	if i.Name != "struct" || i.Info != 2 {
		t.Fatal("Storage.Struct: FATAL")
	}

	// セットしたデータに対して、構造体以外の変数をセットされた場合エラーが発生するかテスト
	var n int
	if err := storage.Struct("STORAGE", &n); err == nil {
		t.Fatal("Storage.Struct: FATAL")
	}

	// セットしたデータをマップに変換できるかテスト
	var m = make(map[string]interface{})
	if err := storage.Struct("STORAGE", &m); err != nil {
		t.Fatal("Storage.Struct: FATAL")
	}
	if fmt.Sprint(m["name"]) != "struct" || fmt.Sprint(m["info"]) != "2" {
		t.Fatal("Storage.Struct: FATAL")
	}

	// セットしたデータが構造体に変換できないかテスト
	storage.Set("STORAGE", map[interface{}]string{nil: "struct"})
	if err := storage.Struct("STORAGE", &m); err == nil {
		t.Fatal("Storage.Struct: FATAL")
	}
}
