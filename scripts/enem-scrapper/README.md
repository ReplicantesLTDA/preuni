# ENEM Scrapper ✅

Descrição curta

- Script para baixar questões e imagens do Repertório ENEM e gerar um JSON organizado (`banco_questoes_completo.json`) e uma pasta `banco_de_questoes/` com imagens.

Requisitos 🔧

- Python 3.8+
- Google Chrome instalado
- Chromedriver compatível com sua versão do Chrome (no PATH)
- Dependências Python: `selenium`, `requests`, `python-dotenv`

Instalação rápida 💡

1. Crie e ative um ambiente virtual (opcional):

```
python -m venv .venv
source .venv/bin/activate
```

2. Instale dependências:

```
pip install selenium requests python-dotenv
```

Configuração — `.env` 🔒

- Crie um arquivo `.env` na raiz do projeto com as credenciais:

```
ENEM_EMAIL=seu_email@example.com
ENEM_PASSWORD=sua_senha
```

Criando uma conta no Repertório ENEM 🔐

- Se ainda não tiver conta, crie uma em: https://app.repertorioenem.com.br/register
- Passos rápidos:
  1. Acesse o link acima.
  2. Preencha seu nome, email e senha e confirme a conta pelo email, se solicitado.
  3. Use o mesmo email e senha no arquivo `.env` acima.
- Observação: se o site exigir verificação adicional (captcha, confirmação por email), conclua esse processo antes de rodar o script.

Uso — passo a passo ▶️

1. Certifique-se que o Chrome e o Chromedriver estão instalados e compatíveis.
2. Preencha o `.env` com seu login.
3. Rode o script:

```
python main.py
```

4. Uma janela pedirá o intervalo de anos (ex.: 2009 até 2024). O programa baixa as questões por página e salva:

- `banco_questoes_completo.json` — JSON final com todas as questões coletadas.
- `banco_de_questoes/` — pasta com diretórios por ano e IDs de questão, contendo imagens baixadas.

Observações ⚠️

- Fechar a janela de seleção de anos cancela a execução.
- O script assume que a página do Repertório ENEM não mudou; se mudar, o extrator em `main.py` pode precisar de ajustes.

Licença

- Use conforme necessário.
