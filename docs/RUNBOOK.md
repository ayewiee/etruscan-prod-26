## Runbook запуска и проверки платформы Etruscan

_PARTIALLY OR FULLY AI GENERATED (but reviewed and refined by a human <3)_

---

## 1. Предусловия

- **ОС**: Linux / macOS / Windows с поддержкой Docker.
- **Docker**: установлен и запущен Docker Engine.
- **Docker Compose**: поддержка команды `docker compose` (v2).
- **Порт 80** свободен на хосте (его занимает nginx‑proxy из `docker-compose.yml`).

Никаких дополнительных ручных шагов (миграции, создание пользователей и т.п.) делать не нужно — они выполняются внутри контейнера API при старте.

---

## 2. Запуск системы и проверки readiness / liveness (B1, B9)

1. **Клонировать репозиторий и перейти в корень**:

```bash
git clone <этот_репозиторий>
cd waw0905
```

2. **Запустить весь стек через Docker Compose**:

```bash
docker compose up -d
```

3. **Дождаться готовности API** (через nginx‑proxy на `localhost:80`):

```bash
curl http://localhost/api/v1/ready
```

Ожидаемый результат:

- HTTP‑код: `200`.
- Тело:

```json
{
  "status": "ready"
}
```

Пока сервис не готов, будет ошибка подключения или не‑`200`.

4. **Проверка живости**:

```bash
curl http://localhost/api/v1/health
```

Ожидаемый результат:

- HTTP‑код: `200`.
- Тело:

```json
{
  "status": "healthy"
}
```

Эти два эндпоинта подтверждают критерии `B9‑1` (readiness) и `B9‑2` (liveness).

---

## 3. Быстрое получение admin‑токена (аутентификация)

Для всех админских API нужен JWT‑токен пользователя с ролью `ADMIN`. Seed‑данные создают пользователя:

- email: `admin@etruscan.com`
- пароль: `admin`

1. **Определить базовый HOST**:

```bash
export HOST=http://localhost
```

2. **Логин admin**:

```bash
curl -s -X POST "$HOST/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@etruscan.com",
    "password": "admin"
  }'
```

Ожидаемый результат (HTTP 200):

```json
{
  "token": "eyJhbGci...",
  "user": {
    "id": "00000000-0000-0000-0000-000000000001",
    "email": "admin@etruscan.com",
    "username": "admin",
    "role": "ADMIN",
     ...
  }
}
```

3. **Сохранить токен в переменную окружения**:

```bash
export TOKEN="eyJhbGci..."   # подставьте значение поля token из ответа
```

Все защищённые запросы далее используют заголовок:

```bash
-H "Authorization: Bearer $TOKEN"
```

---

## 4. Happy‑path: `purchase_button_text` → `/decide` → `/track` → отчёт

Этот сценарий показывает основной поток:

- `B1` – система поднята и работает;
- `B2` – `default/variant`, таргетинг, детерминизм и веса;
- `B3` – жизненный цикл эксперимента и ревью;
- `B4` – приём, валидация и атрибуция событий;
- `B6` – отчёт и фиксация исхода эксперимента.

### 4.1 Создать Approver‑пользователя и группу ревьюеров (B3‑1..B3‑5)

1. **Создать пользователя‑аппрувера** (под admin‑токеном):

```bash
curl -s -X POST "$HOST/api/v1/admin/users" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "email": "approver@etruscan.com",
    "password": "P4SSW0RD",
    "username": "approver",
    "role": "APPROVER"
  }'
```

Сохраните `id` пользователя как:

```bash
export APPROVER_ID="<uuid из поля user.id>"
```

2. **Создать approver‑группу и добавить участника**:

```bash
curl -s -X POST "$HOST/api/v1/admin/approverGroups" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "First approver group"
  }'
```

Ответ содержит `id` группы:

```bash
export APPROVER_GROUP_ID="<uuid группы>"
```

Добавляем аппрувера в группу:

```bash
curl -s -X POST "$HOST/api/v1/admin/approverGroups/$APPROVER_GROUP_ID/members/add" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "users": ["$APPROVER_ID"]
  }'
```

### 4.2 Создать feature flag `purchase_button_text`

```bash
curl -s -X POST "$HOST/api/v1/flags" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "key": "purchase_button_text",
    "valueType": "string",
    "defaultValue": "Purchase"
  }'
```

