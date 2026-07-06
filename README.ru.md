# CS2 Demo Highlighter

> **Notice**
> Этот репозиторий публичный, но по сути это домашний персональный проект.
> Интерфейсы, детали формата вывода и поведение записи могут меняться.
> Возможны баги и неочевидные edge cases.
> Используйте с осторожностью и проверяйте результаты перед применением в production.

English version: [README.md](./README.md)

CS2 Demo Highlighter это CLI-инструмент, который парсит `.dem` файлы, извлекает хайлайт-события для конкретного игрока и генерирует HLAE-скрипты для автоматической записи.

Проект сфокусирован на извлечении событий и оркестрации записи, а не на полном постпродакшн-пайплайне.

## Возможности

- Парсинг демо через `github.com/markus-wa/demoinfocs-golang/v5`
- Детекция хайлайт-событий:
  - `kill_in_smoke`
  - `kill_blinded`
  - `wallbang`
  - `noscope`
  - `headshot_kill`
  - `round_multikill`
  - `clutch_win`
- Фильтрация типов хайлайтов (`--types`)
- Гибкие render-таргеты: любой набор типов как **клипы** (отдельная запись на сегмент) или **монтаж** (одна непрерывная запись с jump cut)
- Генерация HLAE-скриптов на базе `mirv_streams` (без `startmovie`)
- POV lock через `spec_player <slot>`
- Расширение сегментов через pre-roll и post-roll
- Автопрыжки между сегментами (`demo_pause -> demo_gototick -> demo_resume`)
- Опциональные прыжки внутри `round_multikill` при больших паузах между киллами

## Выходные Артефакты

