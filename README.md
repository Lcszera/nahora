# ⚡ naHora

Rotina diária, tarefas e estudos no Windows — com avisos na hora certa, sem procrastinar.

- **Rotina do dia** com itens fixos (repetem nos dias marcados) e avulsos; aviso X minutos antes.
- **Tarefas** com início e entrega; avisos 2 dias, 1 dia, 12 h e 1 h antes. Concluiu, sumiu.
- **Catálogos** (Faculdade, Trabalho, Estudos, Plano DevJr…) com seções e assuntos de estudo.
- **Temas e wallpapers** (12 artes originais) ou imagem própria; tema escuro/claro.
- **Som de notificação** personalizado (MP3/WAV/WMA).
- Fica na **bandeja** gastando pouquíssima memória e **se atualiza sozinho** pelo GitHub Releases.

## Como publicar uma versão nova
```bash
git tag v1.3.0 -m "O que mudou"
git push origin v1.3.0
```
O GitHub Actions compila o `naHora.exe` e publica no Releases; o app instalado baixa, confere o SHA-256 e se atualiza.

## Desenvolvimento
- Go 1.24, sem dependências externas (Win32 direto via `syscall`).
- Interface em `ui/` (HTML único embutido no executável).
- Wallpapers: `build/wallpapers/gen.py` + `pixel.py` + `render.py`.
- Dados do usuário ficam em `%APPDATA%\naHora\dados.json`.