Ожидаем HTTP 201 и объект флага. Сохраните идентификатор:

```bash
export FLAG_ID="<uuid из поля id>"
```

### 4.3 Создать эксперимент на этом флаге (B2, B3)

```bash
curl -s -X POST "$HOST/api/v1/experiments" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Test purchase button text",
    "flagId": "$FLAG_ID",
    "audiencePercentage": 100,
    "metricKeys": ["click_count"],
    "primaryMetricKey": "click_count",
    "targetingRule": "region != 'US' AND (version >= '1.2.3' OR premium = 'true')",
    "variants": [
      {
        "name": "Big text: PURCHASE",
        "value": "PURCHASE",
        "weight": 25
      },
      {
        "name": "Medium: Purchase",
        "value": "Purchase",
        "weight": 25,
        "isControl": true
      },
      {
        "name": "Small: purchase",
        "value": "purchase",
        "weight": 50
      }
    ]
  }'
```

Ожидаем HTTP 201 и статус `DRAFT`. Сохраняем идентификатор:

```bash
export EXP_ID="<uuid эксперимента>"
```

### 4.4 Ревью и запуск эксперимента (B3‑1..B3‑4)

1. **Отправить эксперимент на ревью**:

```bash
curl -s -X POST "$HOST/api/v1/experiments/$EXP_ID/sendOnReview" \
  -H "Authorization: Bearer $TOKEN"
```

Статус становится `ON_REVIEW`.

2. **Одобрить как ADMIN**:

```bash
curl -s -X POST "$HOST/api/v1/experiments/$EXP_ID/approve" \
  -H "Authorization: Bearer $TOKEN"
```

3. **Залогиниться как Approver и одобрить второй раз**:

```bash
curl -s -X POST "$HOST/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "approver@etruscan.com",
    "password": "P4SSW0RD"
  }'
```

Скопируйте поле `token` и временно замените переменную:

```bash
export TOKEN="<approver JWT>"
```

Повторно вызовите approve:

```bash
curl -s -X POST "$HOST/api/v1/experiments/$EXP_ID/approve" \
  -H "Authorization: Bearer $TOKEN"
```

После достижения порога одобрений (`DEFAULT_MIN_APPROVALS=2`) статус становится `APPROVED`.

4. **Вернуться к admin и запустить эксперимент**:

```bash
curl -s -X POST "$HOST/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@etruscan.com",
    "password": "admin"
  }'
export TOKEN="<admin JWT>"

curl -s -X POST "$HOST/api/v1/experiments/$EXP_ID/launch" \
  -H "Authorization: Bearer $TOKEN"
```

Статус → `LAUNCHED`. Эксперимент участвует в выдаче значений флага.

### 4.5 Решение во время показа (`POST /api/v1/decide`, B2‑1..B2‑4)

```bash
curl -s -X POST "$HOST/api/v1/decide" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "client123",
    "flagKey": "purchase_button_text",
    "context": {
      "region": "RU",
      "version": "1.3",
      "premium": true
    }
  }'
```

Ожидаемый ответ:

```json
{
  "decisionId": "....",
  "value": "PURCHASE"
}
```

Возможные случаи:

- если нет активного эксперимента или пользователь не попадает в таргетинг → `value = "Purchase"` (default);
- если эксперимент применим → `value` равно одному из вариантов;
- повторные вызовы для того же `userId` и неизменной конфигурации возвращают то же значение (детерминизм).

Сохраните идентификатор решения:

```bash
export DECISION_ID="..."
```

### 4.6 Отправка событий (`POST /api/v1/track`, B4‑1..B4‑5)

```bash
curl -s -X POST "$HOST/api/v1/track" \
  -H "Content-Type: application/json" \
  -d '{
    "events": [
      {
        "userId": "client123",
        "eventId": "event-exposure-1",
        "eventTypeKey": "exposure",
        "decisionId": "$DECISION_ID",
        "timestamp": "2026-01-01T00:00:00Z"
      },
      {
        "userId": "client123",
        "eventId": "event-click-1",
        "eventTypeKey": "click",
        "decisionId": "$DECISION_ID",
        "timestamp": "2026-01-01T00:00:05Z"
      }
    ]
  }'
```