- `highlights.json`: нормализованные метаданные хайлайтов
- По одному `.cfg` на каждый render-таргет (см. [Render-таргеты](#render-таргеты)). По умолчанию — один clips-скрипт со всеми типами.

## Требования

- Go `1.26+`
- Валидный непустой CS2 `.dem` файл
- SteamID64 целевого игрока (17 цифр)
- Настроенный HLAE для записи CS2 (AfxHookSource2)

## Установка

```bash
git clone https://github.com/eSheikh/cs2-demo-highlighter.git
cd cs2-demo-highlighter
go mod download
```

## Быстрый Старт

```bash
go run ./cmd/highlighter \
  --demo /path/to/match.dem \
  --steamid 7656119XXXXXXXXXX \
  --out highlights.json \
  --clips highlights.cfg
```

Запуск тестов:

```bash
go test ./...
```

## Render-таргеты

Вывод записи настраивается повторяемыми флагами `--clips` и `--montage`. Каждый флаг создаёт один `.cfg` и имеет вид:

```
[types=]path.cfg
```

- `types` — типы хайлайтов через запятую (опустите или используйте `all` для всех типов). Значение делится по **первому** `=`, поэтому Windows-пути с буквой диска (`C:\...`) не ломаются.
- `path.cfg` — путь к выходному скрипту. Его базовое имя также идёт в конец `mirv_streams record name`, поэтому разные таргеты пишутся в разные папки.

Если ни один флаг не задан, по умолчанию создаётся один clips-таргет со всеми типами в `highlights.cfg`.

Примеры:

```bash
# Клипы всех хайлайтов (дефолт, явно)
go run ./cmd/highlighter ... --clips highlights.cfg

# Разные выводы за один прогон: клипы клатчей + монтаж хедшотов
go run ./cmd/highlighter ... \
  --clips clutch_win,wallbang=clutches.cfg \
  --montage headshot_kill=headshots.cfg

# Монтаж всех смок-килов и отдельно всех ноускопов
go run ./cmd/highlighter ... \
  --montage kill_in_smoke=smokes.cfg \
  --montage noscope=noscopes.cfg
```

## CLI Параметры

| Flag              | По умолчанию         | Описание                                                                          |
| ----------------- | -------------------- | --------------------------------------------------------------------------------- |
| `--demo`          | -                    | Путь к входному `.dem` файлу (обязательно)                                        |
| `--steamid`       | -                    | Целевой SteamID64 (обязательно, 17 цифр)                                          |
| `--out`           | `highlights.json`    | Путь к выходному JSON (пустое значение отключает JSON)                            |
| `--types`         | (все)                | Типы хайлайтов через запятую, оставляемые в результате (пусто/`all` = все)        |
| `--clips`         | `highlights.cfg`     | Clips render-таргет `[types=]path.cfg` (повторяемый)                              |
| `--montage`       | -                    | Montage render-таргет `[types=]path.cfg` (повторяемый)                            |
| `--hlae-path`     | текущая директория   | Директория для `mirv_streams record name`                                        |
| `--hlae-preset`   | `afxFfmpegYuv420p`   | HLAE FFmpeg preset                                                                |
| `--hlae-fps`      | `60`                 | FPS записи                                                                        |
| `--hlae-preroll`  | `3`                  | Секунды до события                                                                |
| `--hlae-postroll` | `2`                  | Секунды после события                                                             |
| `--hlae-kill-gap` | `10`                 | Секунды между киллами в `round_multikill` для прыжка внутри записи (`0` отключает) |

Отключить JSON-вывод:

```bash
go run ./cmd/highlighter ... --out ""
```

## Сценарии Записи

### Клипы (отдельная запись на сегмент)

1. Запустите CS2 через HLAE.
2. Загрузите демо: `playdemo <demo_name>`.
3. Вставьте clips-`.cfg` в консоль HLAE.
4. Дождитесь `All N segments recorded`.
5. Скрипт завершится `disconnect` и вернет в главное меню.

Результат: несколько выходных файлов, по одному на сегмент.

### Монтаж (одна непрерывная запись)

1. Загрузите то же демо.
2. Вставьте montage-`.cfg`.
3. Запись стартует один раз, прыгает между выбранными сегментами и завершается один раз.
4. Скрипт завершится `disconnect` и вернет в главное меню.

Результат: один монтажный выходной файл.

## Примеры Сгенерированных Файлов

### `highlights.json`

Раунды 1-based (раунд 1 — первый раунд).

```json
{
  "demo": "mirage1.dem",
  "steamid": "7656119XXXXXXXXXX",
  "tick_rate": 64,
  "highlights": [
    {
      "type": "round_multikill",
      "round": 5,
      "tick_start": 34217,
      "tick_end": 34714,
      "kills": 2,
      "kill_ticks": [34217, 34714],
      "victims": ["7656119XXXXXXXXXX", "7656119XXXXXXXXXX"],
      "weapon": "AWP",
      "player_slot": 9,
      "steamid": "7656119XXXXXXXXXX",
      "demo": "mirage1.dem",
      "segment_tick_start": 34217,
      "segment_tick_end": 34714
    }
  ]
}
```

### Clips `.cfg` (сокращённо)

Setup-блок пишется один раз, далее идут построчные `mirv_cmd addAtTick`. Вывод без комментариев — консоль CS2/HLAE может ломаться на строках-комментариях.

```cfg
mirv_cvar_unhide_all;
mirv_cmd clear;
mirv_streams record end;
mirv_streams record name "<hlae-path>/<steamid>/<date>/<target>";
mirv_streams settings edit afxDefault settings afxFfmpegYuv420p;
mirv_streams record fps 60;
...

mirv_cmd addAtTick 34025 "spec_player 9; host_framerate 60; mirv_streams record start";
mirv_cmd addAtTick 34817 "mirv_streams record end; host_framerate 0";
mirv_cmd addAtTick 34818 "demo_pause; demo_gototick 40130; spec_player 9; demo_resume";
```

## Валидация и Обработка Ошибок

- Fail-fast валидация конфигурации до старта парсинга:
  - пустой путь к демо
  - неверное расширение (не `.dem`)
  - отсутствующий / не обычный / пустой файл
  - некорректный формат SteamID64
  - неизвестный тип хайлайта в `--types` / render-таргетах
  - лидирующие/хвостовые пробелы в строковых CLI-флагах автоматически обрезаются
- Защитное поведение парсера:
  - дополнительная проверка пути
  - конвертация parser panic в обычную ошибку
  - явная ошибка для битого/обрезанного демо
  - поддержка отмены через `context`

## Архитектура

- `cmd/highlighter`: CLI entrypoint
- `internal/bootstrap`: разбор флагов и запуск CLI (запись файлов здесь)
- `internal/engine`: ядро без I/O — список игроков, парсинг + извлечение хайлайтов, стрим прогресса
- `internal/parser`: извлечение событий из демо (`demoinfocs`)
- `internal/service`: правила хайлайтов и доменная логика
- `internal/hlae`: render-таргеты, планирование сегментов, рендеринг скриптов
- `internal/repository`: слой сохранения данных
- `internal/model`: общие типы

## Ограничения

- Качество вывода зависит от целостности демо и качества parser-событий.
- Определение clutch основано на правилах, а не на model/vision подходе.
- Монтаж это jump-cut автоматизация в демо-плеере, а не NLE постпродакшн.

## Roadmap

1. Новые типы хайлайтов (`awp_flick`, `360` и т.д.).
2. Интерактивный TUI (выбор демо/игрока, живой прогресс, выбор хайлайтов).
3. Автозапуск записи через HLAE (`recorder`).
4. Добавление звука в записанные видео с хайлайтами.

## License

Проект распространяется по лицензии MIT. См. [LICENSE](./LICENSE).
