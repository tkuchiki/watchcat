# watchcat

- **[WIP / TODO] 説明を書き直す** 
- ファイルサイズ、更新日時を監視して、ファイルがしばらく更新されなくなり、指定したバイト数以上増加されたときに、差分を標準出力、またはコマンドを実行して pipe で標準入力に渡します 
    - ベンチマークをするときに、ベンチマークごとにログファイルを空にしなくても、ベンチマークごとのログを取り出すような用途を想定しています
    - 指定したバイト数増分がなければファイルを読み飛ばすので、ベンチマークの合間に手動で操作をしたときのログは除外したい、といったケースにも対応できる想定です

## 使い方

```console
$ watchcat --help
usage: watchcat --file=FILE [<flags>]

watchcat

Flags:
      --help             Show context-sensitive help (also try --help-long and --help-man).
      --interval=1s      interval
      --no-changed=60    no changed seconds
      --filesize=10240   filesize
  -f, --file=FILE        file
  -c, --command=COMMAND  command
      --debug            debug
      --version          Show application version.
```

## Example

- 1秒ごとにファイルサイズ・更新日時を監視して、更新されなくなってから 60 秒経過し、1024B 以上更新されていたら command.sh を実行する
    - 以下の例では、条件を満たしたらファイルの増分に対して wc -l を実行します
    - ファイルに1024B書かれて60秒経過するとコマンドが実行され、また 1024B 書き込まれて60秒経過すると、1025B目から新しく増分した 1024B がコマンドに pipe で渡されます

```console
$ cat command.sh
#!/bin/bash

cat - | wc -l

$ watchcat --interval 1s --no-changes=60 --filesize=1024 --file /path/to/mysql-slow.log --command command.sh 
```