Типы событий `exposure` и `click_count` создаются заранее через `POST /api/v1/events/types` или поднимаются миграциями (см. `docs/METRICS_AND_GUARDRAILS.md`).

Ожидаемый ответ:

```json
{
  "accepted": 2,
  "duplicates": 0,
  "rejected": 0,
  "errors": []
}
```

- чтобы показать валидацию типов и обязательных полей (B4‑1, B4‑2), отправьте событие без `eventTypeKey` или с неверным типом поля – оно попадёт в `rejected`;
- чтобы показать дедупликацию (B4‑3), повторно отправьте событие с тем же `eventId` – оно попадёт в `duplicates`.

### 4.7 Отчёт и фиксация исхода (B6‑1..B6‑5)

1. **Построить отчёт по эксперименту**:

```bash
curl -s -X GET "$HOST/api/v1/experiments/$EXP_ID/report?from=2026-01-01T00:00:00Z&to=2026-02-24T00:00:00Z" \
  -H "Authorization: Bearer $TOKEN"
```

Ожидаем:

- HTTP 200;
- в ответе — показатели по каждому варианту и выбранным метрикам за указанный период (подтверждает B6‑1, B6‑2, B6‑3).

2. **Зафиксировать исход эксперимента**:

```bash
curl -s -X POST "$HOST/api/v1/experiments/$EXP_ID/finish" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "outcome": "ROLLOUT",
    "comment": "Variant with better click_count and no guardrail degradation."
  }'
```

Статус становится `FINISHED`/`ARCHIVED`, а исход (`ROLLOUT`/`ROLLBACK`/`NO_EFFECT`) и комментарий сохраняются в истории (B6‑4, B6‑5).

---

## 5. Guardrails и уведомления (B5, C7)

Минимальный сценарий для демонстрации safety‑механик и нотификаций:

1. Создать подходящую метрику (например, `error_rate`) по инструкции из `docs/METRICS_AND_GUARDRAILS.md`.
2. При создании или обновлении эксперимента добавить в тело поле `guardrails` с элементом:

   - `metricKey`: `"error_rate"`;
   - `threshold`: например, `0.05`;
   - `thresholdDirection`: `"upper"`;
   - `windowSeconds`: например, `300`;
   - `action`: `"pause"` или `"rollback"`.

3. Сгенерировать пачку событий с высоким числом `error` для decisionId этого эксперимента.
4. Подождать не менее одного интервала `GUARDRAIL_CHECK_INTERVAL_MINUTES` (см. `docker-compose.yml`).
5. Проверить:

   - статус эксперимента стал `PAUSED` или `FINISHED` с исходом `ROLLBACK` (B5‑3, B5‑4);
   - в истории guardrail‑срабатываний есть запись с метрикой, порогом, окном, действием и фактическим значением (B5‑1, B5‑2, B5‑5);
   - при `NOTIFICATIONS_ENABLED=true` и настроенных получателях пришло уведомление (Telegram виден как реальный канал, Email логируется).

---

## 6. Ограничение участия пользователя в экспериментах (B5‑6)

Чтобы показать, что пользователь не превращается в «вечного подопытного»:

1. Убедиться, что сервис запущен вместе с Redis (это делает `docker compose up`) и включён `ParticipationTracker` (по умолчанию лимит — не более 2 одновременных экспериментов на пользователя).
2. Запустить минимум **3** эксперимента на разных флагах.
3. С одним и тем же `userId` вызвать `/api/v1/decide` по каждому флагу:

   - для первых двух флагов пользователь будет получать экспериментальные значения;
   - начиная с третьего — только `defaultValue`, участие не регистрируется.

4. В Redis по ключу `user_active_exps:<userId>` виден список активных экспериментов и TTL «отдыха» (cooldown‑период).

---

## 7. Как использовать этот runbook на демо

Этот runbook работает в связке с:

- `docs/TESTING_REPORT.md` – формальный отчёт по тестированию (B8);
- `docs/TRACEABILITY_MATRIX.md` – матрица «задание–критерий–реализация»;
- `docs/ARCHITECTURE.md` и ADR – архитектурный контекст и список ключевых решений (B7‑4..B7‑9);
- `docs/OPERATIONS.md` – детали по метрикам, логам и эксплуатационной модели (B9).

