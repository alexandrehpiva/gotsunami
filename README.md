# GoTsunami

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)]()

**GoTsunami** é uma ferramenta CLI enterprise-grade para testes de carga de APIs REST (HTTP/HTTPS) com foco em produção e integração contínua.

## 🚀 Características

- **Interface CLI intuitiva** com comandos simples e flags poderosas
- **Configuração via JSON** para cenários complexos de teste
- **Métricas em tempo real** com relatórios detalhados
- **Validação avançada** de respostas HTTP
- **Padrões de carga flexíveis** (steady, spike, ramp-up, stress)
- **Suporte completo a HTTP/HTTPS** com connection pooling otimizado
- **Integração CI/CD** com exit codes padronizados
- **Arquitetura modular** preparada para extensão

## 📦 Instalação

### Pré-requisitos

- Go 1.21 ou superior
- Git

### Instalação via Go

```bash
go install github.com/alexandrehpiva/gotsunami/cmd/gotsunami@latest
```

### Instalação via Build Local

```bash
git clone https://github.com/alexandrehpiva/gotsunami.git
cd gotsunami
make build
```

## 🎯 Quick Start

### 1. Criar um cenário de teste

Crie um arquivo `scenario.json`:

```json
{
  "name": "api_health_check",
  "description": "Teste de saúde da API",
  "method": "GET",
  "url": "/api/v1/health",
  "base_url": "https://httpbin.org",
  "headers": {
    "Content-Type": "application/json",
    "User-Agent": "GoTsunami/1.0"
  },
  "validation": {
    "status_codes": [200],
    "response_time_max": "2s",
    "body_contains": ["status"]
  }
}
```

### 2. Executar o teste

```bash
# Teste básico
gotsunami run scenario.json

# Teste com métricas em tempo real
gotsunami run scenario.json --live --vus 10 --duration 30s

# Teste com padrão de picos
gotsunami run scenario.json --pattern spike --vus 50 --duration 60s
```

### 3. Validar cenário

```bash
gotsunami validate scenario.json
```

## 📋 Comandos

### `gotsunami run <scenario.json>`

Executa um teste de carga baseado em um cenário JSON.

**Flags principais:**
- `--vus int`: Número de usuários virtuais (padrão: 10)
- `--duration duration`: Duração do teste (padrão: 30s)
- `--pattern string`: Padrão de carga (steady, spike, ramp-up, stress)
- `--live`: Mostrar métricas em tempo real
- `--quiet`: Modo silencioso (apenas erros)
- `--verbose`: Output detalhado

**Exemplo:**
```bash
gotsunami run scenario.json --vus 50 --duration 2m --pattern spike --live
```

### `gotsunami validate <scenario.json>`

Valida um arquivo de cenário sem executar o teste.

**Exemplo:**
```bash
gotsunami validate scenario.json
```

### `gotsunami version`

Mostra informações de versão e build.

**Exemplo:**
```bash
gotsunami version
```

## ⚙️ Configuração de Cenários

### Estrutura do Arquivo JSON

```json
{
  "name": "nome_do_teste",
  "description": "Descrição do teste",
  "method": "GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS",
  "url": "/endpoint/path",
  "base_url": "https://api.example.com",
  "headers": {
    "Authorization": "Bearer {{token}}",
    "Content-Type": "application/json"
  },
  "query_params": {
    "param1": "value1",
    "param2": "{{variable}}"
  },
  "body": {
    "key": "value"
  },
  "timeout": "30s",
  "retry": {
    "attempts": 3,
    "backoff": "exponential",
    "max_delay": "5s"
  },
  "validation": {
    "status_codes": [200, 201],
    "response_time_max": "2s",
    "body_contains": ["success"],
    "body_not_contains": ["error"],
    "body_regex": "\\\"status\\\":\\\"ok\\\"",
    "body_json_path": "$.data.id"
  },
  "environment": {
    "env": "production"
  },
  "variables": {
    "token": "{{env.API_TOKEN}}",
    "user_id": "{{random.uuid}}"
  }
}
```

### Validação de Resposta

O GoTsunami suporta validação avançada de respostas:

- **Status Codes**: Múltiplos códigos aceitos
- **Tempo de Resposta**: Limite máximo configurável
- **Conteúdo do Body**: Contains, not contains, regex, JSON path
- **Headers**: Validação de headers específicos
- **Tamanho da Resposta**: Limites mínimo e máximo

### Variáveis e Templates

- `{{env.VARIABLE}}`: Variáveis de ambiente
- `{{random.uuid}}`: UUID aleatório
- `{{random.string}}`: String aleatória
- `{{timestamp}}`: Timestamp atual

## 📊 Padrões de Carga

### Steady (Constante)
Carga constante durante todo o teste.

```bash
gotsunami run scenario.json --pattern steady
```

### Spike (Picos)
Simula picos de tráfego.

```bash
gotsunami run scenario.json --pattern spike
```

