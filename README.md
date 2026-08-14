# Go Base Module

![Current Version](https://img.shields.io/badge/latest-v0.1.0-orange) ![Technology Version](https://img.shields.io/badge/go-%3E%3D1.20-blue)

Este repositório é o pacote base em **Go** utilizado por todos os microsserviços do ecosistema **RestAG**. Ele fornece utilitários compartilhados, padronização de logs, e middlewares de resiliência e segurança, garantindo consistência em toda a arquitetura de backend.

## TL;DR

Para adicionar este módulo ao seu microsserviço:

```bash
go get github.com/atosramos/go-base-module
```

### Inicializando o Logger:
```go
import "github.com/atosramos/go-base-module/logger"

func main() {
    isProduction := true
    log, err := logger.New(isProduction)
    if err != nil {
        panic(err)
    }
    defer log.Sync()

    log.Info("Service started!")
}
```

### Acoplando Middlewares Globais (Gin):
```go
import (
    "github.com/gin-gonic/gin"
    basemiddleware "github.com/atosramos/go-base-module/middleware"
)

func main() {
    engine := gin.New()
    
    // Injeta o logger estruturado nas requisições
    engine.Use(basemiddleware.RequestLoggerMiddleware(log))
    
    // Captura panics e retorna 500 para evitar interrupção do container
    engine.Use(basemiddleware.PanicRecoveryMiddleware(log.Zap()))
}
```

## Arquitetura e Utilitários
Este módulo contém:
- **`logger`**: Um wrapper em cima do `uber-go/zap` padronizado. Lê automaticamente o nível de verbosidade através da variável de ambiente `LOG_LEVEL` (`debug`, `info`, `warn`, `error`).
- **`middleware`**: 
  - `RequestLoggerMiddleware`: Monitora tempo de resposta, path, método HTTP e captura o status code para definir o nível do log (INFO para 2xx/3xx, WARN para 4xx e ERROR para 5xx).
  - `PanicRecoveryMiddleware`: Garante que o servidor não sofra *crash* devido a ponteiros nulos ou panics inesperados, logando o stack trace e retornando erro 500 limpo para o cliente.

---

## 📜 Regras do Repositório (Governance)

Para manter a integridade, performance e baixo acoplamento nos microsserviços, o desenvolvimento neste repositório deve seguir as regras abaixo:

1. **ZERO Regras de Negócio (Business Logic):**
   - Este repositório é estritamente voltado para **infraestrutura de código** (logs, métricas, clientes HTTP base, wrappers de banco de dados).
   - NUNCA adicione lógicas relacionadas a domínios do RestAG (ex: validação de estoque, regras de catálogo, etc).

2. **Gerenciamento Restrito de Dependências:**
   - Evite adicionar dependências externas pesadas no `go.mod`.
   - Lembre-se: todo microsserviço que importar o `go-base-module` também fará o download de todas as dependências aqui contidas. Mantenha-o enxuto.

3. **Retrocompatibilidade Rigorosa (Backward Compatibility):**
   - Nenhuma assinatura de função pública (`Exported Functions` com letra maiúscula) deve ter seus parâmetros ou retornos alterados de forma destrutiva após publicadas.
   - Caso precise criar uma nova versão incompatível, crie uma versão `V2` do pacote ou da função (Ex: `logger.NewV2()`) e deprecie a antiga.

4. **Versionamento Semântico Rigoroso (SemVer):**
   - Todas as modificações aprovadas neste repositório devem ser tageadas usando [SemVer](https://semver.org/).
   - `MAJOR`: Quando quebrar compatibilidade (ex: mudar o logger de Zap para Logrus, alterando a interface).
   - `MINOR`: Quando adicionar novas funcionalidades de forma retrocompatível (ex: novo middleware).
   - `PATCH`: Quando realizar correções de bugs (ex: consertar um panic no middleware de log).

5. **Testes Unitários:**
   - Todo código base compartilhado exige cobertura de testes próxima a 100%. Falhas aqui se propagam para dezenas de microsserviços simultaneamente.

6. **Não utilizar estado global não-seguro (Goroutine Safe):**
   - O código deve ser seguro para concorrência (thread-safe), dado que será utilizado ativamente por servidores web multi-thread (`gin`). Evite variáveis globais mutáveis.
