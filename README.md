# Digital University Bot

**Чат-бот для образовательной экосистемы**
_Решение для хакатона VK & Max_

## Основные возможности

Чат-бот в мессенджере **Max** предоставляет интегрированные инструменты для автоматизации учебного процесса:

### Для студентов

- **Интерактивное расписание** с навигацией по дням недели
- **Успеваемость и посещаемость** в реальном времени
- **Мгновенные уведомления** о новых оценках и посещаемости
- **Аналитика успеваемости** по дисциплинам (средний балл и процент посещаемости)

### Для преподавателей

- **Интерактивное расписание** с навигацией по дням недели
- **Выставление оценок и посещаемости** прямо в чате

### Для администраторов

- **Массовая загрузка данных** через CSV-файлы
- **Управление ролями** (студенты/преподаватели)
- **Конфигурирование расписания** и учебных групп

## Архитектура системы

```mermaid
graph TD
    A[Max Messenger] --> B(Backend)
    B <--> |Данные| C[(PostgreSQL)]
    B --> D[CSV Importer]
    D -->|Данные| C
```

## Запуск проекта

Для локального запуска выполните следующие команды:

```bash
git clone https://github.com/Feof1l/DigitalUniversity.git

cd /Путь/DigitalUniversity
```

### 1. Настройка окружения

Создайте файл `.env` в корне проекта на основе шаблона `.env.example`. Необходимо указать свой токен, остальные переменные можно оставить без изменения:

```ini
`LOG_LEVEL=1` - Уровень логирования

`LOG_DIR=../logs` - Директория для файлов логов

`DATABASE_URI=postgres://user:password@postgres_db:5432/db?sslmode=disable` - URI для связи с базой данных

`MAX_TOKEN`=your_bot_token_here - Токен бота в Max

`POSTGRES_USER=user` - Имя пользователя в PostgresDB

`POSTGRES_PASSWORD=password` - Пароль в PostgresDB

`POSTGRES_DB=db` - Название бд в PostgresDB
```

### 2. Способы запуска

#### Через Docker-compose (рекомендуется)

```bash
# Запуск
docker-compose up --build

# Остановка
docker-compose down

# Полная очистка
docker stop $(docker ps -a -q)
docker rm $(docker ps -a -q)
docker rmi $(docker images -q)
docker system prune -a --volumes -f
```

Для запуска потребуется установка `make`

#### Через Makefile

```bash
make up    # Запуск сервисов
make down  # Остановка
make reup  # Перезапуск
make clean # Очистка сборок
```

### 3. Инициализация данных

1. Загрузите начальные данные администратора .
   В конце файла [`db/initdb.sql`](db/initdb.sql) есть пример заполнения. Для заполнения поля `usermax_id` используется `user_id` пользователя на платформе MAX, который можно получить, выполнив:

```
cd script/
go run main.go
```

Открыв тот же чат с ботом, Вам нужно нажать кнопку старт/отправить любое сообщение, и Вы получите свой id.

2. Импортируйте учебные данные через админ-панель:
   - Список студентов ([students_example.csv](docs/Students_example.csv))
   - Расписание ([schedule_example.csv](docs/Schedule_example.csv))
   - Преподаватели ([teachers_example.csv](docs/Teachers_example.csv))

## Структура базы данных

![ER-диаграмма](screenshots/db.jpg)

##### Структура базы данных

###### Таблица пользователей (users)

```sql
CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    usermax_id BIGINT UNIQUE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role_id INT REFERENCES roles(role_id),
    group_id INT REFERENCES groups(group_id)
);
```

###### Таблица ролей (roles)

```sql
CREATE TABLE IF NOT EXISTS roles (
    role_id SERIAL PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL
);
```

###### Таблица дисциплин (subjects)

```sql
CREATE TABLE IF NOT EXISTS subjects (
    subject_id SERIAL PRIMARY KEY,
    subject_name VARCHAR(255) NOT NULL,
    teacher_id INT NOT NULL REFERENCES users(user_id)
);
```

###### Таблица групп (groups)

```sql
CREATE TABLE IF NOT EXISTS groups (
    group_id SERIAL PRIMARY KEY,
    group_name VARCHAR(100) UNIQUE NOT NULL
);
```

###### Таблица связи групп и дисциплин (groups_subjects)

```sql
CREATE TABLE IF NOT EXISTS groups_subjects (
    group_id INT NOT NULL REFERENCES groups(group_id),
    subject_id INT NOT NULL REFERENCES subjects(subject_id),
    PRIMARY KEY (group_id, subject_id)
);
```

###### Таблица типов занятий (lesson_types)

```sql
CREATE TABLE IF NOT EXISTS lesson_types (
    lesson_type_id SERIAL PRIMARY KEY,
    type_name VARCHAR(50) UNIQUE NOT NULL
);

