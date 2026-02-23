# Octane Collection Tool (Go CLI)

## Features

- Liest JUnit XML und erzeugt Octane-internes `<test_result>` XML.
- Optionales Schreiben der internen XML-Datei über `--output-file`.
- Direkter Upload zu OpenText ALM Octane (`test-results` API).
- Optionales Polling des Result-Status (`--check-result`).
- Konfiguration via `config.properties`.
- Unterstützung von `--tag TYPE:VALUE` und `--field TYPE:VALUE`.

## Build

```bash
go build ./cmd/octane-collection-tool
```

## Nutzung

```bash
./octane-collection-tool [optionen] <junit-oder-internal-xml-dateien...>
```

Beispiel: Interne XML schreiben

```bash
./octane-collection-tool --output-file internal.xml ./report.xml
```

Beispiel: Push nach Octane

```bash
./octane-collection-tool \
  -s "http://octane.example.com:8080" \
  -d 1001 \
  -w 1002 \
  -u "client-id" \
  -p "client-secret" \
  -t "OS:Linux" \
  -f "Framework:JUnit" \
  --release-default \
  ./report.xml
```

## Docker Bereitstellung

```
image: ghcr.io/MASYONY/octane-collection-tool
```

## Konfiguration

`config.properties` (im Arbeitsverzeichnis oder über `--config-file` OPTIONAL):

```properties
server=http://octane.example.com:8080
sharedspace=1001
workspace=1002
user=client-id
password=client-secret
```


## Flags (`octane-collection-tool`)

### Lange Flags
- `--config-file <pfad>`: Pfad zur `config.properties`.
- `--server <url>`: Octane Server-URL.
- `--shared-space <id>`: Shared Space ID.
- `--workspace <id>`: Workspace ID.
- `--user <name>`: Benutzername / Client-ID.
- `--password <secret>`: Passwort / Client-Secret.
- `--access-token <token>`: Bearer Token für Auth.
- `--tag TYPE:VALUE`: Environment-Tag (mehrfach nutzbar).
- `--field TYPE:VALUE`: Test-Feld (mehrfach nutzbar).
- `--release <id>`: Release-ID setzen.
- `--release-default`: Default-Release `_default_` setzen.
- `--suite <id>`: Suite-ID setzen.
- `--started <epoch-ms>`: Startzeit in Millisekunden.
- `--internal`: Input ist bereits internes Octane-XML.
- `--output-file <datei>`: Internes XML lokal schreiben statt pushen.
- `--check-result`: Nach Push den Status pollen.
- `--check-result-timeout <sek>`: Timeout fürs Polling (Default: `10`).

### Kurze Flags
- `-c <pfad>`: wie `--config-file`
- `-s <url>`: wie `--server`
- `-d <id>`: wie `--shared-space`
- `-w <id>`: wie `--workspace`
- `-u <name>`: wie `--user`
- `-p <secret>`: wie `--password`
- `-t TYPE:VALUE`: wie `--tag` (mehrfach)
- `-f TYPE:VALUE`: wie `--field` (mehrfach)
- `-r <id>`: wie `--release`
- `-i`: wie `--internal`
- `-o <datei>`: wie `--output-file`