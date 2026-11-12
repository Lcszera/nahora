package main

// Conteúdo inicial dos catálogos.
// Plano DevJr: os 39 itens da print "Fundamentos Básicos de Backend", filtrados e
// ordenados pelo perfil Java Backend Jr. (prioridade: 1 essencial, 2 importante,
// 3 diferencial, 4 depois). A Etapa 0 vem do perfil, não da print.

type seedTopic struct {
	prio       int
	src, title string
	note       string
}

type seedGroup struct {
	title, gate string
	topics      []seedTopic
}

var seedCatalogs = []Catalog{
	{Name: "Faculdade", Emoji: "🎓", Color: "#a78bfa"},
	{Name: "Trabalho", Emoji: "💼", Color: "#f59e0b"},
	{Name: "Estudos", Emoji: "📚", Color: "#34d399"},
	{Name: "Plano DevJr", Emoji: "☕", Color: "#3d8bff"},
}

var seedDevJr = []seedGroup{
	{"Etapa 0 · Base Java", "Escrever classes, usar coleções e tratar exceções sem precisar consultar a todo momento.", []seedTopic{
		{1, "Perfil", "POO na prática", "Encapsulamento, herança, polimorfismo; interface x classe abstrata."},
		{1, "Perfil", "Collections", "List, Set e Map: quando usar ArrayList, HashSet e HashMap."},
		{1, "Perfil", "Generics", "Ler e escrever List<T>, Map<K,V> e métodos genéricos simples."},
		{1, "Perfil", "Exceptions", "Checked x unchecked, try/catch/finally, try-with-resources, exceção própria."},
		{1, "Perfil", "Lambda e Stream API", "filter, map, sorted, collect; method references."},
		{1, "Perfil", "Optional", "Evitar null: orElse, orElseThrow, map."},
		{1, "Perfil", "equals() e hashCode()", "Por que andam juntos e o que quebra em HashSet/HashMap."},
		{2, "Perfil", "Java moderno", "records, var, switch expressions, text blocks, pattern matching."},
	}},
	{"Etapa 1 · Git e GitHub", "Trabalhar com branches, resolver um conflito e abrir um pull request sozinho.", []seedTopic{
		{1, "#36", "Git", "status, add, commit, log, diff, branch, merge, stash, .gitignore, conflitos; rebase só o básico."},
		{1, "#37", "GitHub", "push/pull/fetch, pull request, code review, GitHub Flow, boas mensagens de commit."},
	}},
	{"Etapa 2 · Banco de dados relacional", "Modelar tabelas com relacionamentos 1:N e N:N e escrever consultas com JOIN e GROUP BY no PostgreSQL.", []seedTopic{
		{1, "#12", "Bancos relacionais", "Tabelas, registros, colunas e constraints. PostgreSQL é o seu banco principal."},
		{1, "#11", "SQL", "SELECT, WHERE, ORDER BY, GROUP BY/HAVING, DISTINCT, LIMIT, INSERT/UPDATE/DELETE, subqueries, CASE, CTE."},
		{1, "#14", "Modelagem de dados", "Relacionamentos 1:1, 1:N e N:N; normalização até a 3FN."},
		{1, "#15", "Chaves primárias e estrangeiras", "PK, FK, UNIQUE, NOT NULL, ON DELETE; tabela associativa no N:N."},
		{1, "#16", "Joins", "INNER, LEFT e RIGHT JOIN: saber prever o resultado de cada um."},
		{1, "#17", "Índices", "Quando criar, o custo na escrita; ler EXPLAIN e EXPLAIN ANALYZE."},
		{1, "#18", "Transações ACID", "BEGIN/COMMIT/ROLLBACK e o que cada letra do ACID garante."},
		{1, "#26", "SQL Injection", "Por que concatenar SQL é perigoso; PreparedStatement no JDBC e parâmetros na JPA."},
	}},
	{"Etapa 3 · Web e APIs REST com Spring Boot", "Explicar o caminho de uma requisição até o banco e devolver o status HTTP certo em cada caso.", []seedTopic{
		{1, "#1", "Client-server", "Quem é cliente, quem é servidor e o que acontece numa requisição."},
		{1, "#2", "HTTP e HTTPS", "Request, response, headers e body; o que o TLS do HTTPS protege."},
		{1, "#4", "Métodos HTTP", "GET, POST, PUT, PATCH, DELETE e idempotência."},
		{1, "#5", "Status HTTP", "200, 201, 204, 400, 401, 403, 404, 409, 422, 500: quando usar cada um."},
		{1, "#6", "JSON", "Estrutura e tipos; serialização com Jackson no Spring."},
		{1, "#3", "API RESTful", "Recursos e URIs, path variable x query param, ResponseEntity, paginação e filtros."},
		{1, "#7", "CRUD", "Create/Read/Update/Delete ligados aos métodos HTTP e ao SQL."},
		{1, "#29", "Arquitetura MVC em camadas", "Controller → Service → Repository, com DTO: a responsabilidade de cada camada."},
		{2, "#9", "Variáveis de ambiente", "application.yml, profiles (dev/test/prod) e segredos fora do código."},
		{2, "#8", "Middleware", "No Spring: Filters e HandlerInterceptor, e onde entram no caminho da requisição."},
	}},
	{"Etapa 4 · Persistência, validação e erros", "Usar JPA sem cair no N+1, validar a entrada com DTO e devolver erros padronizados.", []seedTopic{
		{1, "#27", "Validação de dados", "DTO de entrada, Bean Validation (@Valid, @NotBlank, @Size, @Email), validação personalizada."},
		{1, "#19", "ORMs (JPA/Hibernate)", "Entidades, relacionamentos, Lazy x Eager, problema N+1, @Transactional, dirty checking."},
		{1, "#32", "Tratamento de erros", "@RestControllerAdvice, @ExceptionHandler e resposta de erro padronizada com o status certo."},
	}},
	{"Etapa 5 · Segurança", "Explicar autenticação x autorização e proteger endpoints com Spring Security e JWT.", []seedTopic{
		{1, "#20", "Autenticação e autorização", "A diferença entre as duas; sessão x token; a filter chain do Spring Security."},
		{1, "#23", "Hashing de senhas", "Nunca guardar senha em texto; BCrypt com PasswordEncoder."},
		{2, "#21", "JWT", "Estrutura (header.payload.assinatura), expiração e envio no header Authorization."},
		{2, "#28", "Controle de acesso", "Roles e authorities, regras por endpoint, @PreAuthorize."},
		{2, "#24", "CORS", "Por que o navegador bloqueia e como liberar origens no Spring (vai importar no Angular)."},
	}},
	{"Etapa 6 · Testes", "Testar um Service com Mockito e um endpoint com MockMvc.", []seedTopic{
		{1, "#34", "Testes unitários", "JUnit 5, assertions, @BeforeEach; Mockito: mock, when, verify."},
		{2, "#35", "Testes de integração", "@SpringBootTest, MockMvc e teste com banco; Testcontainers é diferencial."},
	}},
	{"Etapa 7 · Docker", "Subir a API e o PostgreSQL juntos com docker compose.", []seedTopic{
		{1, "#38", "Docker", "Imagem x container, Dockerfile, volumes, portas, variáveis e docker compose com PostgreSQL."},
	}},
	{"Etapa 8 · Arquitetura e boas práticas", "Justificar a organização do seu código usando SOLID.", []seedTopic{
		{1, "#31", "SOLID", "Os 5 princípios com exemplos em Java, principalmente o S e o D."},
		{2, "#33", "Logs e monitoramento", "SLF4J/Logback e níveis de log; Actuator e health check."},
		{3, "#30", "Clean Architecture", "Só os fundamentos, e só depois de dominar a arquitetura em camadas."},
	}},
	{"Etapa 9 · Diferenciais", "Estude quando as etapas anteriores estiverem firmes.", []seedTopic{
		{3, "#39", "CI/CD", "GitHub Actions rodando build e testes a cada push; noção de deploy e secrets."},
		{3, "#10", "Programação assíncrona", "Em Java: CompletableFuture e @Async; Virtual Threads só o conceito."},
	}},
	{"Depois · Fora do foco agora", "Não estude agora. Volte aqui quando a base de backend Java estiver firme.", []seedTopic{
		{4, "#13", "Bancos não relacionais", "Só saber o que são e quando se usa (MongoDB, Redis). Seu foco é PostgreSQL."},
		{4, "#22", "OAuth 2.0", "O conceito do 'entrar com Google'. A configuração fica para depois do JWT."},
		{4, "#25", "Rate limit", "O conceito e o status 429. A implementação fica para depois."},
	}},
}
