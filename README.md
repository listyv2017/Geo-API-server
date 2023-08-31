# Finatextグループ ソフトウェアエンジニア選考課題

このREADMEはfinatext インターンシップ ソフトウェアエンジニア選考課題に対するプログラムの説明書です

## step1

`GET http://localhost:8080`でアクセスできるAPIサーバを⽴ち上げる

```bash
go run 01.go &
```

`GET http://localhost:8080`でサーバーにアクセスする

```bash
curl  http://localhost:8080
```

以下のレスポンスが返ってくる。

```txt
Hello
```

## step2

リクエストパラメータに7桁の郵便番号を与えると、指定した形式のJSONで該当する住所を返すAPIを実装した。

APIエンドポイントは、以下の仕様を満たす。

### Request

    ・Method: GET
    ・Path: : /address?postal_code=[郵便番号7桁(ハイフン無し)]

### Response

    ・postal_code : リクエストパラメータで与えた郵便番号
    ・hit_count : 該当する地域の数
    ・address : 上記外部APIから取得した各住所のうち、共通する部分の住所
    ・tokyo_sta_distance : 上記外部APIから取得した各住所のうち、東京駅から最も離れている地域から東京駅までの距離 [km]


住所の取得には以下の外部APIを利⽤します。下記APIは postal に7桁の郵便番号を与えるとJSON形式で住所を返します。
[外部API] `https://geoapi.heartrails.com/api/json?method=searchByPostal&postal=9220004`

以下のコマンドでサーバー`:8080`を起動させる。

```bash
go run 02.go &
```

`GET http://localhost:8080/address?postal_code=5016121` でサーバーにアクセスする。

```bash
curl http://localhost:8080/address?postal_code=5016121
```

以下のレスポンスが返ってくる。

```txt
{
    "postal_code": "5016121",
    "hit_count": 6,
    "address": "岐阜県岐阜市柳津町",
    "tokyo_sta_distance": 278.3
}
```


## step3-1

step2のAPIをコールした時に、データベースにアクセスログを保存するように処理を追加した。データベースには、MySQLを使用した。データベース名は`finatext`, テーブル名は`access_logs` アクセスログの仕様は下記の通り。


テーブル定義

| カラム      | 型       | 説明                                |
|------------|----------|-------------------------------------|
|id          | INT      | PRIMARY KEY                         |
|postal_code | VARCHAR  | リクエストされた郵便番号(ハイフンなし) |
|created_at  | DATETIME | レコード作成⽇時                     |
|

step2のときと同様に、サーバーにアクセスする。

```bash
go run 03_01.go &
```

保存されたログにアクセスには、まず以下のコマンドでMySQLにログインする
尚且つ、データベース`finatext`, テーブル`access_logs`には、以下のようにPRIMARY KEY, リクエストされた郵便番号,　レコード作成日時が保存される。

MySQLにログインする。

```bash
mysql -u root-p
```

ログインした後、次のコマンドで使用するデータベース、テーブルを指定する。
その後、テーブルの中身を表示させる。

```bash
USE finatext;
USE access_logs;
SELECT * FROM access_logs;
```

入力後、以下のような結果が返ってくる。

```txt
+----+-------------+---------------------+
| id | postal_code | created_at          |
+----+-------------+---------------------+
|  1 | 5016121     | 2023-08-11 05:28:48 |
|  2 | 5016121     | 2023-08-11 05:28:51 |
|  3 | 5016121     | 2023-08-11 05:48:29 |
+----+-------------+---------------------+
```

MySQLを終了させるには、以下のコマンドを入力する

```bash
quit
```

## step3-2

郵便番号ごとに、何回リクエストが来たかを集計してレスポンスを返す下記APIエンドポイントを作成した。APIの使用は以下の通り。
リクエスト回数が0回のものはリストに含めていない。

step2と同じような手順で、サーバーにアクセスする。

### Request

    ・Method: GET
    ・Path: : /address?postal_code=[郵便番号7桁(ハイフン無し)]

### Response

    ・access_logs: アクセスログ集計のリスト(リクエスト回数の降順でソート)
        ・postal_code: 郵便番号7桁(ハイフン無し)
        ・request_count: リクエスト回数


以下のコマンドでアクセスログの集計リストを表示させる

```bash
curl http://localhost:8080/address/access_logs
```

以下のようなレスポンスが返ってくる。例は、`5016121`を3回、`1770033`を1回検索した時の出力である。

```text

{
    "access_logs": [
        {
            "postal_code": "5016121",
            "request_count": 3
        },
        {
            "postal_code": "1770033",
            "request_count": 1
        }
    ]
}
```

一度も郵便番号を検索していないときのアクセスログは以下のように表示される。

```text
{
    "access_logs": null
}
```




## step4

step3までで実装したサーバーを、docker-compose.ymlを使用してコンテナとして起動できるようにした。yamlファイルにて環境変数は、各々設定する。

以下のコマンドを使用して、コンテナdetachモードで作成、起動する

```bash
docker-compose up -d
```

その後に、`docker exec`コマンドを使用して起動したコンテナに対してbashシェルを実行する

```bash
docker exec -it apiserver_ubuntu_1 /bin/bash
```

検索したい郵便番号を`curl`コマンドを用いて検索する。

```bash

curl http://localhost:8080/address?postal_code=5016121

```

これによる検索結果の出力はJSON形式でlogファイルとMySQLデータベースに保存される。確認方法は以下の通り。

### logファイル

以下のコマンドを用いて、サーバーコンテナ内で行われた操作に対する標準出力の表示を行う。docker-compose内でのサーバーコンテナ名は　`apiserver_ubuntu_1`

```bash
sudo docker logs apiserver_ubuntu_1
```

以下のような出力が返ってくる

```txt
Database created successfully.
Database and table created successfully.
Server listening on :8080...
{
    "postal_code": "5016121",
    "hit_count": 6,
    "address": "岐阜県岐阜市柳津町",
    "tokyo_sta_distance": 278.3
}
```

`hit_count`は郵便番号の検索結果に該当する地域名、`tokyo_sta_distance`はそれらの値域と東京駅との距離を計算した際に、一番遠い地域との直線距離　を表している。

### MySQLデータベース

今回、検索するごとにアクセスログをMySQLデータベースに保存している。データベース名は`finatext`、テーブル名は`access_logs`

アクセスログを確認するには、まず最初に以下のコマンドでMySQLコンテナにログインしbashを起動させる。

```bash
docker exec -it apiserver_mysql_1 bash
```

MySQLにrootユーザーでログインする。

```bash
mysql -u root -p
```

以下の操作によりテーブル`access_logs`にアクセスする。

```bash
USE finatext;
SELECT * FROM access_logs; 
```

以下のようなレスポンスが返ってくる。

```text

+----+-------------+---------------------+
| id | postal_code | created_at          |
+----+-------------+---------------------+
|  1 | 5016121     | 2023-08-15 02:32:38 |
+----+-------------+---------------------+
1 row in set (0.00 sec)

```

`exit`コマンドでログアウトする。

このようにして、Docker-composeを用いて、一連のプログラムを実行させることができる。
