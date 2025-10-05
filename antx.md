# Descrição Detalhada do Projeto: antx

## Visão Geral

O `antx` é uma ferramenta de linha de comando (CLI) desenvolvida em Go, projetada para fornecer uma interface interativa, semelhante a um shell, para gerenciar e operar implantações do Antbox. Ele permite que os usuários interajam com o sistema de arquivos do Antbox de maneira eficiente, oferecendo comandos para manipulação de arquivos, navegação em pastas, uploads, downloads e gerenciamento de "nós" (arquivos e pastas).

## Arquitetura e Estrutura do Código

O projeto é modular e bem estruturado, com uma clara separação de responsabilidades entre os diferentes pacotes.

### `main.go`

O ponto de entrada da aplicação. Ele simplesmente chama a função `Execute` do pacote `cmd`.

### Pacote `cmd`

-   **`root.go`**: Utiliza a biblioteca `cobra` para definir o comando raiz `antx`, seus argumentos e flags. Ele lida com a análise dos argumentos da linha de comando (URL do servidor, chave de API, senha root, JWT) e inicia a CLI interativa.

### Pacote `cli`

-   **`prompt.go`**: O coração da interface do usuário. Ele usa a biblioteca `go-prompt` para criar um shell interativo com recursos avançados:
    -   **Executor de Comandos**: Processa e executa os comandos inseridos pelo usuário (`ls`, `cd`, `mkdir`, etc.).
    -   **Auto-complete**: Oferece sugestões de comandos e argumentos (nomes de arquivos/pastas e UUIDs) para facilitar o uso. As sugestões são contextuais (por exemplo, `cd` sugere apenas pastas).
    -   **Gerenciamento de Estado**: Mantém o estado da pasta atual (`currentFolder`) e os nós na pasta atual (`currentNodes`).

### Pacote `antbox`

Este pacote implementa o cliente para a API REST do Antbox.

-   **`antbox.go`**: Define a interface `Antbox`, que abstrai as operações do cliente, como login, obtenção de nós, listagem de nós, etc.
-   **`client.go`**: A implementação da interface `Antbox`. Ele lida com a construção e envio de requisições HTTP para o servidor Antbox, bem como o tratamento das respostas.
    -   **Autenticação**: Suporta autenticação via chave de API, senha de root (que é hasheada antes de ser enviada) ou token JWT.
    -   **Operações de CRUD**: Implementa a lógica para criar, ler, atualizar e deletar nós.
    -   **Upload e Download**: Gerencia uploads de arquivos (usando `multipart/form-data`) e downloads de arquivos.
-   **`types.go`**: Define as estruturas de dados usadas no cliente, como `Node` (representando um arquivo ou pasta), `Permissions` e `HttpError`. A estrutura `HttpError` é particularmente útil para depuração, pois captura detalhes completos da requisição e resposta HTTP em caso de erro.

### Outros Arquivos

-   **`go.mod`**: Define o módulo Go e suas dependências, como `github.com/c-bata/go-prompt` e `github.com/spf13/cobra`.
-   **`openapi.yaml`**: A especificação da API OpenAPI (v3.1.0) para o serviço Antbox. Descreve todos os endpoints, parâmetros e esquemas de dados, servindo como uma documentação valiosa para a API.
-   **`Makefile`**: Fornece comandos para construir (`build`) e testar (`test`) o projeto.
-   **`PROJECT_OVERVIEW.md`**: O documento original que fornece uma visão geral de alto nível do projeto.

## Funcionalidades Principais

-   **Operações de Sistema de Arquivos**: `ls`, `cd`, `pwd`, `mkdir`, `rm`, `mv`, `rename`.
-   **Transferência de Arquivos**: `cp` (upload de arquivos locais) e `get` (download de nós para a pasta `~/Downloads`).
-   **Inspeção de Nós**: `stat` para exibir metadados detalhados de um nó.
-   **Busca**: `find` para buscar nós com base em critérios de filtragem.
-   **Experiência do Usuário**:
    -   Auto-complete interativo e inteligente.
    -   Sugestões contextuais (por exemplo, apenas pastas para `cd`).
    -   Operações baseadas em UUID com nomes amigáveis para o usuário.
-   **Tratamento de Erros Robusto**: Mensagens de erro HTTP detalhadas com informações de requisição/resposta.

## Como Funciona

1.  O usuário inicia o `antx` com a URL do servidor e as credenciais de autenticação.
2.  O pacote `cmd` analisa os argumentos e chama a função `cli.Start`.
3.  `cli.Start` inicializa o cliente `antbox`, faz o login (se necessário) e inicia o prompt interativo.
4.  O usuário digita comandos no prompt.
5.  O `executor` em `cli/prompt.go` chama a função correspondente ao comando.
6.  A função do comando usa o cliente `antbox` para fazer requisições à API do Antbox.
7.  O cliente `antbox` envia a requisição HTTP e retorna a resposta ou um erro.
8.  A função do comando exibe o resultado para o usuário.
9.  O `completer` em `cli/prompt.go` fornece sugestões dinâmicas com base no que o usuário está digitando e no contexto atual (comando, pasta atual).
