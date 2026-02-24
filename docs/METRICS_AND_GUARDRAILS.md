# Метрики и Guardrails

_PARTIALLY OR FULLY AI GENERATED (but reviewed and refined by a human <3)_

---

## Обзор

- **Метрики (Metrics)** описывают, **что именно мы измеряем**: это сущности каталога с ключом, именем, типом, привязанным типом события и способом агрегации. Метрику создают один раз в каталоге, затем привязывают к экспериментам.
- **Guardrails** — это **правила безопасности на уровне эксперимента**: “если метрика X вышла за порог в окне времени — поставь эксперимент на паузу или откати”. Guardrails всегда используют метрики из каталога.

---

## Метрики

### Что такое метрика?

**Метрика** описывает, как посчитать одно числовое значение. Есть два вида:

1. **Примитивная метрика** — считается напрямую по событиям: один тип события + одна агрегация (`count`, `sum`, `avg`, `p95`).
2. **Производная (ratio) метрика** — считается как **метрика‑числитель / метрика‑знаменатель** (например, conversion rate = `conversion_count / exposure_count`).

Общие поля:

- **key** — уникальный идентификатор (например, `signup_completed`, `conversion_rate`);
- **name** — человекочитаемое название;
- **type** — `binomial` (счёт событий) или `continuous` (сумма/среднее/перцентиль или отношение);
- **isGuardrail** — подсказка для UI/конвенций, что метрика часто используется как guardrail.

Для **примитивных** метрик также задаются:

- **eventTypeKey** — по каким событиям (из `POST /api/v1/track`) считается метрика;
- **aggregationType** — как объединять события в одно число: `count`, `sum`, `avg`, `p95`.

Для **производных** метрик задаются:

- **numeratorMetricKey** — ключ метрики‑числителя;
- **denominatorMetricKey** — ключ метрики‑знаменателя.

Платформа гарантирует отсутствие циклов (A/B и B/A) и существование обеих базовых метрик. Деление на ноль трактуется как значение `0` (без `NaN`/`Inf`).

### Пример создания примитивной метрики

`POST /api/v1/metrics` (авторизованный запрос):

```json
{
  "key": "purchase_count",
  "name": "Number of purchases",
  "description": "Count of purchase events per user",
  "type": "continuous",
  "eventTypeKey": "purchase",
  "aggregationType": "sum",
  "isGuardrail": false
}
```

### Пример создания производной (ratio) метрики

В этом случае **не** указываются `eventTypeKey` и `aggregationType`; вместо них задаются `numeratorMetricKey` и `denominatorMetricKey`:

```json
{
  "key": "conversion_rate",
  "name": "Conversion rate",
  "description": "Conversions / exposures",
  "type": "continuous",
  "numeratorMetricKey": "conversion_count",
  "denominatorMetricKey": "exposure_count",
  "isGuardrail": false
}
```

Типичные примеры:

- **conversion rate** = `conversion_count / exposure_count`;
- **error rate** = `error_count / exposure_count`;
- **click-through rate** = `click_count / exposure_count`.

События приходят через `POST /api/v1/track` с полем `eventTypeKey` (и дополнительными `properties`).  
Компонент **MetricComputer** (используется в отчётах и GuardrailRunner) по окну времени:

- загружает решения эксперимента;
- загружает события нужных типов;
- для примитивных метрик применяет агрегацию;
- для ratio‑метрик считает числитель и знаменатель и возвращает их отношение.

### Зачем метрике поле `isGuardrail`?

Поле **`isGuardrail`** — это **подсказка уровня каталога**, что метрика чаще всего используется как защитная (safety) метрика, а не как основная целевая:

- не меняет способ вычисления метрики;
- не создаёт автоматически никаких правил guardrails;
- служит для:
  - **UI**: отдельный список “Guardrail metrics” при настройке эксперимента;
  - **конвенций**: пометить `error_rate`, `crash_rate` и подобные как хорошие кандидаты для guardrails.

Сами **правила guardrails** задаются на уровне эксперимента (порог, направление, окно, действие):  
сначала создаётся метрика в каталоге (при желании с `isGuardrail: true`), затем при создании/редактировании эксперимента в теле добавляется массив `guardrails`, который ссылается на эту метрику по `metricKey` и задаёт `threshold`, `windowSeconds`, `action`.

---

## Guardrails

### Что такое guardrail?

**Guardrail** — это правило на уровне эксперимента:

- **metricKey** — ключ одной из метрик, доступных эксперименту;
- **threshold** — числовой порог;
- **thresholdDirection** — направление: `upper` (триггер, если значение > порога) или `lower` (если < порога);
- **windowSeconds** — окно наблюдения (например, последние 300 секунд);
- **action** — что делать при срабатывании: `pause` (поставить на паузу) или `rollback`.

### Как задать guardrails

Отдельного API “создать guardrail” нет — guardrails задаются **внутри конфигурации эксперимента**:

1. Сначала создаются нужные метрики (`POST /api/v1/metrics`), которые планируется использовать в guardrails (например, `error_rate`).
2. При создании или обновлении эксперимента (`POST /api/v1/experiments` или `PUT /api/v1/experiments/:id`) в теле добавляется массив `guardrails`:

```json
{
  "flagId": "...",
  "name": "My experiment",
  "audiencePercentage": 50,
  "variants": [ ... ],
  "metricKeys": ["primary_metric"],
  "primaryMetricKey": "primary_metric",
  "guardrails": [
    {
      "metricKey": "error_rate",
      "threshold": 0.05,
      "thresholdDirection": "upper",
      "windowSeconds": 300,
      "action": "pause"
    }
  ]
}
```

Каждый элемент ссылается на **ключ метрики** из каталога. Бэкенд разрешает ключ в ID метрики и сохраняет правило guardrail для этого эксперимента.

### Как GuardrailRunner оценивает правила

Фоновый процесс **GuardrailRunner** периодически запускается (интервал задаётся в конфиге, по умолчанию несколько минут):

1. Находит все эксперименты в статусе **LAUNCHED**.
2. Для каждого эксперимента загружает его **guardrails**.
3. Для каждого guardrail:
   - загружает метрику из каталога;
   - считает значение метрики для данного эксперимента за последние `windowSeconds` (по той же схеме, что отчёты: решения в окне → события → агрегация);
   - сравнивает значение с порогом с учётом направления (`upper`/`lower`);
   - если условие выполнено:
     - обновляет статус эксперимента (`PAUSED` или завершает с `ROLLBACK` — в зависимости от политики);
     - записывает **факт срабатывания guardrail** в историю (для аудита);
     - пишет структурированный лог.

Итого:

- **метрики** описывают *что* считать;
- **guardrails** описывают *когда и как реагировать* на деградацию (порог + окно + действие) и всегда привязаны к конкретному эксперименту.

---

## Краткое резюме

| Концепция           | Где живёт              | Назначение |
|---------------------|------------------------|------------|
| **Metric**          | Каталог метрик         | Описывает измерение: примитивное (eventTypeKey + aggregation) или ratio (numeratorMetricKey / denominatorMetricKey). Создаётся через `POST /metrics`. |
| **isGuardrail**     | Поле на Metric         | Подсказка, что метрика обычно используется как guardrail (влияет на UI/конвенции, но не на вычисление). |
| **Guardrails**      | Конфигурация эксперимента | Правила “если метрика X вышла за порог в окне N секунд — поставь на паузу/откатись”. Настраиваются при создании/обновлении эксперимента. |
| **GuardrailRunner** | Фоновый процесс        | Периодически пересчитывает guardrails для `LAUNCHED`‑экспериментов, меняет статус и записывает триггеры. |

Чтобы увидеть полную архитектуру и эксплуатационное поведение guardrails (как они связаны с отчётами, логированием и критериями `B5` и `B9`), см. также:

- `docs/ARCHITECTURE.md` — разделы про GuardrailRunner и критический путь `decide → event → report/guardrail`;
- `docs/OPERATIONS.md` — секции про триггеры guardrails и эксплуатационную модель;
- ADR‑003 в `docs/adr` — обоснование выбора фонового runner‑а поверх каталога метрик.

# Metrics and Guardrails

## Overview

- **Metrics** define what you measure: they are catalog entities (key, name, type, event type, aggregation). You create them once, then attach them to experiments.
- **Guardrails** are experiment-level rules: “if metric X goes above/below a threshold over a time window, pause or rollback the experiment.” They use metrics from the catalog.

---

## Metrics

### What is a metric?

A **metric** describes how to compute a single number. There are two kinds:

1. **Primitive metric** – computed from events: one event type + one aggregation (`count`, `sum`, `avg`, `p95`).
2. **Derived (ratio) metric** – computed as **numerator metric / denominator metric** (e.g. conversion rate = conversion_count / exposure_count).

Common fields:

- **Key** – unique identifier (e.g. `signup_completed`, `conversion_rate`).
- **Name** – human-readable label.
- **Type** – `binomial` (count of events) or `continuous` (sum/avg/p95 or a ratio).
- **IsGuardrail** – hint for UI/conventions (see below).

For **primitive** metrics you also set:

- **EventTypeKey** – which ingested events (from `POST /track
`) this metric is based on.
- **AggregationType** – how to turn events into one number: `count`, `sum`, `avg`, `p95`.

For **derived** metrics you set:

- **NumeratorMetricKey** – key of the metric used as the numerator.
- **DenominatorMetricKey** – key of the metric used as the denominator.

The platform ensures no cycles (e.g. A/B and B/A) and that numerator/denominator metrics exist. Zero denominator is treated as 0 (no NaN/Inf).

### How to create a primitive metric

**POST /api/v1/metrics** (authenticated):

```json
{
  "key": "purchase_count",
  "name": "Number of purchases",
  "description": "Count of purchase events per user",
  "type": "continuous",
  "eventTypeKey": "purchase",
  "aggregationType": "sum",
  "isGuardrail": false
}
```

