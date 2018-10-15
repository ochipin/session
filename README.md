セッション管理ライブラリ
===
セッション管理ライブラリです。セッション情報をオンメモリ上で保管します。


```go
package main

import (
    "fmt"
    "net/http"

    "github.com/ochipin/session"
)

func main() {
    // 86400 秒間有効なクッキーを生成する。
    // 60 秒に1度有効期限切れのクッキーを探す
    sess := session.New("Cookiename", 86400, 60, false)

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // セッション管理開始。アクセスする度にセッションIDは変わる
        sid := sess.Start(w, r)
        // アクセスする度にアクセスカウンターがプラスされる
        n, _ := sid.Int("ACCESS_COUNT")
        sid.Set("ACCESS_COUNT", n+1)

        if r.URL.Path == "/" {
            // アクセスカウントを表示
            fmt.Fprintf(w, "%d", n+1)
        } else if r.URL.Path == "/delete" {
            // セッションを破棄
            sess.Destroy(w, r)
        }
    })
    http.ListenAndServe(":8080", nil)
}
```

## New()
セッション管理構造体を生成します。次の4つの引数を受け取ります。

| 引数    | 名前 | 型 | 説明 |
|:--      |:--   |:-- |:-- |
| 第1引数 | name | string | クッキー名を指定します |
| 第2引数 | maxage | int | クッキーの有効期間を指定します |
| 第3引数 | keep | int | クッキーの有効期限切れを検査する時間を指定します |
| 第4引数 | secure | bool | https 通信のみで有効にしたい場合 true を指定します |

復帰値は、`*Session`型を返却します。

### 使用例:
```go
sess := session.New("Cookiename", 86400, 60, false)
```

## Session.Start() / Session.Destroy()
セッションの作成(Start)、または破棄(Destroy)を行います。
Start, Destroy それぞれの関数の引数は次のとおりです。

| 引数    | 名前 | 型 |
|:--      |:--   |:-- |
| 第1引数 | w | http.ResponseWriter |
| 第2引数 | r | *http.Request |

### 使用例:
```go
sess := session.New("Cookiename", 86400, 60, false)
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    // セッション管理開始
    sid := sess.Start(w, r)

    if r.URL.Path == "/delete" {
        // セッションの破棄
        sess.Destroy(w, r)
    }
})
```

## データのやりとり
セッション情報の取得、設定、削除は Get/Set/Delete を使用して行います。  
また、現在のセッションIDを取得する場合はSessionID関数をコールすることで取得できます。

### 使用例:
```go
sess := session.New("Cookiename", 86400, 60, false)
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    // セッション管理開始
    sid := sess.Start(w, r)

    // 値をセット。KVS形式。
    sid.Set("KEYNAME", "VALUE")

    // 値を取得。キー名を指定することで、interface{}型で取得可能
    value, _ := sid.Get("KEYNAME")
    fmt.Println(value) // "VALUE" を表示

    // 値を削除
    sid.Delete("KEYNAME")

    // セッションIDを表示
    fmt.Println(sid.SessionID())

    // bool 型で取得する
    sid.Set("KEYNAME", true)
    sid.Bool("KEYNAME") // true

    // int 型で取得する
    sid.Set("KEYNAME", 100)
    sid.Int("KEYNAME") // 100

    // string 型で取得する
    sid.Set("KEYNAME", "VALUE")
    sid.Str("KEYNAME") // "VALUE"

    // 構造体で取得する
    sid.Set("KEYNAME", map[string]interface{}{
        "name": "struct",
        "num":  200,
    })
    if err := sid.Struct("KEYNAME", &info); err == nil {
        fmt.Println(info) // struct 200 を表示
    }
})
```