```

###### Таблица расписания (schedule)

```sql
CREATE TABLE IF NOT EXISTS schedule (
    schedule_id SERIAL PRIMARY KEY,
    weekday SMALLINT NOT NULL CHECK (
        weekday BETWEEN 1 AND 7
    ),
    class_room VARCHAR(100),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    subject_id INT NOT NULL REFERENCES subjects(subject_id),
    teacher_id INT NOT NULL REFERENCES users(user_id),
    group_id INT NOT NULL REFERENCES groups(group_id),
    lesson_type_id INT NOT NULL REFERENCES lesson_types(lesson_type_id)
);
```

###### Таблица оценок (grades)

```sql
CREATE TABLE IF NOT EXISTS grades (
    grade_id SERIAL PRIMARY KEY,
    student_id INT NOT NULL REFERENCES users(user_id),
    teacher_id INT NOT NULL REFERENCES users(user_id),
    subject_id INT NOT NULL REFERENCES subjects(subject_id),
    schedule_id INT NOT NULL REFERENCES schedule(schedule_id),
    grade_value INT NOT NULL CHECK (
        grade_value BETWEEN 0 AND 5
    ),
    grade_date TIMESTAMP DEFAULT NOW()
);
```

###### Таблица посещаемости (attendance)

```sql
CREATE TABLE IF NOT EXISTS attendance (
    attendance_id SERIAL PRIMARY KEY,
    student_id INT NOT NULL REFERENCES users(user_id),
    schedule_id INT NOT NULL REFERENCES schedule(schedule_id),
    attended BOOLEAN NOT NULL,
    mark_time TIMESTAMP DEFAULT NOW()
);
```

Полная схема: [`db/initdb.sql`](db/initdb.sql)

## Структура проекта

```
src/
├── application              # Иициализация приложения
│   └── application.go
├── backend.yml
├── config                   # Конфигурация
│   └── config.go
├── database                 # Работа с БД
│   ├── database.go
│   ├── models.go            # ORM-модели
│   └── repository.go        # Методы доступа к данным
├── Dockerfile
├── go.mod
├── go.sum
├── logger
│   └── logger.go            # Кастомный логгер
├── main.go                  # Точка входа
├── maxAPI                   # Интеграция с Max
│   ├── attendance.go        # Работа с посещаемостью
│   ├── bot.go               # Инициализация бота
│   ├── handlers.go          # Обработчики команд
│   ├── keyboard.go          # Генерация клавиатур
│   ├── message.go           # Работа с расписанием
│   ├── schedule.go          # Работа с посещаемостью
│   ├── student_grades.go    # Работа с отправкой сообщений
│   ├── teacher_grades.go    # Работа с оценками для студента
│   └── utils.go             # Работа с оценками для преподавателя
├── services                 # Вспомогательные методы
│   ├── importer.go          # Импорт данных
│   └── validator.go         # Валидация входных данных
```

## Используемые библиотеки

Работа с конфигом и окржунием

- github.com/caarlos0/env/v11 v11.3.1
- github.com/joho/godotenv v1.5.1

Max Bot API Client library for Golang

- maxbot "github.com/max-messenger/max-bot-api-client-go" v1.0.3

Работа с бд

- github.com/jmoiron/sqlx v1.4.0
- github.com/lib/pq v1.10.9

Кастомный логгер

- gopkg.in/natefinch/lumberjack. v2 v2.2.1

## Дальнейшее развитие

Проект является расширяемым и масштабируемым, благодаря чему можно без проблем вносить новые функциональности.
Наиболее перспективными и полезными являются:

- Внедрение сервиса электронного документооборота: загружается шаблон справки или заявления;
  пользователь запрашивает получение данного документа; бот выдает пользователю файл, подставляя в пустые поля необходимые данные (ФИО, группа, текущая дата)
- Прикрепление документов к дисциплинам. Для преподавателей появится возможность загружать документы (задания лабораторных работ, методички, лекции) и прикреплять к своим дисциплинам.Студенты смогут посмотреть список доступных файлов и скачать нужный при необходимости.

## Команда

| Роль     | Имя            | Контакт                                |
| -------- | -------------- | -------------------------------------- |
| Техлид   | Алибеков Аслан | [AlibekovAA](https://t.me/alibekov_05) |
| Backend  | Василов Иван   | [VasilovIS](https://t.me/feof1l)       |
| Аналитик | Ногеров Ислам  | [NogerovIA](https://t.me/mrsigint)     |