### Ramp-up (Crescimento)
Crescimento linear da carga.

```bash
gotsunami run scenario.json --pattern ramp-up
```

### Stress (Estresse)
Teste de estresse com carga crescente.

```bash
gotsunami run scenario.json --pattern stress
```

## 📈 Métricas e Relatórios

### Métricas em Tempo Real

Use a flag `--live` para ver métricas em tempo real:

```bash
gotsunami run scenario.json --live
```

### Relatórios JSON

```bash
# Salvar relatório em arquivo
gotsunami run scenario.json --outfile report.json

# Output para stdout (CI/CD)
gotsunami run scenario.json --stdout
```

### Exemplo de Relatório

```json
{
  "metadata": {
    "tool": "GoTsunami",
    "version": "1.0.0",
    "timestamp": "2024-01-15T10:30:00Z",
    "duration": "30s",
    "scenario": "api_health_check"
  },
  "summary": {
    "total_requests": 1500,
    "successful_requests": 1485,
    "failed_requests": 15,
    "success_rate": 99.0
  },
  "latency": {
    "mean": "245ms",
    "median": "198ms",
    "p90": "456ms",
    "p95": "678ms",
    "p99": "1.2s"
  },
  "throughput": {
    "requests_per_second": 49.2,
    "bytes_per_second": 1024000
  }
}
```

## 🔧 Configuração Avançada

### Variáveis de Ambiente

Crie um arquivo `.env`:

```env
API_BASE_URL=https://api.example.com
API_TOKEN=your_token_here
DEFAULT_VUS=10
DEFAULT_DURATION=30s
LOG_LEVEL=info
```

### Flags Avançadas

```bash
# Configurações de rede
gotsunami run scenario.json \
  --connections 200 \
  --keep-alive \
  --timeout 60s \
  --proxy http://proxy:8080

# Configurações de workers
gotsunami run scenario.json \
  --workers 8 \
  --ramp-up 10s \
  --ramp-down 5s

# Validação customizada
gotsunami run scenario.json \
  --expect-status 200,201 \
  --expect-body "success" \
  --expect-response-time 2s
```

## 🚀 Integração CI/CD

### GitHub Actions

```yaml
name: Load Test
on: [push, pull_request]

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      
      - name: Install GoTsunami
        run: go install github.com/alexandrehpiva/gotsunami/cmd/gotsunami@latest
      
      - name: Run Load Test
        run: |
          gotsunami run tests/scenarios/api_test.json \
            --vus 10 \
            --duration 30s \
            --stdout \
            --expect-status 200
        env:
          API_TOKEN: ${{ secrets.API_TOKEN }}
```

### Exit Codes

- `0`: Sucesso
- `1`: Erro geral
- `2`: Validação falhou (success rate < 95%)

## 📁 Estrutura do Projeto

```
gotsunami/
├── cmd/gotsunami/          # CLI principal
├── internal/
│   ├── cli/               # Comandos CLI
│   ├── config/            # Configuração
│   ├── engine/            # Engine de load testing
│   ├── protocols/         # Protocolos (HTTP, etc.)
│   ├── metrics/           # Coleta de métricas
│   ├── validation/        # Validação de resposta
│   └── reporting/         # Geração de relatórios
├── pkg/                   # Pacotes utilitários
├── examples/              # Exemplos e cenários
├── tests/                 # Testes
└── docs/                  # Documentação
```

## 🧪 Testes

```bash
# Executar todos os testes
make test

# Testes com cobertura
make test-coverage

# Testes de integração
make test-integration

# Benchmarks
make benchmark
```

## 🛠️ Desenvolvimento

### Setup do Ambiente

```bash
# Clone o repositório
git clone https://github.com/alexandrehpiva/gotsunami.git
cd gotsunami

# Instale dependências
make deps

# Configure ambiente
make dev-setup

# Build
make build
```

### Comandos de Desenvolvimento

```bash
# Desenvolvimento rápido
make dev

# Linting
make lint

# Formatação
make fmt

# Limpeza
make clean
```

## 📚 Exemplos

Veja a pasta `examples/` para cenários de teste completos:

- `basic_get.json`: Teste básico GET
- `post_with_auth.json`: POST com autenticação
- `complex_validation.json`: Validação avançada

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🆘 Suporte

- **Documentação**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/alexandrehpiva/gotsunami/issues)
- **Discussions**: [GitHub Discussions](https://github.com/alexandrehpiva/gotsunami/discussions)

## 🎯 Roadmap

- [ ] Suporte a WebSockets
- [ ] Suporte a gRPC
- [ ] Suporte a GraphQL
- [ ] Interface web para monitoramento
- [ ] Integração com Prometheus/Grafana
- [ ] Suporte a múltiplos protocolos simultâneos
- [ ] Sistema de plugins

---

**GoTsunami** - Teste de carga enterprise para APIs REST 🚀
