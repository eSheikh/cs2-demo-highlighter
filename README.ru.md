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
  - `round_multikill`
  - `clutch_win`
  - `headshot_kill`
  - `headshot_collection` (summary-событие)
- Генерация HLAE-скриптов на базе `mirv_streams` (без `startmovie`)
- POV lock через `spec_player <slot>`
- Расширение сегментов через pre-roll и post-roll
- Автопрыжки между сегментами (`demo_pause -> demo_gototick -> demo_resume`)
- Опциональные прыжки внутри `round_multikill` при больших паузах между киллами

## Выходные Артефакты

Инструмент может генерировать:

- `highlights.json`: нормализованные метаданные хайлайтов
- `highlights.cfg`: HLAE-скрипт для обычных хайлайт-клипов
- `headshots.cfg`: HLAE-скрипт для one-file headshot-монтажа с jump cut

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
  --hlae highlights.cfg \
  --hlae-headshots headshots.cfg \
  --hlae-path highlights \
  --hlae-preset afxFfmpegYuv420p
```

Запуск тестов:

```bash
go test ./...
```

## CLI Параметры

| Flag                    | По умолчанию          | Описание                                                                           |
| ----------------------- | --------------------- | ---------------------------------------------------------------------------------- |
| `--demo`                | -                     | Путь к входному `.dem` файлу (обязательно)                                         |
| `--steamid`             | -                     | Целевой SteamID64 (обязательно, 17 цифр)                                           |
| `--out`                 | `highlights.json`     | Путь к выходному JSON (пустое значение отключает JSON)                             |
| `--hlae`                | `highlights.cfg`      | Путь к основному HLAE-скрипту                                                      |
| `--hlae-headshots`      | `headshots.cfg`       | Путь к HLAE-скрипту headshot-монтажа                                               |
| `--hlae-headshots-name` | `headshot_collection` | Имя выходной записи для headshot-монтажа                                           |
| `--hlae-path`           | `highlights`          | Префикс для `mirv_streams record name`                                             |
| `--hlae-preset`         | `afxFfmpegYuv420p`    | HLAE FFmpeg preset                                                                 |
| `--hlae-fps`            | `60`                  | FPS записи                                                                         |
| `--hlae-preroll`        | `3`                   | Секунды до события                                                                 |
| `--hlae-postroll`       | `2`                   | Секунды после события                                                              |
| `--hlae-kill-gap`       | `10`                  | Секунды между киллами в `round_multikill` для прыжка внутри записи (`0` отключает) |

Отключить генерацию headshot-монтажа:

```bash
go run ./cmd/highlighter ... --hlae-headshots ""
```

Отключить JSON-вывод:

```bash
go run ./cmd/highlighter ... --out ""
```

## Сценарии Записи

### Обычные хайлайты (`highlights.cfg`)

1. Запустите CS2 через HLAE.
2. Загрузите демо: `playdemo <demo_name>`.
3. Вставьте `highlights.cfg` в консоль HLAE.
4. Дождитесь `All N segments recorded`.
5. Скрипт завершится `disconnect` и вернет в главное меню.

Результат: несколько выходных файлов, по одному на сегмент.

### Headshot-монтаж (`headshots.cfg`)

1. Загрузите то же демо.
2. Вставьте `headshots.cfg`.
3. Запись стартует один раз, прыгает между headshot-сегментами и завершается один раз.
4. Скрипт завершится `disconnect` и вернет в главное меню.

Результат: один монтажный выходной файл.

## Примеры Сгенерированных Файлов

### `highlights.json`

```json
{
  "demo": "mirage.dem",
  "steamid": "7656119XXXXXXXXXX",
  "tick_rate": 64,
  "highlights": [
    {
      "type": "round_multikill",
      "round": 16,
      "tick_start": 112258,
      "tick_end": 112610,
      "kills": 3,
      "weapon": "M4A1",
      "player_slot": 10
    }
  ]
}
```

### `highlights.cfg`

```cfg
mirv_streams settings edit afxDefault settings afxFfmpegYuv420p;
mirv_streams record fps 60;
spec_show_xray 0;

mirv_cmd addAtTick 112066 "spec_player 10; host_framerate 60; mirv_streams record name highlights_hl_0005_r16_round_multikill; mirv_streams record start";
mirv_cmd addAtTick 112738 "mirv_streams record end; host_framerate 0";
mirv_cmd addAtTick 112739 "demo_pause; demo_gototick 118230; spec_player 10; demo_resume";
```

### `headshots.cfg`

```cfg
mirv_streams settings edit afxDefault settings afxFfmpegYuv420p;
mirv_streams record fps 60;

mirv_cmd addAtTick 26746 "spec_player 10; host_framerate 60; mirv_streams record name highlights_headshot_collection; mirv_streams record start";
mirv_cmd addAtTick 27067 "demo_pause; demo_gototick 59674; spec_player 10; demo_resume";
mirv_cmd addAtTick 118664 "mirv_streams record end; host_framerate 0";
```

## Валидация и Обработка Ошибок

- Fail-fast валидация конфигурации до старта парсинга:
  - пустой путь к демо
  - неверное расширение (не `.dem`)
  - отсутствующий / не обычный / пустой файл
  - некорректный формат SteamID64
  - лидирующие/хвостовые пробелы в строковых CLI-флагах автоматически обрезаются
- Защитное поведение парсера:
  - дополнительная проверка пути
  - конвертация parser panic в обычную ошибку
  - явная ошибка для битого/обрезанного демо
  - поддержка отмены через `context`

## Архитектура

- `cmd/highlighter`: CLI entrypoint
- `internal/bootstrap`: разбор конфигурации и запуск пайплайна
- `internal/parser`: извлечение событий из демо (`demoinfocs`)
- `internal/service`: правила хайлайтов и доменная логика
- `internal/hlae`: планирование сегментов и рендеринг скриптов
- `internal/repository`: слой сохранения данных

## Ограничения

- Качество вывода зависит от целостности демо и качества parser-событий.
- Определение clutch основано на правилах, а не на model/vision подходе.
- Headshot-монтаж это jump-cut автоматизация в демо-плеере, а не NLE постпродакшн.

## Roadmap

1. Новые типы хайлайтов (`awp_flick`, `360` и т.д.).
2. Выбор правил генерации хайлайтов.
3. Добавление звука в записанные видео с хайлайтами.

## License

Проект распространяется по лицензии MIT. См. [LICENSE](./LICENSE).
