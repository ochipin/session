package session

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"
)

func Test__SESSION_NEW(t *testing.T) {
	// セッション管理情報を構築する
	session := New("", 2, 0, false)
	// デフォルトでKeepTimeが1分でない場合テスト失敗
	if session.keeptime != 60 {
		t.Fatal("Session: Fatal")
	}
	session = New("", 2, 1, false)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/destroy" {
			session.Destroy(w, r)
			return
		}

		// セッション管理開始
		sess := session.Start(w, r)
		n, _ := sess.Int("INT")
		sess.Set("INT", n+1)

		if r.URL.Path == "/" {
			fmt.Fprintf(w, "%d", n+1)
		} else if r.URL.Path == "/ssid" {
			fmt.Fprintf(w, "%s", sess.SessionID())
		}
	}))

	client := ts.Client()

	// 1 回目のアクセス開始
	res, _ := client.Get(ts.URL)
	// クッキー情報を保存
	cookies := res.Cookies()
	client.Jar, _ = cookiejar.New(nil)
	client.Jar.SetCookies(res.Request.URL, cookies)
	// Bodyをクローズ
	res.Body.Close()

	// New関数で設定した値どおりのクッキー情報になっているか検証を行う
	if len(cookies) <= 0 {
		t.Fatal("Cookie: Fatal")
	}
	if cookies[0].Name != "noname" {
		t.Fatal("Cookie: Fatal")
	}
	if cookies[0].MaxAge != 2 {
		t.Fatal("Cookie: Fatal")
	}
	if cookies[0].Secure != false {
		t.Fatal("Cookie: Fatal")
	}
	if cookies[0].Path != "/" {
		t.Fatal("Cookie: Fatal")
	}
	if cookies[0].HttpOnly != true {
		t.Fatal("Cookie: Fatal")
	}

	// 2回目のアクセス開始
	res, _ = client.Get(ts.URL + "/ssid")
	// クッキー情報を保存
	cookies = res.Cookies()
	client.Jar.SetCookies(res.Request.URL, cookies)
	buf, _ := ioutil.ReadAll(res.Body)
	// サーバ上にあるクッキーIDとクライアント側のクッキーIDが違う場合はテスト失敗
	if string(buf) != cookies[0].Value {
		t.Fatal("Cookie: Fatal")
	}
	// Bodyをクローズ
	res.Body.Close()

	// 3回目のアクセス
	res, _ = client.Get(ts.URL)
	// クッキー情報を保存
	cookies = res.Cookies()
	client.Jar.SetCookies(res.Request.URL, cookies)
	buf, _ = ioutil.ReadAll(res.Body)
	// サーバ上でカウントしているアクセス回数と、合わなければテスト失敗
	if string(buf) != "3" {
		t.Fatal("Cookie: Fatal")
	}
	// Bodyをクローズ
	res.Body.Close()

	time.Sleep(3 * time.Second)

	defer ts.Close()
}

func Test__SESSION(t *testing.T) {
	// セッション管理情報を構築する
	session := New("", 2, 0, false)
	// デフォルトでKeepTimeが1分でない場合テスト失敗
	if session.keeptime != 60 {
		t.Fatal("Session: Fatal")
	}
	session = New("", 2, 1, false)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/destroy" {
			session.Destroy(w, r)
			return
		}

		// セッション管理開始
		sess := session.Start(w, r)
		n, _ := sess.Int("INT")
		sess.Set("INT", n+1)

		if r.URL.Path == "/" {
			fmt.Fprintf(w, "%d", n+1)
		} else if r.URL.Path == "/ssid" {
			fmt.Fprintf(w, "%s", sess.SessionID())
		}
	}))

	client := ts.Client()

	// 1 回目のアクセス開始
	res, _ := client.Get(ts.URL)
	// クッキー情報を保存
	cookies := res.Cookies()
	client.Jar, _ = cookiejar.New(nil)
	client.Jar.SetCookies(res.Request.URL, cookies)
	// Bodyをクローズ
	res.Body.Close()

	// New関数で設定した値どおりのクッキー情報になっているか検証を行う
	if len(cookies) <= 0 {
		t.Fatal("Cookie: Fatal")
	}
	if cookies[0].Name != "noname" {
		t.Fatal("Cookie: Fatal")
	}
	if cookies[0].MaxAge != 2 {
		t.Fatal("Cookie: Fatal")
	}
	if cookies[0].Secure != false {
		t.Fatal("Cookie: Fatal")
	}
	if cookies[0].Path != "/" {
		t.Fatal("Cookie: Fatal")
	}
	if cookies[0].HttpOnly != true {
		t.Fatal("Cookie: Fatal")
	}

	// 2回目のアクセス開始
	res, _ = client.Get(ts.URL + "/ssid")
	// クッキー情報を保存
	cookies = res.Cookies()
	client.Jar.SetCookies(res.Request.URL, cookies)
	buf, _ := ioutil.ReadAll(res.Body)
	// サーバ上にあるクッキーIDとクライアント側のクッキーIDが違う場合はテスト失敗
	if string(buf) != cookies[0].Value {
		t.Fatal("Cookie: Fatal")
	}
	// Bodyをクローズ
	res.Body.Close()

	// 3回目のアクセス
	res, _ = client.Get(ts.URL)
	// クッキー情報を保存
	cookies = res.Cookies()
	client.Jar.SetCookies(res.Request.URL, cookies)
	buf, _ = ioutil.ReadAll(res.Body)
	// サーバ上でカウントしているアクセス回数と、合わなければテスト失敗
	if string(buf) != "3" {
		t.Fatal("Cookie: Fatal")
	}
	// Bodyをクローズ
	res.Body.Close()

	// 4回目のアクセスでセッションを破棄する
	res, _ = client.Get(ts.URL + "/destroy")
	// クッキー情報を保存
	cookies = res.Cookies()
	client.Jar.SetCookies(res.Request.URL, cookies)
	// Bodyをクローズ
	res.Body.Close()

	// 本当に破棄されたかサーバ上に再度問い合わせる
	res, _ = client.Get(ts.URL + "/destroy")
	// クッキー情報を保存
	cookies = res.Cookies()
	client.Jar.SetCookies(res.Request.URL, cookies)
	// Bodyをクローズ
	res.Body.Close()
	if len(cookies) != 0 {
		t.Fatal("Cookie: Fatal")
	}

	defer ts.Close()
}
