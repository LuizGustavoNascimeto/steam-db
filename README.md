# Steam DB

Um projeto em Go para gerenciar dados de jogos usando a API RAWG e banco de dados PostgreSQL.

## Pré-requisitos

- Go 1.25.1 ou superior
- PostgreSQL
- Chave da API RAWG (disponível em [rawg.io](https://rawg.io/apidocs))

## Configuração

### 1. Clone o repositório

```bash
git clone <url-do-repositorio>
cd steam-db
```

### 2. Configure o banco de dados PostgreSQL

Crie um banco de dados PostgreSQL com as seguintes configurações:
- Host: `localhost`
- Usuário: `postgres`
- Senha: `root`
- Nome do banco: `rawg-db`
- Porta: `5432`

### 3. Instale as dependências

```bash
go mod download
```

### 4. Configure a chave da API RAWG

No arquivo `main.go`, substitua a chave da API na linha:

```go
url := fmt.Sprintf("https://api.rawg.io/api/games/%d?key=SUA_CHAVE_AQUI", rawgID)
```

## Executando o projeto

### Executar diretamente

```bash
go run main.go
```

### Compilar e executar

```bash
go build -o steam-db
./steam-db
```

## Estrutura do projeto

- `main.go` - Arquivo principal com a lógica de busca e atualização de jogos
- `types/types.go` - Definições de tipos e estruturas de dados
- `go.mod` - Configuração do módulo Go e dependências

## Funcionalidades

O projeto atualmente:

1. Conecta-se ao banco de dados PostgreSQL
2. Executa migrações automáticas das tabelas usando GORM
3. Busca jogos existentes no banco
4. Atualiza informações dos jogos usando a API RAWG
5. Implementa rate limiting para respeitar os limites da API

## Estruturas de dados

O projeto trabalha com as seguintes entidades principais definidas em `types/types.go`:

- `Game` - Informações principais do jogo
- `Platform` - Plataformas de jogos
- `Store` - Lojas digitais
- `Genre` - Gêneros de jogos
- `Tag` - Tags dos jogos
- `Developer` - Desenvolvedores

## Dependências

- `gorm.io/gorm` - ORM para Go
- `gorm.io/driver/postgres` - Driver PostgreSQL para GORM