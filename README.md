# Digital University Bot
**Чат-бот для образовательной экосистемы**  
_Решение для хакатона VK & Max_



## Основные возможности
Чат-бот в мессенджере **Max** предоставляет интегрированные инструменты для автоматизации учебного процесса:

###  Для студентов
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

### 1. Настройка окружения

Создайте файл `.env` в корне проекта на основе шаблона `.env.example`:

```ini
`LOG_LEVEL=0` - Уровень логирования

`LOG_DIR=../logs` - Директория для файлов логов

`DATABASE_URI=postgres://user:password@localhost:5432/dbname?sslmode=disable` - URI для связи с базой данных

`MAX_TOKEN`=your_bot_token_here - Токен бота в Max

`POSTGRES_USER=user` - Имя пользователя в PostgresDB

`POSTGRES_PASSWORD=password` - Пароль в PostgresDB

`POSTGRES_DB=dbname` - Название бд в PostgresDB
```

### 2. Способы запуска

#### Через Makefile (рекомендуется)
```bash
make up    # Запуск сервисов
make down  # Остановка
make reup  # Перезапуск
make clean # Очистка сборок
```

#### Через Docker-compose
```bash
# Запуск
docker-compose up --build -d

# Остановка
docker-compose down

# Полная очистка 
docker stop $(docker ps -a -q)
docker rm $(docker ps -a -q)
docker rmi $(docker images -q)
docker system prune -a --volumes -f
```



### 3. Инициализация данных
1. Загрузите начальные данные администратора .
В конце  файла  [`db/initdb.sql`](db/initdb.sql)  есть пример заполнения. 
2. Импортируйте учебные данные через админ-панель:
   - Список студентов ([students_example.csv](docs/Students_example.csv))
   - Расписание ([schedule_example.csv](docs/Schedule_example.csv))
   - Преподаватели ([teachers_example.csv](docs/Teachers_example.csv))

## Структура базы данных
![ER-диаграмма](screenshots/db.jpg) 

### Ключевые сущности:
| Таблица | Описание | Связи |
|---------|----------|-------|
| `users` | Пользователи системы | role_id → roles, group_id → groups |
| `schedule` | Расписание занятий | subject_id → subjects, group_id → groups |
| `grades` | Успеваемость | student_id → users, subject_id → subjects |
| `attendance` | Посещаемость | student_id → users, schedule_id → schedule |

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

###### Таблица материалов (materials)

```sql
CREATE TABLE IF NOT EXISTS materials (
    material_id SERIAL PRIMARY KEY,
    subject_id INT NOT NULL REFERENCES subjects(subject_id),
    file_url TEXT NOT NULL,
    uploaded_at TIMESTAMP DEFAULT NOW()
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


## Команда
| Роль | Имя | Контакт |
|------|-----|---------|
| Тимлид | Алибеков Аслан | [AlibekovAA](https://github.com/AlibekovAA) |
| Backend | Василов Иван |  [VasilovIS](https://github.com/Feof1l)|
| Аналитик | Ногеров Ислам | [NogerovIA](https://github.com/Samuel1-1) |