### How to create a derived (ratio) metric

Omit `eventTypeKey` and `aggregationType`; set `numeratorMetricKey` and `denominatorMetricKey`:

```json
{
  "key": "conversion_rate",
  "name": "Conversion rate",
  "description": "Conversions / exposures",
  "type": "continuous",
  "numeratorMetricKey": "conversion_count",
  "denominatorMetricKey": "exposure_count",
  "isGuardrail": false
}
```

Examples: **conversion rate** = conversion_count / exposure_count, **error rate** = error_count / exposure_count, **click-through rate** = click_count / exposure_count.

Events are ingested via **POST /api/v1/track
** with `eventTypeKey` (and optional `properties`). The **MetricComputer** (used by reports and the guardrail runner) loads decisions for the experiment in a time window; for primitive metrics it loads events and applies the aggregation; for derived metrics it computes numerator and denominator metrics and returns their ratio.

### Why does a metric have `isGuardrail`?

**`isGuardrail`** is a **catalog-level hint**: “this metric is typically used as a guardrail (safety check), not as a primary success metric.”

- It does **not** change how the metric is computed.
- It does **not** automatically create a guardrail rule.
- Use it for:
  - **UI**: e.g. show these metrics in a “Guardrail metrics” list when configuring an experiment.
  - **Conventions**: mark metrics like error rate, crash rate, so teams know they’re good candidates for guardrail rules.

Guardrail **rules** are defined per experiment (threshold, direction, window, action). So: you create a **metric** in the catalog (optionally with `isGuardrail: true`), then when creating/editing an **experiment** you add **guardrails** that reference that metric and set threshold/window/action.

---

## Guardrails

### What is a guardrail?

A **guardrail** is an experiment-level rule:

- **Metric** – one of the metrics (by key) attached or available to the experiment.
- **Threshold** – numeric bound.
- **ThresholdDirection** – `upper` (trigger when value > threshold) or `lower` (trigger when value < threshold).
- **WindowSeconds** – time window to compute the metric over (e.g. last 5 minutes).
- **Action** – what to do when the rule fires: `pause` (set experiment to PAUSED) or `rollback`.

### How to create guardrails

You **don’t** create guardrails via a standalone “create guardrail” API. They are part of an **experiment**:

1. **Create metrics** (POST /api/v1/metrics) that you’ll use as guardrails (e.g. error rate, revenue drop).
2. **Create or update an experiment** (POST /api/v1/experiments or PUT /api/v1/experiments/:id) and in the body include a **guardrails** array:

```json
{
  "flagId": "...",
  "name": "My experiment",
  "audiencePercentage": 50,
  "variants": [ ... ],
  "metricKeys": [ "primary_metric" ],
  "primaryMetricKey": "primary_metric",
  "guardrails": [
    {
      "metricKey": "error_rate",
      "threshold": 0.05,
      "thresholdDirection": "upper",
      "windowSeconds": 300,
      "action": "pause"
    }
  ]
}
```

Each item references a **metric key** from the catalog and defines the rule. The backend resolves the key to the metric ID and stores the guardrail for that experiment.

### How guardrails are evaluated

The **GuardrailRunner** (background process) runs periodically (e.g. every 2 minutes):

1. Lists experiments in status **LAUNCHED**.
2. For each experiment, loads its **guardrails**.
3. For each guardrail:
   - Loads the **metric** (from catalog).
   - Computes the metric value for the experiment over the last **windowSeconds** (same logic as reports: decisions in that window → events → aggregation).
   - Compares the value to the **threshold** (upper/lower).
   - If the rule fires:
     - Updates experiment status to **PAUSED**.
     - Records a **guardrail trigger** (for auditing).
     - Logs the event.

So: **metrics** define *what* to compute; **guardrails** define *when* to react (threshold + window + action) and are always tied to an experiment.

---

## Summary

| Concept        | Where it lives      | Purpose |
|----------------|---------------------|--------|
| **Metric**     | Catalog (metrics)   | Define a measurable: primitive (eventTypeKey + aggregation) or derived (numeratorMetricKey / denominatorMetricKey). Create via POST /metrics. |
| **isGuardrail**| On Metric           | Hint that this metric is often used for safety guardrails (UI/conventions only). |
| **Guardrails** | On Experiment       | Rules: “if metric X &lt; or &gt; threshold in last N seconds, pause/rollback.” Set in experiment create/update. |
| **GuardrailRunner** | Background      | Periodically evaluates guardrails for LAUNCHED experiments and pauses + records triggers. |

For the full architecture and operational behavior of guardrails (how they interact with reports, what is logged and how this maps to criteria B5 and B9), see also:

- `docs/ARCHITECTURE.md` – разделы про GuardrailRunner и критический путь `decide → event → report/guardrail`;
- `docs/OPERATIONS.md` – секции про guardrail‑триггеры и эксплуатационную модель;
- ADR‑003 в `docs/adr` – архитектурное обоснование выбора фонового runner‑а поверх каталога метрик.
