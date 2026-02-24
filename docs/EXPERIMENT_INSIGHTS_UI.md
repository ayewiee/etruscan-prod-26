## Experiment Insights UI — мини-док

_PARTIALLY OR FULLY AI GENERATED (but reviewed and refined by a human <3)_

---

Этот фронтенд — небольшой SPA на `Vue 3 + Vite`, который помогает быстро **просматривать эксперименты и их метрики** поверх REST‑API платформы Etruscan.  
Он поднимается вместе с backend через Docker Compose и доступен через общий nginx‑proxy.

---

## Как запустить UI

UI уже включён в `docker-compose.yml`:

- сервис `ui` билдится из `ui/experiment-insights`;
- сервис `proxy` (nginx) пробрасывает порт `80` наружу и проксирует:
  - API на `/api/v1/...`;
  - фронтенд UI на `/` (порт `5173` внутри сети).

**Запуск** (как и в runbook):

```bash
cd waw0905
docker compose up -d
```

После того как `/api/v1/ready` начинает отвечать `200`, UI доступен по адресу:

- `http://localhost/` — главная страница Experiment Insights.

---

## Аутентификация в UI

UI **не реализует полноценный логин**, он ожидает уже полученный JWT‑токен:

1. Получите token через backend:

```bash
curl -s -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@etruscan.com",
    "password": "admin"
  }'
```

2. Скопируйте значение поля `token`.
3. Откройте `http://localhost/` в браузере.
4. В правой части верхней панели вставьте токен в поле **“JWT token”** и нажмите кнопку **Apply**.

UI:

- сохраняет токен в `localStorage` (ключ `experiment_insights_jwt`);
- прокидывает его во все API‑запросы через `Authorization: Bearer <token>`.

---

## Основные экраны

### 1. Список экспериментов (`ExperimentListView`)

Компонент `src/views/ExperimentListView.vue`:

- Загружает список экспериментов через `GET /api/v1/experiments` (см. `src/api/experiments.ts`).
- Показывает таблицу с колонками:
  - **Name** (название эксперимента);
  - **Status** (текущий статус, `tag`);
  - **Audience** (доля аудитории, %);
  - **Primary metric** (ключ основной метрики);
  - **Outcome** (`ROLLOUT` / `ROLLBACK` / `NO_EFFECT`, если есть);
  - **Updated** (время последнего обновления).
- Поддерживает фильтры:
  - по **Status** (`DRAFT`, `ON_REVIEW`, `APPROVED`, `LAUNCHED`, ...);
  - по **Outcome** (`ROLLOUT`, `ROLLBACK`, `NO_EFFECT`).
- Состояние фильтров синхронизируется с query‑параметрами URL (`status`, `outcome`), чтобы ссылкой можно было поделиться.
- Клик по строке таблицы открывает страницу эксперимента `/experiments/:id`.

### 2. Детальная страница эксперимента (`ExperimentDetailView`)

Компонент `src/views/ExperimentDetailView.vue`:

- Загружает:
  - сам эксперимент (`GET /api/v1/experiments/:id`);
  - каталог метрик (`GET /api/v1/metrics`);
  - отчёт по эксперименту (`GET /api/v1/experiments/:id/report?from=...&to=...`).
- В шапке показывает:
  - название;
  - статус и исход (`Outcome`);
  - долю аудитории;
  - основную метрику (по каталогу метрик, если доступна).
- Слева:
  - блок **Metrics** с выбором метрики (radio‑кнопки по `metricKey`, primary помечен тегом `primary`);
  - таблица значений по вариантам: вариант, значение, вес, признак контрольного;
  - бар‑чарт (Chart.js через `vue-chartjs`) по выбранной метрике.
- Справа:
  - блок **Guardrails** с перечислением настроенных правил: `metricKey`, направление (`upper/lower`), порог, окно и действие (`pause/rollback`);
  - блок **Export & sharing**:
    - кнопка **Download CSV** — выгружает CSV с метриками по вариантам;
    - кнопка **Download chart PNG** — сохраняет текущий график.
- Вверху есть контролы `From/To` (datetime‑local):
  - UI пересчитывает отчёт через `/report` в выбранном окне;
  - параметры окна (`from`, `to`) и выбранная метрика (`metric`) кладутся в query‑параметры, так что ссылку можно шарить как “перманентный отчёт”.

---

## Как использовать UI на демо

Типичный сценарий:

1. Поднять стек (`docker compose up -d`) и убедиться, что `/api/v1/ready` отвечает `200`.
2. Вытянуть admin‑token и вставить его в поле “JWT token” в UI.
3. На бэкенде (через `curl` или Postman) создать эксперимент и собрать события (`decide` + `track`), как описано в `docs/RUNBOOK.md`.
4. В UI:
   - на списке экспериментов найти нужный эксперимент (по имени/статусу);
   - открыть детальную страницу, выбрать нужное окно времени и метрику;
   - показать таблицу и график по вариантам;
   - продемонстрировать, как отображаются guardrails и как выглядит CSV/PNG‑экспорт.

UI не покрывает весь функционал админки (создание флагов, редактирование экспериментов и т.п.) — он сфокусирован именно на **инсайтах по экспериментам** (C8 из задания) и помогает быстрее пройти блоки `B6` и допфичу “Experiment Insights UI”.

